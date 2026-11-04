package executor

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

func sortedCopy(in []string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	return out
}

func TestDroppedEnvironmentNamesOnly(t *testing.T) {
	t.Setenv("LIC02_DECLARED", "declared-value-123")
	g, err := DeclareEnvironment([]string{"LIC02_DECLARED"})
	if err != nil {
		t.Fatal(err)
	}
	environ := []string{
		"GOMAXPROCS=2", "GOFLAGS=-mod=mod", "PATH=/bin", "HOME=/h", "GOPROXY=off",
		"GRYPE_DB_CACHE_DIR=/c", "LIC02_DECLARED=declared-value-123",
		"API_TOKEN=secret-token-value", "API_TOKEN=duplicate-entry", "LC_ALL=de_DE", "GOWORK=/x/go.work",
		"bad name=oops", "NOEQUALS",
	}
	got := DroppedEnvironment(environ, g)
	want := sortedCopy([]string{"NOEQUALS", `"bad name"`, "API_TOKEN", "GOFLAGS", "GOMAXPROCS", "GOWORK", "LC_ALL"})
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("dropped = %q, want %q", got, want)
	}
	if strings.Contains(strings.Join(got, " "), "secret-token-value") {
		t.Fatal("a value leaked into the dropped list")
	}
}

func TestDroppedEnvironmentNilDeclared(t *testing.T) {
	if got := DroppedEnvironment([]string{"PATH=/bin", "GOFLAGS=x"}, nil); !reflect.DeepEqual(got, []string{"GOFLAGS"}) {
		t.Fatalf("got %q", got)
	}
	if got := DroppedEnvironment(nil, nil); len(got) != 0 {
		t.Fatalf("empty environ: %q", got)
	}
}

func TestAllowedEnvironmentIsTheInheritedSet(t *testing.T) {
	got := AllowedEnvironment()
	for _, name := range append(append([]string{}, goEnvironmentNames...), recordEnvironmentNames...) {
		found := false
		for _, g := range got {
			found = found || g == name
		}
		if !found {
			t.Fatalf("%s missing from AllowedEnvironment", name)
		}
	}
	if len(got) == 0 {
		t.Fatal("empty allowlist")
	}
	got[0] = "MUTATED"
	if AllowedEnvironment()[0] == "MUTATED" {
		t.Fatal("AllowedEnvironment returns shared state")
	}
}
