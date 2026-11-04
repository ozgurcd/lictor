Release-notes fragment (LICTOR-FIXES-1104, for the next patch release):

A refusal caused by a declared `LICTOR_VERSION` pin that differs from the running
version now says that `--unpinned` does not override a declared pin, and how to
proceed: install the declared version, or remove `LICTOR_VERSION` from
`.github/workflows/ci.yml` and pass `--unpinned`. Pin behavior is unchanged: a
declared pin still wins.

`witness run` now prints one stderr line naming the inherited environment
variables it does not pass to gate targets (for example `GOFLAGS`, `GOMAXPROCS`),
names only. The README documents the full allowlist; a test keeps it in sync.
Records and exit codes are unchanged.
