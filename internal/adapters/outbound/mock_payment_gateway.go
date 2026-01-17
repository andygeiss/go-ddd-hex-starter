package outbound

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"github.com/andygeiss/hotel-booking/internal/domain/payment"
	"github.com/andygeiss/hotel-booking/internal/domain/shared"
)

// MockPaymentGateway simulates a payment gateway for testing and demonstration.
type MockPaymentGateway struct {
	transactions map[string]shared.Money
	FailureRate  float64 // 0.0 to 1.0, probability of random failures
	ShouldFail   bool
}

// NewMockPaymentGateway creates a new mock payment gateway.
func NewMockPaymentGateway() *MockPaymentGateway {
	return &MockPaymentGateway{
		ShouldFail:   false,
		FailureRate:  0.0,
		transactions: make(map[string]shared.Money),
	}
}

// cryptoRandFloat64 returns a random float64 in [0.0, 1.0) using crypto/rand.
func cryptoRandFloat64() float64 {
	maxVal := big.NewInt(1 << 53)
	n, err := rand.Int(rand.Reader, maxVal)
	if err != nil {
		return 0
	}
	return float64(n.Int64()) / float64(1<<53)
}

// Authorize simulates authorizing a payment.
func (g *MockPaymentGateway) Authorize(ctx context.Context, pay *payment.Payment) (string, error) {
	if g.ShouldFail || (g.FailureRate > 0 && cryptoRandFloat64() < g.FailureRate) {
		return "", errors.New("payment authorization failed: insufficient funds")
	}

	transactionID := fmt.Sprintf("txn_%s_%d", pay.ID, pay.Amount.Amount)
	g.transactions[transactionID] = pay.Amount

	return transactionID, nil
}

// Capture simulates capturing an authorized payment.
func (g *MockPaymentGateway) Capture(ctx context.Context, transactionID string, amount shared.Money) error {
	if g.ShouldFail || (g.FailureRate > 0 && cryptoRandFloat64() < g.FailureRate) {
		return errors.New("payment capture failed: gateway timeout")
	}

	authorizedAmount, exists := g.transactions[transactionID]
	if !exists {
		return fmt.Errorf("transaction %s not found", transactionID)
	}

	if authorizedAmount.Amount != amount.Amount || authorizedAmount.Currency != amount.Currency {
		return fmt.Errorf("capture amount mismatch: authorized %v, requested %v", authorizedAmount, amount)
	}

	return nil
}

// Refund simulates refunding a captured payment.
func (g *MockPaymentGateway) Refund(ctx context.Context, transactionID string, amount shared.Money) error {
	if g.ShouldFail || (g.FailureRate > 0 && cryptoRandFloat64() < g.FailureRate) {
		return errors.New("payment refund failed: gateway error")
	}

	_, exists := g.transactions[transactionID]
	if !exists {
		return fmt.Errorf("transaction %s not found", transactionID)
	}

	delete(g.transactions, transactionID)

	return nil
}

// SetShouldFail configures the mock to always fail (for testing error paths).
func (g *MockPaymentGateway) SetShouldFail(shouldFail bool) {
	g.ShouldFail = shouldFail
}

// SetFailureRate sets the probability of random failures (0.0 to 1.0).
func (g *MockPaymentGateway) SetFailureRate(rate float64) {
	g.FailureRate = rate
}

// Reset clears all transaction state.
func (g *MockPaymentGateway) Reset() {
	g.transactions = make(map[string]shared.Money)
	g.ShouldFail = false
	g.FailureRate = 0.0
}
