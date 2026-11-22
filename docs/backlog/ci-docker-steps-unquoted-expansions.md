---
worth: later
where: .github/workflows/ci.yml
added: 2026-11-22
---
# docker steps in CI use unquoted shell expansions

Both "build and push master image" and "build and push tagged image" use `${DOCKER_HUB_TOKEN}` and
`${GIT_TAG}` unquoted (shellcheck SC2086, reported by actionlint). No failure today: the token and tag names
carry no spaces or globs. Quoting both steps together keeps them consistent. Raised in
https://github.com/radio-t/tg-retrans/issues/7 and left out of #8 to keep that fix to the trigger and token.
