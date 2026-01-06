package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/andygeiss/cloud-native-utils/messaging"
	"github.com/andygeiss/go-ddd-hex-starter/internal/adapters/inbound"
	"github.com/andygeiss/go-ddd-hex-starter/internal/adapters/outbound"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/agent"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/event"
	"github.com/andygeiss/go-ddd-hex-starter/internal/domain/indexing"
)

// cliToolExecutor is a simple tool executor for CLI demonstration.
type cliToolExecutor struct {
	tools map[string]func(ctx context.Context, args string) (string, error)
}

// newCLIToolExecutor creates a new CLI tool executor with sample tools.
func newCLIToolExecutor() *cliToolExecutor {
	executor := &cliToolExecutor{
		tools: make(map[string]func(ctx context.Context, args string) (string, error)),
	}
	// Register a sample "echo" tool for demonstration
	executor.tools["echo"] = echoTool
	return executor
}

// echoTool is a sample tool that echoes the input.
func echoTool(_ context.Context, args string) (string, error) {
	if args == "" {
		return "", errors.New("echo requires non-empty input")
	}
	return "Echo: " + args, nil
}

// Execute runs the specified tool with the given input arguments.
func (e *cliToolExecutor) Execute(ctx context.Context, toolName string, arguments string) (string, error) {
	if tool, ok := e.tools[toolName]; ok {
		return tool(ctx, arguments)
	}
	return "", fmt.Errorf("tool not found: %s", toolName)
}

// GetAvailableTools returns the list of available tool names.
func (e *cliToolExecutor) GetAvailableTools() []string {
	tools := make([]string, 0, len(e.tools))
	for name := range e.tools {
		tools = append(tools, name)
	}
	return tools
}

// HasTool returns true if the specified tool is available.
func (e *cliToolExecutor) HasTool(toolName string) bool {
	_, ok := e.tools[toolName]
	return ok
}

// getEnvOrDefault returns the environment variable value or default if not set.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// runAgentAnalysis runs the agent to analyze the created index.
func runAgentAnalysis(ctx context.Context, taskService *agent.TaskService, evt *indexing.EventFileIndexCreated) {
	fmt.Printf("❯ agent: starting agent loop for index analysis...\n")

	agentInstance := agent.NewAgent(
		agent.AgentID("agent-"+string(evt.IndexID)),
		"You are a helpful assistant that analyzes file indexes. When given information about indexed files, provide a summary of the project structure.",
	)
	agentInstance.WithMaxIterations(5)

	task := agent.NewTask(
		agent.TaskID("task-analyze-"+string(evt.IndexID)),
		"analyze-index",
		fmt.Sprintf("Analyze the file index with ID '%s' containing %d files. Provide a brief summary.", evt.IndexID, evt.FileCount),
	)

	result, runErr := taskService.RunTask(ctx, &agentInstance, task)
	if runErr != nil {
		fmt.Printf("❯ agent: error running task - %v\n", runErr)
		return
	}

	if result.Success {
		fmt.Printf("❯ agent: task completed successfully\n")
		fmt.Printf("❯ agent: output - %s\n", result.Output)
	} else {
		fmt.Printf("❯ agent: task failed - %s\n", result.Error)
	}
}

// createAgentEventHandler creates the event handler for index created events.
func createAgentEventHandler(
	ctx context.Context,
	taskService *agent.TaskService,
	wg *sync.WaitGroup,
	once *sync.Once,
) event.EventHandlerFn {
	return func(e event.Event) error {
		// Ensure we only process once (in case of duplicate events)
		once.Do(func() {
			defer wg.Done()

			evt := e.(*indexing.EventFileIndexCreated)
			fmt.Printf("❯ event: received EventFileIndexCreated - IndexID: %s, FileCount: %d\n", evt.IndexID, evt.FileCount)

			runAgentAnalysis(ctx, taskService, evt)
		})
		return nil
	}
}

// printIndexSummary prints a summary of the index.
func printIndexSummary(index *indexing.Index) {
	fmt.Printf("❯ main: index created at %s with %d files\n", index.CreatedAt.Format(time.RFC3339), len(index.FileInfos))
	fmt.Printf("❯ main: index hash: %s\n", index.Hash())

	// Demonstrate listing some file infos from the index.
	fmt.Printf("❯ main: first 5 files in index:\n")
	for i, fi := range index.FileInfos {
		if i >= 5 {
			fmt.Printf("  ... and %d more files\n", len(index.FileInfos)-5)
			break
		}
		fmt.Printf("  - %s (%d bytes)\n", fi.AbsPath, fi.Size)
	}
}

// setupAgentComponents creates and returns the agent service components.
func setupAgentComponents(eventPublisher event.EventPublisher) *agent.TaskService {
	lmStudioURL := getEnvOrDefault("LM_STUDIO_URL", "http://localhost:1234")
	lmStudioModel := getEnvOrDefault("LM_STUDIO_MODEL", "default")
	llmClient := outbound.NewLMStudioClient(lmStudioURL, lmStudioModel)
	toolExecutor := newCLIToolExecutor()
	return agent.NewTaskService(llmClient, toolExecutor, eventPublisher)
}

// waitForAgentCompletion waits for the agent to complete or timeout.
func waitForAgentCompletion(ctx context.Context, wg *sync.WaitGroup) {
	fmt.Printf("❯ main: waiting for agent to complete...\n")
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		fmt.Printf("❯ main: agent completed\n")
	case <-ctx.Done():
		fmt.Printf("❯ main: agent timed out\n")
	}
}

func main() {
	// Create a context with timeout for the agent task
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Setup the messaging dispatcher (external Kafka pub/sub).
	dispatcher := messaging.NewExternalDispatcher()

	// Setup the inbound adapters.
	fileReader := inbound.NewFileReader()
	eventSubscriber := inbound.NewEventSubscriber(dispatcher)

	// Setup the outbound adapters.
	indexPath := "./index.json"
	defer func() { _ = os.Remove(indexPath) }()
	indexRepository := outbound.NewFileIndexRepository(indexPath)
	eventPublisher := outbound.NewEventPublisher(dispatcher)

	// Setup the agent components.
	taskService := setupAgentComponents(eventPublisher)

	// WaitGroup and Once to coordinate agent completion and prevent duplicate processing
	var wg sync.WaitGroup
	var once sync.Once
	wg.Add(1)

	// Subscribe to the EventFileIndexCreated event to start the agent loop.
	err := eventSubscriber.Subscribe(
		ctx,
		indexing.EventTopicFileIndexCreated,
		func() event.Event { return indexing.NewEventFileIndexCreated() },
		createAgentEventHandler(ctx, taskService, &wg, &once),
	)
	if err != nil {
		panic(err)
	}

	// Create the indexing service with all dependencies injected.
	indexingService := indexing.NewIndexingService(fileReader, indexRepository, eventPublisher)

	// Use the service to create an index (this will also publish the event).
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	fmt.Printf("❯ main: creating index for path: %s\n", wd)

	createErr := indexingService.CreateIndex(ctx, wd)
	if createErr != nil {
		panic(createErr)
	}

	// Wait for agent task to complete
	waitForAgentCompletion(ctx, &wg)

	// Read the index back from the repository to demonstrate the full cycle.
	id := indexing.IndexID(wd)
	index, err := indexRepository.Read(context.Background(), id)
	if err != nil {
		panic(err)
	}

	printIndexSummary(index)
}
