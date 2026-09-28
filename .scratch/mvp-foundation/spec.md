# Spec: Safora MVP — Automated Backup Management (Foundation & Core)

Status: ready-for-agent

## Problem Statement

System administrators, workstations, and small-to-medium offices rely heavily on manual or fragile backup routines, commonly executed via manual batch scripts (e.g., Windows `.bat` executing `robocopy` and VBScript date calculators). These legacy routines suffer from critical vulnerabilities:
- They require manual double-clicks by a logged-in user, meaning backups are missed when machines are locked, users are away, or operators forget.
- They lack visibility and automated failure alerting: errors (such as disconnected drives or locked files) are buried in obscure local text logs without immediate operator notice.
- They create unmanaged disk bloat: folders are created daily (e.g., `DD-MM-YYYY`) and accumulate indefinitely until target disks fill up, with no retention pruning.
- Open or locked files (spreadsheets, database files, active clinical records) fail or get skipped during standard file copies.
- Setup and migration from existing batch routines is manual, tedious, and error-prone.

## Solution

Safora provides a standalone, automated, open-source backup management application designed to eliminate manual batch routines on workstations and servers.
Operating as a background system service with an embedded local Web UI and Windows Toast notifications, Safora delivers:
1. Automated scheduling (cron intervals, daily runs, and automatic device-arrival triggers when target drives or USB disks are connected).
2. Dynamic path template resolution (evaluating expressions like `{yesterday}`, `{today:DD-MM-YYYY}`, and wildcard patterns) to seamlessly ingest third-party system outputs.
3. A declarative Backup Engine supporting date-stamped mirroring, retries, and comprehensive directory/file exclusion rules.
4. Volume Shadow Copy (VSS) integration on Windows to guarantee point-in-time consistency for locked and active files.
5. A Retention Policy engine protected by a Retention Lock safeguard that prevents pruning when runs fail or when only a single backup copy remains.
6. A Legacy Script Importer that parses existing `robocopy` commands and `.bat` scripts directly into structured Jobs.
7. Real-time telemetry via Server-Sent Events (SSE) streaming progress and transfer metrics to the embedded Web UI, styled according to the Safora Design System.

## User Stories

1. As a workstation operator, I want my scheduled backup to run automatically in the background even when I am logged out or the workstation is locked, so that my data is protected without manual intervention.
2. As a system administrator, I want to paste my existing Windows `.bat` script or `robocopy` command into an importer, so that I can migrate my existing backup routine into Safora in seconds without retyping paths and flags.
3. As a user with third-party software generating folders named after yesterday's date, I want Safora to resolve `{yesterday:DD-MM-YYYY}` in the source path dynamically, so that the correct daily files are captured automatically.
4. As an operator using an external USB hard drive or NAS letter, I want Safora to trigger the backup Job automatically the moment the drive is plugged in or mounted, so that backups happen even if the drive was disconnected during the scheduled cron time.
5. As an administrator, I want to configure folder and file exclusion rules (such as skipping specific directory trees and temporary files), so that unnecessary data does not consume storage bandwidth.
6. As a business owner backing up active database files and open documents, I want Safora to use Volume Shadow Copy (VSS) snapshots when running on Windows, so that locked files are backed up consistently without application errors.
7. As an administrator, I want old backup folders to be pruned automatically according to retention rules (e.g., keep the last 30 days), so that destination disks never run out of space.
8. As a safety-conscious user, I want the Retention Lock to halt pruning whenever the current backup run fails, so that I never lose my last good backup copy due to an automated cleanup.
9. As a workstation user, I want native desktop toast notifications when a backup completes or fails, so that I am immediately aware of issues without having to inspect log files.
10. As an operator watching an active backup, I want a real-time progress bar, throughput gauge, and current file indicator in the local Web dashboard, so that I know exactly what the engine is transferring.
11. As an administrator, I want to inspect historical backup runs with structured logs and search/filter capabilities in the Web UI, so that I can audit backup health and verify file transfer details.
12. As a DevOps engineer or advanced user, I want a clean CLI interface to run jobs, inspect status, and manage the system service directly from the terminal, so that I can integrate Safora into automated operational scripts.
13. As an operator, I want the web interface to match the official Safora Design System (dark technical theme, teal accents, Outfit and JetBrains Mono typography, Lucide icons), so that the experience is professional, clean, and visually cohesive.
14. As an evaluator, I want Safora to run in portable mode from a single executable without mandatory system installation, so that I can test it immediately on any machine.

