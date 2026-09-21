package route

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAdversarialRoute(t *testing.T) {
	t.Run("empty-directory", func(t *testing.T) {
		root := workflow(t, "")
		if err := os.Remove(filepath.Join(root, ".github/workflows/ci.yml")); err != nil {
			t.Fatal(err)
		}
		got, err := Selected(root)
		t.Logf("selected=%v error=%v", got, err)
		if err != nil || len(got) != 0 {
			t.Fatalf("got %v, %v", got, err)
		}
	})
	t.Run("same-basename", func(t *testing.T) {
		root := workflow(t, "run: brew install rulefloor\n")
		if err := os.WriteFile(filepath.Join(root, ".github/workflows/ci.yaml"), []byte("run: brew install lictor\n"), 0600); err != nil {
			t.Fatal(err)
		}
		got, err := Selected(root)
		t.Logf("selected=%v error=%v", got, err)
		if err != nil || !got["rulefloor"] || !got["lictor"] {
			t.Fatalf("lost one workflow: %v, %v", got, err)
		}
	})
	for _, kind := range []string{"file", "directory"} {
		t.Run("outside-symlink-"+kind, func(t *testing.T) {
			root := t.TempDir()
			dir := filepath.Join(root, ".github/workflows")
			outside := t.TempDir()
			if err := os.WriteFile(filepath.Join(outside, "ci.yml"), []byte("run: brew install lictor\n"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Dir(dir), 0700); err != nil {
				t.Fatal(err)
			}
			if kind == "directory" {
				if err := os.Symlink(outside, dir); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.Mkdir(dir, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(filepath.Join(outside, "ci.yml"), filepath.Join(dir, "ci.yml")); err != nil {
					t.Fatal(err)
				}
			}
			got, err := Selected(root)
			t.Logf("selected=%v error=%v", got, err)
			if err == nil {
				t.Fatalf("workflow outside selected tree was read: %v", got)
			}
		})
	}
}
