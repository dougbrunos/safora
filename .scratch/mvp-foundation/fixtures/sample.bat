@echo off
set qty=-1
echo WScript.Echo DateAdd("d", %qty%, Date) > "%temp%\yesterday.vbs"
for /f "tokens=1-3 delims=/" %%a in ('cscript //nologo "%temp%\yesterday.vbs"') do (
  set DD=%%a
  set MM=%%b
  set YY=%%c
)
set YESTERDAY=%DD%-%MM%-%YY%

set SOURCE=C:\Data\Production\%YESTERDAY%
set DEST=D:\Backups\Daily\%YESTERDAY%

robocopy "%SOURCE%" "%DEST%" /MIR /XD 123LAUDOS123 temp /XF *.tmp *.bak /R:5 /W:5 /LOG:C:\logs\backup.log
