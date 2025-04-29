@echo off
echo Building Reed Torrent Client...

rem Check if Go is installed
where go >nul 2>nul
if %ERRORLEVEL% neq 0 (
    echo Error: Go is not installed or not in your PATH.
    echo Please install Go from https://golang.org/dl/
    pause
    exit /b 1
)

rem Check for required tools
echo Checking for required tools...
go version

rem Install Fyne CLI if needed
echo Checking for Fyne CLI...
go install fyne.io/fyne/v2/cmd/fyne@latest

rem Build the executable
echo Building executable...
go mod tidy
go get -u github.com/go-gl/gl/v3.2-core/gl
go build -ldflags="-H=windowsgui" -o reed.exe

if %ERRORLEVEL% neq 0 (
    echo Build failed. Please check error messages above.
    pause
    exit /b 1
)

rem Check if icon file exists and package with Fyne if needed
if exist icon.png (
    echo Using custom icon for packaging...
    fyne package -os windows -icon icon.png
)

echo Build complete! You can now run reed.exe by double-clicking it.
pause
