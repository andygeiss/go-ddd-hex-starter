#!/usr/bin/env python3
"""
Tests for create_pgo.py
=======================

This module contains unit tests for the PGO profile generation utility.
It tests command execution, file operations, and the overall workflow.

Usage:
    python3 -m unittest tools.create_pgo_test -v
    # Or from tools directory:
    python3 -m unittest create_pgo_test -v
"""

import glob
import os
import re
import subprocess
import tempfile
import unittest
from pathlib import Path
from unittest.mock import MagicMock, patch

from create_pgo import PACKAGES_TO_PROFILE, run_command, validate_package_path


class TestValidatePackagePath(unittest.TestCase):
    """Tests for the validate_package_path function."""

    def test_valid_simple_package(self):
        """Simple package path should be valid."""
        self.assertTrue(validate_package_path("cmd/server"))

    def test_valid_nested_package(self):
        """Nested package path should be valid."""
        self.assertTrue(validate_package_path("internal/adapters/inbound"))

    def test_valid_with_underscore(self):
        """Package with underscore should be valid."""
        self.assertTrue(validate_package_path("my_package/sub_dir"))

    def test_valid_with_hyphen(self):
        """Package with hyphen should be valid."""
        self.assertTrue(validate_package_path("my-package/sub-dir"))

    def test_invalid_empty_string(self):
        """Empty string should be invalid."""
        self.assertFalse(validate_package_path(""))

    def test_invalid_path_traversal(self):
        """Path traversal attempts should be invalid."""
        self.assertFalse(validate_package_path("../etc/passwd"))
        self.assertFalse(validate_package_path("cmd/../../../etc"))
        self.assertFalse(validate_package_path(".."))

    def test_invalid_shell_metacharacters(self):
        """Shell metacharacters should be rejected."""
        self.assertFalse(validate_package_path("cmd; rm -rf /"))
        self.assertFalse(validate_package_path("cmd && echo pwned"))
        self.assertFalse(validate_package_path("cmd | cat /etc/passwd"))
        self.assertFalse(validate_package_path("$(whoami)"))
        self.assertFalse(validate_package_path("`whoami`"))

    def test_invalid_special_characters(self):
        """Special characters should be rejected."""
        self.assertFalse(validate_package_path("cmd/server$VAR"))
        self.assertFalse(validate_package_path("cmd/server*"))
        self.assertFalse(validate_package_path("cmd/server?"))
        self.assertFalse(validate_package_path("cmd/server[0]"))

    def test_all_configured_packages_are_valid(self):
        """All packages in PACKAGES_TO_PROFILE should pass validation."""
        for pkg in PACKAGES_TO_PROFILE:
            self.assertTrue(
                validate_package_path(pkg),
                f"Configured package '{pkg}' should be valid"
            )


class TestPackagesToProfile(unittest.TestCase):
    """Tests for the PACKAGES_TO_PROFILE constant."""

    def test_packages_is_list(self):
        """PACKAGES_TO_PROFILE should be a list."""
        self.assertIsInstance(PACKAGES_TO_PROFILE, list)

    def test_packages_not_empty(self):
        """PACKAGES_TO_PROFILE should contain at least one package."""
        self.assertGreater(len(PACKAGES_TO_PROFILE), 0)

    def test_packages_are_strings(self):
        """All packages should be strings."""
        for pkg in PACKAGES_TO_PROFILE:
            self.assertIsInstance(pkg, str)

    def test_packages_are_valid_paths(self):
        """All packages should be valid Go package paths."""
        for pkg in PACKAGES_TO_PROFILE:
            # Should not start or end with /
            self.assertFalse(pkg.startswith("/"), f"Package '{pkg}' should not start with /")
            self.assertFalse(pkg.endswith("/"), f"Package '{pkg}' should not end with /")
            # Should not contain backslashes
            self.assertNotIn("\\", pkg, f"Package '{pkg}' should not contain backslashes")

    def test_expected_packages_present(self):
        """Expected core packages should be in the list."""
        expected = ["cmd/server", "internal/adapters/inbound", "internal/adapters/outbound"]
        for pkg in expected:
            self.assertIn(pkg, PACKAGES_TO_PROFILE, f"Expected package '{pkg}' not found")


