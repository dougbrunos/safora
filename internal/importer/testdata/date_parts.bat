@echo off
set date1=today
set qty=-1
set separator=%~3
if /i "%date1%" EQU "TODAY" (set date1=now) else (set date1="%date1%")
echo >"%temp%\%~n0.vbs" s=DateAdd("d",%qty%,%date1%)
echo>>"%temp%\%~n0.vbs" d=weekday(s)
echo>>"%temp%\%~n0.vbs" WScript.Echo year(s)^&_
for /f %%a in ('cscript //nologo "%temp%\%~n0.vbs"') do set result=%%a
del "%temp%\%~n0.vbs"
endlocal& (
set "YY=%result:~0,4%"
set "MM=%result:~4,2%"
set "DD=%result:~6,2%"
)
    set unidade="F"
    cd \
    %unidade%:
    cd %unidade%:\Backup Incremental CH
    md %DD%-%MM%-%YY%
    set source="D:\BACKUPCHAPECO\Incremental\%DD%-%MM%-%YY%"
    set destin="%unidade%:\Backup Incremental CH\%DD%-%MM%-%YY%\%DD%-%MM%-%YY%CH"
    set directories="Cache"
    set logpath=D:\logs
    set filename=\IncrementalHD\%DATE:/=-%.txt
    cd\
    robocopy %source% %destin% /E /R:5 /W:5 /V /NP /XF /NDL /XD %directories% /LOG:"%logpath%%filename%" /ETA /TEE
pause
