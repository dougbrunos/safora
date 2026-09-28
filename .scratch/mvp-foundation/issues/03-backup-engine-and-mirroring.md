# 03: Date-Stamped Mirroring Backup Engine with Exclusions & Telemetry

**What to build:**
The core execution engine that runs a Job end-to-end. It evaluates dynamic Source and Destination paths, executes recursive file synchronization with configurable retries, skips directories and files matching Exclusion Rules, tracks transfer telemetry (bytes copied, files processed, throughput, ETA), handles Windows VSS snapshot triggering when enabled, and records the Run result with structured logs in the SQLite database. A user or agent can run a backup via `safora run <job-id>`.

**Blocked by:** 01 (Project Foundation & Path Template Engine), 02 (Legacy Batch Script & Robocopy Importer)

**Status:** done

- [x] Implementation of `BackupEngine` interface with `DateStampedMirroring` strategy.
- [x] Recursive filesystem traversal with directory exclusion matching (`123LAUDOS123`, `node_modules`, etc.) and file patterns.
- [x] Retry loop for busy/locked files with configurable attempt count and delay interval.
- [x] Telemetry tracker calculating transferred bytes, file counts, and instantaneous transfer rate.
- [x] Run state recorder persisting start/finish timestamps, final status (`success`, `warning`, `failed`), and structured event logs into SQLite.
- [x] CLI command `safora run <job-id>` with styled terminal output showing live progress and completion summary.