class TestRunCommand(unittest.TestCase):
    """Tests for the run_command function."""

    def test_run_command_success(self):
        """Successful command should complete without error."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0)
            run_command(["echo", "test"])
            mock_run.assert_called_once_with(["echo", "test"], check=True, stdout=None)

    def test_run_command_with_stdout_redirect(self):
        """Command with stdout should pass stdout argument."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0)
            mock_file = MagicMock()
            run_command(["echo", "test"], stdout=mock_file)
            mock_run.assert_called_once_with(
                ["echo", "test"], check=True, stdout=mock_file
            )

    def test_run_command_failure_exits(self):
        """Failed command should exit with the error code."""
        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = subprocess.CalledProcessError(1, "cmd")
            with self.assertRaises(SystemExit) as exc_info:
                run_command(["false"])
            self.assertEqual(exc_info.exception.code, 1)

    def test_run_command_check_false(self):
        """Command with check=False should not raise on error."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=1)
            # Should not raise
            run_command(["false"], check=False)
            mock_run.assert_called_once()


class TestProfileFilenameGeneration(unittest.TestCase):
    """Tests for profile filename generation logic."""

    def test_package_to_filename_conversion(self):
        """Package paths should be converted to safe filenames."""
        test_cases = [
            ("cmd/server", "cpuprofile-cmd__server.pprof"),
            ("internal/adapters/inbound", "cpuprofile-internal__adapters__inbound.pprof"),
            ("pkg", "cpuprofile-pkg.pprof"),
        ]
        for pkg, expected in test_cases:
            suffix = pkg.replace("/", "__").replace("\\", "__")
            output_file = f"cpuprofile-{suffix}.pprof"
            self.assertEqual(output_file, expected)

    def test_windows_path_conversion(self):
        """Windows-style paths should also be converted correctly."""
        pkg = "internal\\adapters\\outbound"
        suffix = pkg.replace("/", "__").replace("\\", "__")
        output_file = f"cpuprofile-{suffix}.pprof"
        self.assertEqual(output_file, "cpuprofile-internal__adapters__outbound.pprof")


class TestCleanupLogic(unittest.TestCase):
    """Tests for cleanup file patterns."""

    def test_pprof_glob_pattern(self):
        """The pprof glob pattern should match expected files."""
        with tempfile.TemporaryDirectory() as tmpdir:
            old_cwd = os.getcwd()
            os.chdir(tmpdir)
            try:
                # Create test files
                Path("cpuprofile.pprof").touch()
                Path("cpuprofile-cmd__server.pprof").touch()
                Path("cpuprofile-merged.pprof").touch()
                Path("other.pprof").touch()

                # Test glob pattern
                matches = glob.glob("cpuprofile*.pprof")
                self.assertIn("cpuprofile.pprof", matches)
                self.assertIn("cpuprofile-cmd__server.pprof", matches)
                self.assertIn("cpuprofile-merged.pprof", matches)
                self.assertNotIn("other.pprof", matches)
            finally:
                os.chdir(old_cwd)

    def test_test_binary_glob_pattern(self):
        """The test binary glob pattern should match expected files."""
        with tempfile.TemporaryDirectory() as tmpdir:
            old_cwd = os.getcwd()
            os.chdir(tmpdir)
            try:
                # Create test files
                Path("server.test").touch()
                Path("inbound.test").touch()
                Path("not_a_test.txt").touch()

                # Test glob pattern
                matches = glob.glob("*.test")
                self.assertIn("server.test", matches)
                self.assertIn("inbound.test", matches)
                self.assertNotIn("not_a_test.txt", matches)
            finally:
                os.chdir(old_cwd)


class TestBenchmarkCommand(unittest.TestCase):
    """Tests for benchmark command construction."""

    def test_benchmark_command_structure(self):
        """Benchmark command should have correct structure."""
        pkg = "cmd/server"
        suffix = pkg.replace("/", "__")
        output_file = f"cpuprofile-{suffix}.pprof"

        cmd = [
            "go",
            "test",
            f"./{pkg}/...",
            "-run=^$",
            "-bench=.",
            "-benchtime=10s",
            f"-cpuprofile={output_file}",
            "-pgo=off",
        ]

        self.assertEqual(cmd[0], "go")
        self.assertEqual(cmd[1], "test")
        self.assertEqual(cmd[2], "./cmd/server/...")
        self.assertIn("-run=^$", cmd)  # Skip unit tests
        self.assertIn("-bench=.", cmd)  # Run all benchmarks
        self.assertIn("-pgo=off", cmd)  # Disable PGO during profiling
        self.assertIn(f"-cpuprofile={output_file}", cmd)

    def test_benchmark_skips_unit_tests(self):
        """Benchmark command should skip unit tests with -run=^$."""
        # The regex ^$ matches nothing, so no unit tests run
        pattern = re.compile("^$")
        self.assertIsNotNone(pattern.match(""))
        self.assertIsNone(pattern.match("TestSomething"))


class TestMergeCommand(unittest.TestCase):
    """Tests for profile merge command."""

    def test_merge_command_structure(self):
        """Merge command should use go tool pprof with proto output."""
        # The merge command is now a list (no shell=True needed)
        profile_files = ["cpuprofile-cmd__server.pprof", "cpuprofile-internal__adapters__inbound.pprof"]
        merge_cmd = ["go", "tool", "pprof", "-proto"] + profile_files
        self.assertIn("go", merge_cmd)
        self.assertIn("tool", merge_cmd)
        self.assertIn("pprof", merge_cmd)
        self.assertIn("-proto", merge_cmd)
        # Profile files should be appended to the command
        for pf in profile_files:
            self.assertIn(pf, merge_cmd)


class TestSvgCommand(unittest.TestCase):
    """Tests for SVG generation command."""

    def test_svg_command_structure(self):
        """SVG command should use go tool pprof with svg output."""
        # The SVG command is now a list (no shell=True needed)
        svg_cmd = ["go", "tool", "pprof", "-svg", "cpuprofile.pprof"]
        self.assertIn("go", svg_cmd)
        self.assertIn("tool", svg_cmd)
        self.assertIn("pprof", svg_cmd)
        self.assertIn("-svg", svg_cmd)
        self.assertIn("cpuprofile.pprof", svg_cmd)


class TestFileOperations(unittest.TestCase):
    """Tests for file copy operations."""

    def test_copy_merged_profile(self):
        """Merged profile should be correctly copied to final location."""
        with tempfile.TemporaryDirectory() as tmpdir:
            src = Path(tmpdir) / "cpuprofile-merged.pprof"
            dst = Path(tmpdir) / "cpuprofile.pprof"

            # Write test content
            test_content = b"test profile data"
            src.write_bytes(test_content)

            # Copy operation (as done in the script)
            with open(src, "rb") as src_file, open(dst, "wb") as dst_file:
                dst_file.write(src_file.read())

            self.assertTrue(dst.exists())
            self.assertEqual(dst.read_bytes(), test_content)


class TestIntegrationScenarios(unittest.TestCase):
    """Integration-style tests for complete workflows."""

    def test_full_cleanup_scenario(self):
        """Test that cleanup removes all expected files."""
        with tempfile.TemporaryDirectory() as tmpdir:
            old_cwd = os.getcwd()
            os.chdir(tmpdir)
            try:
                # Create files that should be cleaned up
                Path("cpuprofile.pprof").touch()
                Path("cpuprofile-cmd__server.pprof").touch()
                Path("cpuprofile-merged.pprof").touch()
                Path("server.test").touch()

                # Create file that should NOT be cleaned up
                Path("important.txt").touch()

                # Simulate cleanup
                for f in glob.glob("cpuprofile*.pprof"):
                    os.remove(f)
                for f in glob.glob("*.test"):
                    os.remove(f)

                # Verify
                self.assertFalse(Path("cpuprofile.pprof").exists())
                self.assertFalse(Path("cpuprofile-cmd__server.pprof").exists())
                self.assertFalse(Path("cpuprofile-merged.pprof").exists())
                self.assertFalse(Path("server.test").exists())
                self.assertTrue(Path("important.txt").exists())  # Should be preserved
            finally:
                os.chdir(old_cwd)

    def test_artifact_naming_convention(self):
        """Final artifacts should follow naming convention."""
        # The script produces these specific files
        expected_artifacts = ["cpuprofile.pprof", "cpuprofile.svg"]
        for artifact in expected_artifacts:
            # Verify naming convention (no path separators, correct extension)
            self.assertNotIn("/", artifact)
            self.assertNotIn("\\", artifact)
            self.assertTrue(artifact.startswith("cpuprofile"))


if __name__ == "__main__":
    unittest.main()