## Implementation Decisions

### 1. Unified Core Architecture
- The application is built in Go as a single binary distributing both the service daemon, the CLI interface, and the embedded Web UI (packaged via Go `embed.FS`).
- SQLite in WAL mode serves as the embedded database for Job definitions, Run history, and structured log events.
- Storage directories follow a hybrid model: defaulting to system-level locations (`%ProgramData%\Safora` on Windows, `/var/lib/safora` on Linux) with automatic local fallback when `--portable` is specified.

### 2. Path Template Resolution Engine
- A dedicated template evaluator parses runtime variables in path strings:
  - `{today}` / `{today:FORMAT}` (e.g., `DD-MM-YYYY`, `YYYY-MM-DD`)
  - `{yesterday}` / `{yesterday:FORMAT}` (offset by -1 day, matching the batch VBScript logic)
  - `{offset:-N:FORMAT}` (arbitrary day offsets)
  - `{hostname}` (local machine name)
- The engine validates that resolved source paths exist before triggering execution.

### 3. Backup Engine & Date-Stamped Mirroring
- The engine implements a pluggable runner interface with a default File Sync / Mirroring strategy.
- It copies file hierarchies recursively with configurable retry count and delay interval (matching `/R:N /W:N`).
- It applies Exclusion Rules matching folder names, file extensions, and glob patterns.
- On Windows, it invokes VSS snapshots when enabled on the Job, falling back gracefully to retry/skip if VSS is unavailable or unprivileged.
- It produces real-time transfer telemetry (bytes copied, current file name, files completed, elapsed time, transfer speed).

### 4. Legacy Script Importer
- A parser accepts raw Windows batch scripts or command lines containing `robocopy` and extracts:
  - Source directory and dynamic date patterns
  - Destination directory
  - Excluded directories (`/XD`) and files (`/XF`)
  - Retry policy (`/R` and `/W`)
  - Log output preferences
- The parser outputs a structured, declarative Job model ready for validation and storage.

### 5. Retention Engine & Retention Lock
- Evaluates destination folders following Date-Stamped Mirroring.
- Applies count-based ("keep last N runs") and age-based ("keep runs within X days") pruning.
- Strictly enforces the Retention Lock: pruning is bypassed if the current Run finished with status `failed`, or if pruning would leave zero valid copies.

### 6. Embedded Web Interface & Design System
- Built with React, Vite, and Tailwind CSS, following `ui/SAFORA_DESIGN_SYSTEM.md`.
- Color tokens: `safora-navy-950` (#0B1220), `safora-navy-900` (#0F172A), `safora-teal-500` (#00D1B2), and semantic status colors.
- Communicates with the Go backend via REST API for CRUD and Server-Sent Events (SSE) for real-time progress streaming.
- Includes a dedicated "Import Script" modal in the Job creation flow.

## Testing Decisions

### What Makes a Good Test
- Tests must verify observable behavior at module boundaries (e.g., given a path template and date, verify the resolved string; given an existing batch script, verify the parsed Job struct; given a failed run, verify that retention does not delete existing copies).
- Avoid testing internal private helper functions or mocking standard library filesystem calls when an in-memory or temporary filesystem can be used.

### Modules Tested
1. `pathresolver`: Evaluates `{yesterday}`, `{today}`, custom date formats, invalid expressions, and edge cases (month/year boundaries, leap years).
2. `importer`: Tests parsing of the user's exact `.bat` script, complex `robocopy` flags, quoted paths, and multiline scripts.
3. `backup`: Verifies recursive mirroring, retry loops, directory exclusions, transfer metrics calculation, and exit status classification.
4. `retention`: Verifies count/age pruning calculations and asserts that the Retention Lock halts deletions on failed runs or single-copy situations.
5. `api`: Tests Job CRUD endpoints and SSE telemetry event streaming.

## Out of Scope

- Centralized multi-tenant Hub-and-Spoke server/agent control plane (reserved for post-MVP / v2.0).
- Remote cloud storage drivers (S3, B2, Azure) for the initial standalone milestone (focus is local disks, USB, and NAS network shares).
- Outbound webhook and SMTP email alerts (reserved for v1.1; MVP uses local Web UI and Windows Toast notifications).
- Block-level deduplicating snapshot repository engine (MVP focuses on Date-Stamped Mirroring compatible with Windows Explorer / NAS).

## Further Notes

- The project canonical name is **Safora**.
- The sample batch script provided by the user serves as the benchmark integration test fixture for the Script Importer and Path Resolver.
