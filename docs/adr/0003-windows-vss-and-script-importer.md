# 0003 - Windows VSS Integration and Legacy Script Ingestion

To guarantee 100% data consistency for databases and locked files without disrupting ongoing workstation operations, we decided to integrate native Windows Volume Shadow Copy (VSS) support as an opt-in per-Job capability with automatic retry/skip fallback. Additionally, to facilitate immediate zero-friction migration from manual batch scripts, Safora includes a dedicated Script Importer that parses legacy `robocopy` and `.bat` command lines directly into declarative Job definitions.
