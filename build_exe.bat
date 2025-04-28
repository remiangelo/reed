@echo off
echo Building Reed Torrent Client executable...

REM Install Fyne CLI if not already installed
go install fyne.io/fyne/v2/cmd/fyne@latest

REM Check if icon file exists
if exist icon.png (
    echo Using custom icon...
    REM Package the application as a Windows executable with custom icon
    fyne package -os windows -icon icon.png
) else (
    echo No custom icon found, using default icon...
    REM Package the application as a Windows executable with default icon
    fyne package -os windows
)

echo Build complete! You can now run reed.exe by double-clicking it.
