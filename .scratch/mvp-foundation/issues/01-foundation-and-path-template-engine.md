# 01: Project Foundation & Path Template Engine

**What to build:**
A foundational runnable Go application and SQLite database schema for Jobs and Runs, integrated with a dynamic Path Template Engine. The engine resolves runtime path variables such as `{today:FORMAT}`, `{yesterday:FORMAT}`, and `{hostname}`, enabling dynamic date calculation without brittle VBScript hacks. A user or agent can test this directly from the CLI via `safora template resolve "<pattern>"`.

**Blocked by:** None (can start immediately)

**Status:** done

- [x] Initialized `go.mod` with Go 1.24 and modular package structure (`cmd/safora`, `internal/database`, `internal/pathresolver`).
- [x] SQLite schema migrations for `jobs`, `sources`, `destinations`, `runs`, and `logs` tables using pure Go SQLite driver (modernc.org/sqlite).
- [x] Path Template Engine supporting `{today}`, `{today:DD-MM-YYYY}`, `{yesterday}`, `{yesterday:DD-MM-YYYY}`, `{offset:-N:FORMAT}`, and `{hostname}`.
- [x] Unit tests covering date boundaries (month-end, leap years, year rollover) and invalid template handling.
- [x] CLI command `safora template resolve "<pattern>"` demonstrating dynamic path output.
