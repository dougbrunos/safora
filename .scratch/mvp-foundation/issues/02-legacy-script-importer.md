# 02: Legacy Batch Script & Robocopy Importer

**What to build:**
A Script Importer that parses existing Windows `.bat` scripts and raw `robocopy` command lines into structured, declarative Job configurations. The importer detects dynamic date patterns (such as yesterday's date logic `qty=-1`), source paths, destination paths, directory exclusions (`/XD`), file exclusions (`/XF`), retry parameters (`/R` and `/W`), and log output paths, saving the resulting Job directly to SQLite. A user or agent can import a script via `safora job import <path-to-script>`.

**Blocked by:** 01 (Project Foundation & Path Template Engine)

**Status:** done

- [x] Parser module in `internal/importer` capable of extracting `robocopy` arguments and Windows batch variables (`%DD%-%MM%-%YY%`, `DateAdd("d",-1,...)`).
- [x] Maps detected date patterns into Safora path templates (`{yesterday:DD-MM-YYYY}`).
- [x] Ingests exclusion arguments (`/XD 123LAUDOS123`) into structured Exclusion Rules.
- [x] Ingests retry arguments (`/R:5 /W:5`) into Job execution configuration.
- [x] Unit test using the exact `.bat` script from the user prompt as an integration fixture.
- [x] CLI command `safora job import <file.bat>` printing the generated Job and persisting to SQLite.
