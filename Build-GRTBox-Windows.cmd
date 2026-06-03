@echo off
setlocal
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0scripts\build-windows.ps1"
set "EXITCODE=%ERRORLEVEL%"
endlocal
exit /b %EXITCODE%
