@echo off
setlocal EnableDelayedExpansion
cd /d "%~dp0"

set CGO_ENABLED=1

where gcc >nul 2>&1
if not errorlevel 1 goto :gcc_ok

set "MINGW="
for /d %%D in ("%LOCALAPPDATA%\Microsoft\WinGet\Packages\BrechtSanders.WinLibs.POSIX.UCRT_*") do (
  if exist "%%D\mingw64\bin\gcc.exe" set "MINGW=%%D\mingw64\bin"
)
if not defined MINGW if exist "C:\mingw64\bin\gcc.exe" set "MINGW=C:\mingw64\bin"
if not defined MINGW if exist "C:\TDM-GCC-64\bin\gcc.exe" set "MINGW=C:\TDM-GCC-64\bin"

if defined MINGW (
  set "PATH=%MINGW%;%PATH%"
) else (
  echo [build.bat] GCC not found. Fyne requires CGO/MinGW on Windows.
  echo   winget install BrechtSanders.WinLibs.POSIX.UCRT
  echo Or use run.bat which has the same requirement.
  pause
  exit /b 1
)

:gcc_ok
if exist logo.png copy /Y logo.png internal\assets\logo.png >nul

echo.
echo [build.bat] Building %CD%\RelayPane.exe ^(no console window^)...
go build -ldflags="-H=windowsgui -s -w" -o RelayPane.exe ./cmd/relaypane
if errorlevel 1 (
  echo.
  echo [build.bat] Build FAILED.
  pause
  exit /b 1
)

if not exist RelayPane.exe (
  echo [build.bat] Build reported success but RelayPane.exe was not created.
  pause
  exit /b 1
)

for %%F in (RelayPane.exe) do (
  echo [build.bat] Built: %%~fF
  echo [build.bat] Size: %%~zF bytes  Modified: %%~tF
)
echo [build.bat] Double-click RelayPane.exe to run. For dev builds with console errors, use run.bat.
echo.
exit /b 0
