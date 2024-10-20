package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/gkwa/creepymatrix/internal/logger"
)

func TestCustomLogger(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := rootCmd
	cmd.SetArgs([]string{"version"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	if err != nil {
		t.Fatalf("Failed to copy output: %v", err)
	}
	output := buf.String()

	if !strings.Contains(output, "Version:") {
		t.Errorf("Expected output to contain version information, but got: %s", output)
	}

	t.Logf("Command output: %s", output)
}

func TestJSONLogger(t *testing.T) {
	oldVerbose, oldLogFormat := verbose, logFormat
	verbose, logFormat = 1, "json"
	defer func() {
		verbose, logFormat = oldVerbose, oldLogFormat
	}()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	customLogger := logger.NewConsoleLogger(verbose, logFormat == "json")
	cliLogger = customLogger

	cmd := rootCmd
	cmd.SetArgs([]string{"version"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	if err != nil {
		t.Fatalf("Failed to copy output: %v", err)
	}
	output := buf.String()

	if !strings.Contains(output, "Version:") {
		t.Errorf("Expected output to contain version information, but got: %s", output)
	}

	t.Logf("Command output: %s", output)
}
