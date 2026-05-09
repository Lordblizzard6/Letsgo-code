package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestHelpTextUsesRootCommandName(t *testing.T) {
	files := []string{
		"tasks.go",
		"plugin.go",
		"../README.md",
		"../docs/PROVIDER_TOOLS_GUIDE.md",
		"../docs/KAIROS_REFERENCE.md",
	}

	legacyNames := []string{"claudego ", "letsGo "}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("failed to read %s: %v", file, err)
		}

		text := string(content)
		for _, legacy := range legacyNames {
			if strings.Contains(text, legacy) {
				t.Errorf("%s contains legacy CLI name %q", file, strings.TrimSpace(legacy))
			}
		}
	}
}
