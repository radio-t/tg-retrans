---
worth: maybe
where: main.go
added: 2026-11-22
---
# main.go has three golangci-lint findings under the personal config

The repo has no `.golangci.yml`, so `golangci-lint run` picks up `~/.golangci.yml` and reports three items
the default linters (what #3 cleaned against) do not enable:

- `main` - gocritic exitAfterDefer: `log.Fatalf` after `defer stop()` exits without running `stop()`.
  Harmless at process exit, since the signal context dies with the process.
- `startRetrans` - gosec G204: `exec.CommandContext(ctx, opts.FfmpegPath, runOpts...)`. The path comes
  from operator config, not untrusted input.
- `run` - misspell: "cancelled" in a comment.

Undecided whether to fix these or add a repo `.golangci.yml` that pins the intended linter set. CI does not
run golangci-lint at all (its step is commented out in `.github/workflows/ci.yml`). Surfaced while fixing
https://github.com/radio-t/tg-retrans/issues/7.
