package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyPatchTool(t *testing.T) {
	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "sample.txt")
	initialContent := "line 1\nline 2\nline 3\nline 4\n"
	if err := os.WriteFile(targetFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("failed to write initial file: %v", err)
	}

	patch := `--- a/sample.txt
+++ b/sample.txt
@@ -1,4 +1,5 @@
 line 1
-line 2
+line 2 updated
+line 2.5 inserted
 line 3
 line 4
`
	tool := &ApplyPatchTool{}
	msg, err := tool.Execute(map[string]interface{}{
		"patch": patch,
		"path":  targetFile,
	})
	if err != nil {
		t.Fatalf("apply_patch failed: %v", err)
	}
	t.Logf("apply_patch output: %s", msg)

	resultBytes, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("failed to read target file: %v", err)
	}

	expected := "line 1\nline 2 updated\nline 2.5 inserted\nline 3\nline 4\n"
	if string(resultBytes) != expected {
		t.Fatalf("expected:\n%q\ngot:\n%q", expected, string(resultBytes))
	}
}
