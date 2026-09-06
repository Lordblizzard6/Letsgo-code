@echo off
setlocal
cd /d "%~dp0"

rem ============================================================
rem  LetsGO GUI build & run script (Windows)
rem  Compiles the frontend (Vite) + Go binary, then launches
rem  `letsgo gui`.
rem
rem  Usage:
rem    build-gui.bat         build frontend + binary, then run the GUI
rem    build-gui.bat build   build only (frontend + binary), do not run
rem ============================================================

set "FRONTEND=cmd\wails\frontend"
set "OUT=letsgo.exe"

echo === LetsGo GUI build ===

if not exist "%FRONTEND%\node_modules" (
    echo Installing frontend dependencies...
    pushd "%FRONTEND%"
    npm install
    if errorlevel 1 goto :fail
    popd
)

echo Building frontend (Vite)...
pushd "%FRONTEND%"
npm run build
if errorlevel 1 goto :fail
popd

echo Building Go binary: %OUT%
go build -o "%OUT%" ./ || goto :fail
echo.
echo Build OK: %OUT%

if "%1"=="build" exit /b 0

echo.
echo Starting LetsGO GUI...
"%OUT%" gui
if errorlevel 1 goto :fail

exit /b 0

:fail
echo.
echo *** BUILD/GUI FAILED ***
exit /b 1