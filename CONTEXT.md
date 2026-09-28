# Safora

Open-source automated backup management system designed to eliminate manual backup routines across computers, servers, NAS, and cloud storage.

## Language

**Job**:
A configured backup routine that defines sources, destinations, storage strategy, and execution schedule.
_Avoid_: Task, batch, script, routine

**Source**:
A local filesystem directory or path designated for backup protection.
_Avoid_: Origin, input path, folder

**Destination**:
A designated storage location—such as a local disk, network share (NAS), or cloud bucket—where backup copies are placed.
_Avoid_: Target, sink, output drive

**Run**:
A single execution instance of a Job with captured state, duration, transfer metrics, and diagnostic logs.
_Avoid_: Execution, process, attempt

**Storage Strategy**:
The structural format used to store backup copies at the Destination, specifically Date-Stamped Mirroring or Snapshot Repository.
_Avoid_: Backup mode, format, type

**Retention Policy**:
The set of rules governing how long historical Runs and old Destination copies are preserved before automatic pruning.
_Avoid_: Rotation rule, cleanup script, purge policy

**Trigger**:
A condition or event—such as a scheduled timestamp, external drive arrival, or manual command—that initiates a Job Run.
_Avoid_: Starter, launcher, invoker

**Path Template**:
A parameterized directory expression in a Source or Destination that evaluates dynamic runtime variables such as date offsets or hostnames.
_Avoid_: Dynamic string, variable folder, wildcard path

**Retention Lock**:
A safety guard that unconditionally halts pruning when the latest Run failed or when pruning would eliminate the sole surviving backup copy.
_Avoid_: Safety guard, deletion block, preserve flag

**Volume Shadow Copy**:
A Windows operating system snapshot capability allowing point-in-time capture of files currently open or locked by other applications.
_Avoid_: Shadow drive, file unlocker, locked file copy

**Exclusion Rule**:
A criteria pattern (directory name, file extension, or glob expression) that exempts matching filesystem items from a Run.
_Avoid_: Ignore list, filter rule, blacklist

**Script Importer**:
A configuration utility that parses legacy batch scripts and command-line parameters (such as robocopy or rsync) directly into structured Job definitions.
_Avoid_: Batch converter, script translator

**Live Stream**:
The real-time telemetry event stream delivering instantaneous progress, file activity, throughput, and diagnostic output to connected clients.
_Avoid_: Progress poller, status socket

**Drive Watcher**:
A background operating system monitor that detects the arrival or mounting of external storage volumes (such as USB drives or network letters) to invoke designated Jobs.
_Avoid_: USB listener, disk detector, drive trigger
