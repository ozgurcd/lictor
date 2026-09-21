package main

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWitnessCLIRefusalsAndHelp(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run(context.Background(), []string{"witness", "--help"}, &out, &errOut, time.Now)
	if code != 0 || !strings.Contains(out.String(), "default 2m0s") || !strings.Contains(out.String(), "check is not implemented") {
		t.Fatalf("help=%d %s %s", code, &out, &errOut)
	}
	for _, entry := range []string{"a=echo a | cat", "a=bash -c 'echo hi'"} {
		out.Reset()
		errOut.Reset()
		root := t.TempDir()
		code = run(context.Background(), []string{"witness", "--repo", root, "--unpinned", "--record", filepath.Join(root, "record"), "--json", "--", entry}, &out, &errOut, time.Now)
		var doc struct {
			Schema string `json:"schema_version"`
			Reason string `json:"reason"`
			Code   int    `json:"exit_code"`
			Pinned bool   `json:"pinned"`
		}
		if code != 2 || json.Unmarshal(out.Bytes(), &doc) != nil || doc.Schema != "lictor.witness.v1" || doc.Code != 2 || doc.Pinned || !strings.Contains(doc.Reason, "target a") {
			t.Fatalf("refusal=%d %s %s", code, &out, &errOut)
		}
		validateDocument(t, "lictor.witness.v1", out.String())
	}
}
