# 04: Retention Engine with Retention Lock Safety Guard

**What to build:**
An automated retention and cleanup subsystem that prunes historical backup copies at Destination targets according to user-defined retention policies (e.g., keep last N backups, keep backups within X days). The engine strictly enforces the **Retention Lock**: if the current or latest backup Run failed, or if pruning would leave fewer than one intact backup copy, all deletion operations are halted and an explanatory warning is logged.

**Blocked by:** 03 (Date-Stamped Mirroring Backup Engine)

**Status:** done

- [x] Implementation of `RetentionPolicy` evaluator supporting count-based and age-based rules.
- [x] Scanning and identification of date-stamped backup folders at local and NAS destinations.
- [x] Retention Lock enforcement: verification that the current Run was successful and that at least one valid backup copy remains before any pruning occurs.
- [x] Post-run hook integrated into the Job runner to automatically trigger retention cleanup upon successful Run completion.
- [x] Comprehensive unit tests verifying prune calculations, edge cases, and that Retention Lock prevents data loss during run failures.
