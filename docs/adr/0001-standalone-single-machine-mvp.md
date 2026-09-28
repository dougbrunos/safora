# 0001 - Standalone Single-Machine Architecture for MVP

To solve the immediate pain of manual backup routines (such as Windows `.bat` scripts) without requiring distributed server infrastructure, we decided to build the MVP as a standalone single-binary application written in Go with an embedded SQLite database and embedded React web interface. This delivers immediate, zero-friction local automation on a single machine while structuring the internal core so it can later split into centralized server and agent daemons.
