package witness

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Git only measures. No write command or shell is available to the recorder.
func gitRead(ctx context.Context, root, input string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", append([]string{"-c", "core.fsmonitor=false"}, args...)...)
	cmd.Dir = root
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME"), "GIT_OPTIONAL_LOCKS=0", "LC_ALL=C"}
	cmd.Stdin = strings.NewReader(input)
	var out bounded
	cmd.Stdout = &out
	err := cmd.Run()
	if out.over {
		return "", fmt.Errorf("git output exceeds 8 MiB")
	}
	return strings.TrimSuffix(out.String(), "\n"), err
}

type bounded struct {
	bytes.Buffer
	over bool
}

func (b *bounded) Write(p []byte) (int, error) {
	n := len(p)
	if b.Len()+n > 8<<20 {
		b.over = true
		return n, nil
	}
	return b.Buffer.Write(p)
}

func state(ctx context.Context, root, exclude string, work bool) (string, error) {
	head, err := gitRead(ctx, root, "", "rev-parse", "--short", "HEAD")
	if err != nil {
		return "", fmt.Errorf("repository HEAD: %w", err)
	}
	args := []string{"status", "--porcelain", "--", ".", ":(exclude)" + exclude}
	if work {
		args = append(args, ":(exclude)GATE-RUN*.txt")
	}
	dirty, err := gitRead(ctx, root, "", args...)
	// The legacy external-record pathspec is invalid. Its header contains HEAD
	// alone; its diagnostic-only record carries EMPTY-TREE. Preserve the bytes,
	// explicitly document this limitation; never present it as a usable witness.
	if err != nil {
		if outside(root, exclude) {
			return head, nil
		}
		return "", err
	}
	if dirty != "" {
		head += " (dirty)"
	}
	return head, nil
}

func outside(root, path string) bool {
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	rel, err := filepath.Rel(root, path)
	return err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func digest(ctx context.Context, root, exclude string) (string, error) {
	files := map[string]bool{}
	for _, kind := range []string{"--cached", "--others"} {
		s, err := gitRead(ctx, root, "", "ls-files", kind, "--exclude-standard", "--", ".", ":(exclude)"+exclude)
		if err != nil {
			if outside(root, exclude) {
				return "EMPTY-TREE", nil
			}
			return "", err
		}
		for _, f := range strings.Split(s, "\n") {
			if st, e := os.Stat(filepath.Join(root, f)); e == nil && st.Mode().IsRegular() {
				files[f] = true
			}
		}
	}
	if len(files) == 0 {
		return "EMPTY-TREE", nil
	}
	var list []string
	for f := range files {
		list = append(list, f)
	}
	sort.Strings(list)
	hashes, err := gitRead(ctx, root, strings.Join(list, "\n")+"\n", "hash-object", "--stdin-paths")
	if err != nil {
		return "", err
	}
	hs := strings.Split(hashes, "\n")
	if len(hs) != len(list) {
		return "", fmt.Errorf("git hash-object returned an incomplete inventory")
	}
	h := sha256.New()
	for i, f := range list {
		fmt.Fprintf(h, "%s %s\n", hs[i], f)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
