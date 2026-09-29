; Inno Setup script for the Safora installer (https://jrsoftware.org/isinfo.php).
; Build from the repository root, after scripts/build-release.sh has produced the executable:
;   iscc /DVersion=1.0.0 installer\windows\safora.iss
; Output: dist\Safora-Setup-<version>.exe

#ifndef Version
  #define Version "0.0.0"
#endif

[Setup]
AppId={{6F6C7D0E-5A3B-4C0E-9C55-5AFE0A1B0C01}
AppName=Safora
AppVersion={#Version}
AppPublisher=Safora
DefaultDirName={autopf}\Safora
DefaultGroupName=Safora
PrivilegesRequired=admin
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
OutputDir=..\..\dist
OutputBaseFilename=Safora-Setup-{#Version}
SetupIconFile=..\..\ui\web\public\favicon.ico
UninstallDisplayIcon={app}\safora.exe
Compression=lzma2
SolidCompression=yes
WizardStyle=modern
DisableProgramGroupPage=yes

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"
Name: "brazilianportuguese"; MessagesFile: "compiler:Languages\BrazilianPortuguese.isl"

[Files]
Source: "..\..\dist\windows-amd64\safora.exe"; DestDir: "{app}"; Flags: ignoreversion

; A .url shortcut opens the dashboard in the default browser.
[INI]
Filename: "{app}\Safora.url"; Section: "InternetShortcut"; Key: "URL"; String: "http://127.0.0.1:3434"
Filename: "{app}\Safora.url"; Section: "InternetShortcut"; Key: "IconFile"; String: "{app}\safora.exe"
Filename: "{app}\Safora.url"; Section: "InternetShortcut"; Key: "IconIndex"; String: "0"

[Icons]
Name: "{group}\Safora"; Filename: "{app}\Safora.url"
Name: "{commondesktop}\Safora"; Filename: "{app}\Safora.url"; Tasks: desktopicon

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"

; Jobs and history live in %ProgramData%\Safora and are not touched by the uninstaller.
[Run]
Filename: "{app}\safora.exe"; Parameters: "service install"; Flags: runhidden; StatusMsg: "Installing the Safora service..."
Filename: "{app}\safora.exe"; Parameters: "service start"; Flags: runhidden; StatusMsg: "Starting the Safora service..."
Filename: "{app}\Safora.url"; Flags: shellexec postinstall skipifsilent nowait; Description: "Open the Safora dashboard"

[UninstallRun]
Filename: "{app}\safora.exe"; Parameters: "service stop"; Flags: runhidden; RunOnceId: "StopService"
Filename: "{app}\safora.exe"; Parameters: "service uninstall"; Flags: runhidden; RunOnceId: "RemoveService"

[UninstallDelete]
Type: files; Name: "{app}\Safora.url"

[Code]
// On upgrade the running service locks safora.exe: stop it before files are replaced.
function PrepareToInstall(var NeedsRestart: Boolean): String;
var
  ResultCode: Integer;
begin
  Exec(ExpandConstant('{sys}\sc.exe'), 'stop Safora', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
  Sleep(2000);
  Result := '';
end;
