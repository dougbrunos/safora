# 06: Embedded Web Dashboard & Script Importer UI

**What to build:**
A responsive, modern web interface matching the specifications in `ui/SAFORA_DESIGN_SYSTEM.md` (dark theme `#0B1220`, teal accents `#00D1B2`, Outfit and JetBrains Mono typography, Lucide icons), built with React, Vite, and Tailwind, and packaged directly into the Go executable via `embed.FS`. The UI provides an operational Dashboard (backup status cards, recent runs, upcoming schedules), active Jobs management, Run history with log viewer, real-time live progress bar during active backups, and the "Import Script / Robocopy" modal.

**Blocked by:** 05 (REST API & Real-time SSE Telemetry Streaming)

**Status:** done

- [x] Vite + React + Tailwind frontend setup adhering to design tokens in `ui/SAFORA_DESIGN_SYSTEM.md`.
- [x] Operational Dashboard displaying system status cards (Backups functioning, failed, active destinations, storage usage).
- [x] Real-time Run monitor subscribing to `/api/stream` with live progress bar (`42% concluído`), current file indicator, and throughput metrics.
- [x] Jobs management view with Job details and "Run Now" action.
- [x] "Import Script" modal allowing users to paste legacy `.bat` / `robocopy` scripts, preview the extracted Job configuration, and save with one click.
- [x] Go binary embedding of the compiled frontend dist directory (`embed.FS`), serving the web app automatically on `http://localhost:3434`.
