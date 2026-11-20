package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.WriteFile("frontmatter-validator.yaml", []byte("invalid: ["), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		rootCmd.SetArgs(nil)
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		if flag := rootCmd.Flags().Lookup("version"); flag != nil {
			if err := flag.Value.Set("false"); err != nil {
				t.Error(err)
			}
		}
	})
	for _, argument := range []string{"version", "--version"} {
		t.Run(argument, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			rootCmd.SetArgs([]string{argument})
			rootCmd.SetOut(&stdout)
			rootCmd.SetErr(&stderr)
			if err := rootCmd.ExecuteContext(t.Context()); err != nil {
				t.Fatalf("execute: %v", err)
			}
			expected := "frontmatter-validator version " + rootCmd.Version + "\n"
			if stdout.String() != expected {
				t.Fatalf("stdout = %q, want %q", stdout.String(), expected)
			}
			if stderr.Len() != 0 {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}
			for _, field := range []string{"commit:", "built:"} {
				if !strings.Contains(stdout.String(), field) {
					t.Errorf("stdout %q does not contain %q", stdout.String(), field)
				}
			}
		})
	}
}
