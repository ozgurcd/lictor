package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdversarialPin(t *testing.T) {
	for _, tc := range []struct{ name, body, want string }{
		{"comment", "# LICTOR_VERSION: v0.4.0\n", ""},
		{"job-env", "jobs:\n  verify:\n    env:\n      LICTOR_VERSION: v0.4.0\n", ""},
		{"duplicate", "env:\n  LICTOR_VERSION: v0.4.0\n  LICTOR_VERSION: v99.0.0\n", "v0.4.0"},
		{"quoted", "env:\n  LICTOR_VERSION: \"v0.4.0\"\n", ""},
		{"trailing-space", "env:\n  LICTOR_VERSION: v0.4.0   \n", "v0.4.0"},
		{"invalid-yaml", "[unterminated\n  LICTOR_VERSION: v0.4.0\n", "v0.4.0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			declarePin(t, root, version)
			if err := os.WriteFile(filepath.Join(root, ".github/workflows/ci.yml"), []byte(tc.body), 0600); err != nil {
				t.Fatal(err)
			}
			v, err := declaredVersion(root, "LICTOR_VERSION")
			got := ""
			if v != nil {
				got = *v
			}
			t.Logf("declaration=%q error=%v", got, err)
			if err != nil || got != tc.want {
				t.Fatalf("got %q, %v; want %q", got, err, tc.want)
			}
		})
	}
	for _, kind := range []string{"directory", "outside-symlink", "oversize"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			declarePin(t, root, version)
			path := filepath.Join(root, ".github/workflows/ci.yml")
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			var err error
			switch kind {
			case "directory":
				err = os.Mkdir(path, 0700)
			case "outside-symlink":
				outside := filepath.Join(t.TempDir(), "outside.yml")
				if err := os.WriteFile(outside, []byte("  LICTOR_VERSION: "+version+"\n"), 0600); err != nil {
					t.Fatal(err)
				}
				err = os.Symlink(outside, path)
			case "oversize":
				err = os.WriteFile(path, []byte(strings.Repeat("x", (1<<20)+1)), 0600)
			}
			if err != nil {
				t.Fatal(err)
			}
			v, err := declaredVersion(root, "LICTOR_VERSION")
			t.Logf("declaration=%v error=%v", v, err)
			if err == nil || v != nil {
				t.Fatalf("unsafe declaration accepted: %v %v", v, err)
			}
		})
	}
}
