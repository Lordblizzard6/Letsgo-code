@echo off
setlocal
cd /d "%~dp0"

rem ============================================================
rem  LetsGO build script (Windows)
rem  Builds the LetsGO binary (TUI + Wails GUI embedded in the
rem  same executable). Optionally runs the full test suite.
rem
rem  Usage:
rem    build.bat             build the app
rem    build.bat test        build and run all tests
rem    build.bat -n          build only, skip tests
rem ============================================================

set "OUT=letsgo.exe"

echo === LetsGo build ===
echo OUT  : %OUT%
echo NOTE : the Wails GUI runs via `letsgo gui` (no CGO/Fyne toolchain)

go build -o "%OUT%" ./ || goto :fail

if "%1"=="-n" goto :done
if "%1"=="test" goto :test

:done
echo.
echo Build OK: %OUT%
exit /b 0

:test
echo.
echo === Running tests ===
go test ./...
if errorlevel 1 goto :fail
echo.
echo Tests OK
exit /b 0

:fail
echo.
echo *** BUILD FAILED ***
exit /b 1