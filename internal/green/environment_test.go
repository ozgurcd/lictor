package green

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ozgurcd/lictor/internal/executor"
)

func TestAmbientGoFlagsCannotHideRedFile(t *testing.T) {
	offline(t)
	t.Setenv("GOFLAGS", "-tags=nonexistent")
	for _, red := range []bool{false, true} {
		root := t.TempDir()
		files := map[string]string{"go.mod": "module fixture\n\ngo 1.21\n", "base.go": "package fixture\n\nfunc Value() int { return 1 }\n"}
		if red {
			files["red.go"] = "//go:build !nonexistent\n\npackage fixture\n\nvar _ = undefinedSymbol\n"
		}
		for name, body := range files {
			if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0600); err != nil {
				t.Fatal(err)
			}
		}
		got := Run(t.Context(), root, executor.GoTools{})
		want := 0
		if red {
			want = 1
		}
		if got.ExitCode != want {
			t.Fatalf("red=%v GOFLAGS=-tags=nonexistent: want %d got %+v", red, want, got)
		}
		t.Logf("red=%v GOFLAGS=-tags=nonexistent: %s", red, got.Line())
	}
}

func TestHumanLineOmitsPlatform(t *testing.T) {
	r := Result{Outcome: "GREEN", Cause: "builds, vets and tests clean", Subject: "fixture", GoVersion: "go version go1.27.1 darwin/arm64"}
	if !strings.HasSuffix(r.Line(), "; go1.27.1") {
		t.Fatalf("human line: %s", r.Line())
	}
}
