package tools

import (
	"path/filepath"
	"testing"
)

func TestGosecTool(t *testing.T) {
	tool := NewGosecTool()

	testDir := filepath.Join("..", "..", "test", "go_io1")

	t.Run("Run on go_io1", func(t *testing.T) {
		output, err := tool.Run(testDir)
		if err != nil {
			t.Fatalf("Failed to run gosec: %v", err)
		}

		t.Logf("Critical: %d", len(output.Critical))
		t.Logf("Warnings: %d", len(output.Warnings))
		t.Logf("Info: %d", len(output.Info))

		for _, finding := range output.Critical {
			t.Logf("CRITICAL: %s at %s", finding.Message, finding.Location)
		}
	})
}

func TestGosecToolOnAllTestDirs(t *testing.T) {
	tool := NewGosecTool()

	testCases := []string{
		filepath.Join("..", "..", "test", "go_basic1"),
		filepath.Join("..", "..", "test", "go_io1"),
	}

	for _, testDir := range testCases {
		t.Run(testDir, func(t *testing.T) {
			output, err := tool.Run(testDir)
			if err != nil {
				t.Logf("Error running on %s: %v", testDir, err)
				return
			}

			t.Logf("Results for %s:", testDir)
			t.Logf("  Critical: %d", len(output.Critical))
			t.Logf("  Warnings: %d", len(output.Warnings))
			t.Logf("  Info: %d", len(output.Info))
		})
	}
}
