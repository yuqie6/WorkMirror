@echo off
setlocal

rem Wrapper to avoid PowerShell profile parse issues (-NoProfile).
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0build-windows-zip.ps1" %*

