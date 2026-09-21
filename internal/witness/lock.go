package witness

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func paths(record string) (string, string, error) {
	dir, e := filepath.EvalSymlinks(filepath.Dir(record))
	if e != nil {
		return "", "", e
	}
	key := fmt.Sprintf("%x", sha256.Sum256([]byte(filepath.Join(dir, filepath.Base(record)))))
	return "/tmp/gate-witness-" + key + ".lock", "/tmp/gate-witness-" + key + ".session", nil
}

func live(owner, key string) bool {
	_, s, ok := strings.Cut(owner, key+"=")
	if !ok {
		return false
	}
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return false
	}
	n, err := strconv.Atoi(fields[0])
	if err != nil || n <= 0 {
		return false
	}
	err = syscall.Kill(n, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}

func sessionGuard(sf, rec, what string, diag io.Writer) (int, error) {
	b, err := os.ReadFile(sf)
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 2, err
	}
	owner := strings.TrimSpace(string(b))
	if !live(owner, "owner") {
		fmt.Fprintf(diag, "gate-witness: breaking a STALE session on %s — its owner (%s) is gone\n", rec, owner)
		return 0, os.Remove(sf)
	}
	fmt.Fprintf(diag, "GATE-WITNESS SESSION OPEN: %s belongs to a stepwise session that is still running (%s)\n  Refusing to start %s: it would truncate that session's header and leave two runs in one record.\n  Wait for its finalize, or if that session is dead remove %s and try again.\n", rec, owner, what, sf)
	return 4, fmt.Errorf("GATE-WITNESS SESSION OPEN: %s", rec)
}

func acquire(ctx context.Context, lock, rec, label string, wait time.Duration, now func() time.Time, diag io.Writer) (func(), int, error) {
	seconds := int(wait / time.Second)
	for waited := 0; ; {
		err := os.Mkdir(lock, 0700)
		if err == nil {
			if err = os.WriteFile(filepath.Join(lock, "owner"), []byte(fmt.Sprintf("pid=%d label=%s started=%s\n", os.Getpid(), label, stamp(now()))), 0600); err != nil {
				_ = os.RemoveAll(lock)
				return nil, 2, err
			}
			return func() { _ = os.RemoveAll(lock) }, 0, nil
		}
		if !os.IsExist(err) {
			return nil, 2, err
		}
		b, readErr := os.ReadFile(filepath.Join(lock, "owner"))
		owner := strings.TrimSpace(string(b))
		if readErr != nil {
			owner = "unknown"
		}
		if strings.HasPrefix(owner, "pid=") && !live(owner, "pid") {
			fmt.Fprintf(diag, "gate-witness: breaking a STALE lock on %s — its holder (%s) is gone\n", rec, owner)
			if e := os.RemoveAll(lock); e != nil {
				return nil, 2, e
			}
			continue
		}
		if waited >= seconds {
			fmt.Fprintf(diag, "GATE-WITNESS LOCKED: %s is held by a live run; waited %ds and refusing\n  (holder: %s)\n  Refusing rather than interleaving: two runs in one record witness nothing.\n  If that run is dead, remove %s and try again.\n", rec, waited, owner, lock)
			return nil, 3, fmt.Errorf("GATE-WITNESS LOCKED: %s", rec)
		}
		if waited == 0 {
			fmt.Fprintf(diag, "gate-witness: waiting for the lock on %s (bound %ds)\n", rec, seconds)
		}
		timer := time.NewTimer(time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, 2, ctx.Err()
		case <-timer.C:
			waited++
		}
	}
}
