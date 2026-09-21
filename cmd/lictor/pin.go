package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type pinMetadata struct {
	DeclaredVersion *string `json:"declared_version"`
	Pinned          bool    `json:"pinned"`
}

// declaredVersion reads the owner's exact line convention, not YAML semantics.
// A malformed or unreadable input is not permission to run unpinned.
func declaredVersion(root, key string) (*string, error) {
	r, err := os.OpenRoot(root)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	f, err := r.Open(".github/workflows/ci.yml")
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !st.Mode().IsRegular() {
		return nil, fmt.Errorf("ci.yml is not a regular file")
	}
	b, err := io.ReadAll(io.LimitReader(f, 1<<20+1))
	if err != nil {
		return nil, err
	}
	if len(b) > 1<<20 {
		return nil, fmt.Errorf("ci.yml exceeds 1 MiB")
	}
	m := regexp.MustCompile(`(?m)^  ` + regexp.QuoteMeta(key) + `: (v[0-9][0-9.]*)`).FindSubmatch(b)
	if m == nil {
		return nil, nil
	}
	v := string(m[1])
	return &v, nil
}

func checkPin(root string, unpinned bool, errOut io.Writer) (pinMetadata, bool) {
	declared, err := declaredVersion(root, "LICTOR_VERSION")
	p := pinMetadata{DeclaredVersion: declared}
	prefix := "lictor " + version + ": "
	base := filepath.Base(root)
	if err != nil {
		fmt.Fprintf(errOut, "%srefused — %s: cannot read .github/workflows/ci.yml: %s\n", prefix, base, strings.ReplaceAll(err.Error(), "\n", " "))
		return p, false
	}
	if declared != nil {
		if *declared != version {
			fmt.Fprintf(errOut, "%srefused — %s declares LICTOR_VERSION %s; install the declared version: brew install ozgurcd/tap/lictor\n", prefix, base, *declared)
			return p, false
		}
		p.Pinned = true
		return p, true
	}
	if !unpinned {
		fmt.Fprintf(errOut, "%srefused — %s declares no LICTOR_VERSION in .github/workflows/ci.yml env; pass --unpinned to run without a pin\n", prefix, base)
		return p, false
	}
	fmt.Fprintf(errOut, "%sunpinned — %s declares no LICTOR_VERSION; running with --unpinned\n", prefix, base)
	return p, true
}
