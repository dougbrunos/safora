# 05: REST API & Real-time SSE Telemetry Streaming

**What to build:**
An embedded HTTP REST and Server-Sent Events (SSE) server exposing endpoints for Job management, script ingestion, Run triggering, and live execution telemetry. The `/api/stream` endpoint broadcasts active transfer events (current file, progress percentage, bytes transferred, speed, and logs) in real time to connected web clients.

**Blocked by:** 03 (Date-Stamped Mirroring Backup Engine), 04 (Retention Engine)

**Status:** ready-for-agent

- [ ] REST API endpoints for Jobs (`GET /api/jobs`, `POST /api/jobs`, `GET /api/jobs/:id`, `DELETE /api/jobs/:id`).
- [ ] Endpoint `POST /api/jobs/:id/run` to trigger immediate Job execution in a background goroutine.
- [ ] Endpoint `POST /api/importer/parse` accepting raw `.bat` or `robocopy` text and returning parsed Job JSON.
- [ ] Endpoint `GET /api/runs` and `GET /api/runs/:id` returning Run execution history and structured logs.
- [ ] SSE endpoint `GET /api/stream` with pub/sub broker streaming real-time Run progress and log events.
- [ ] Localhost bind security with optional authorization token support.
