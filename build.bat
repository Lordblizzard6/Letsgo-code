@echo off
setlocal
cd /d "%~dp0"

rem ============================================================
rem  LetsGO build script (Windows)
rem  Builds the LetsGO desktop app binary with CGO (Fyne needs
rem  a C toolchain). Optionally runs the full test suite.
rem
rem  Usage:
rem    build.bat             build the app
rem    build.bat test        build and run all tests
rem    build.bat -n          build only, skip tests
rem ============================================================

rem --- C toolchain for CGO (Fyne/GLFW) ------------------------
set "UCRT64=D:\msys64\ucrt64\bin"
if exist "%UCRT64%\gcc.exe" (
    set "PATH=%UCRT64%;%PATH%"
    set "CGO_ENABLED=1"
) else (
    echo [warn] gcc not found at "%UCRT64%;"
    echo [warn] falling back to CGO_DISABLED. The GUI may not build.
    set "CGO_ENABLED=0"
)

set "OUT=letsgo.exe"

echo === LetsGo build ===
echo OUT  : %OUT%
echo CGO  : %CGO_ENABLED%

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