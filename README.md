# Reed Torrent Client
<img src="icon.svg" alt="Reed Logo" width="300"/>

A GUI torrent client built with Go, using the [anacrolix/torrent](https://github.com/anacrolix/torrent) library and [Fyne](https://fyne.io/) for the user interface.

## Quick Start

### Windows
<details>
1. Install [Go](https://golang.org/dl/) (version 1.18 or later)
2. Install [GCC](https://jmeubank.github.io/tdm-gcc/) for Windows
3. Install [MSYS2](https://www.msys2.org/)
4. Open MSYS2 terminal and run:
   ```
   pacman -S mingw-w64-x86_64-gcc mingw-w64-x86_64-pkg-config
   ```
5. Add MSYS2 bin to your PATH: `C:\msys64\mingw64\bin`
6. Open Command Prompt and run:
   ```
   git clone https://github.com/your-username/reed.git
   cd reed
   go mod tidy
   go get -u github.com/go-gl/gl/v3.2-core/gl
   go build -o reed.exe
   .\reed.exe
   ```
</details>

### macOS
<summary>
1. Install [Go](https://golang.org/dl/) and Xcode Command Line Tools
2. Run:
   ```
   git clone https://github.com/your-username/reed.git
   cd reed
   go mod tidy
   go build
   ./reed
   ```
</summary>


### Linux
<summary>
1. Install Go and required packages:
   ```
   sudo apt-get install gcc libgl1-mesa-dev xorg-dev
   ```
2. Run:
   ```
   git clone https://github.com/your-username/reed.git
   cd reed
   go mod tidy
   go build
   ./reed
   ```
</summary>

## Features

- Add torrents via magnet links
- Open torrent files from your computer
- View download progress
- Remove torrents
- Automatically saves files to your Downloads folder

## Prerequisites

Before building and running the application, you need to install the following dependencies:

### Windows

1. Install [Go](https://golang.org/dl/) (version 1.18 or later)
2. Install [GCC](https://jmeubank.github.io/tdm-gcc/) for CGo support
3. Install [MSYS2](https://www.msys2.org/) and run the following commands in the MSYS2 terminal:
   ```
   pacman -S mingw-w64-x86_64-gcc
   pacman -S mingw-w64-x86_64-pkg-config
   ```
4. Add the MSYS2 bin directory to your PATH (typically `C:\msys64\mingw64\bin`)

### macOS

1. Install [Go](https://golang.org/dl/) (version 1.18 or later)
2. Install [Xcode Command Line Tools](https://developer.apple.com/xcode/resources/)

### Linux

1. Install [Go](https://golang.org/dl/) (version 1.18 or later)
2. Install the required packages:
   ```
   sudo apt-get install gcc libgl1-mesa-dev xorg-dev
   ```

## Building and Running

1. Clone the repository:
   ```
   git clone https://github.com/your-username/reed.git
   cd reed
   ```

2. Install dependencies:
   ```
   go mod tidy
   ```

3. Build the application:
   ```
   # On Linux/macOS
   go build

   # On Windows
   go build -o reed.exe
   ```

4. Run the application:
   ```
   # On Linux/macOS
   ./reed

   # On Windows
   .\reed.exe
   ```

### Troubleshooting Build Issues

If you encounter build errors related to missing packages or build constraints, ensure that:

1. All prerequisites are installed correctly (Go, GCC, MSYS2 on Windows)
2. Your PATH environment variable includes the MSYS2 bin directory
3. You've installed all required packages with MSYS2

For Windows users specifically:
1. Make sure you've run the MSYS2 commands in the Prerequisites section
2. Verify your PATH includes `C:\msys64\mingw64\bin` (or your MSYS2 installation path)
3. You may need to restart your terminal or computer after updating PATH

If you see errors related to `github.com/go-gl/gl`, try running:
```
go get -u github.com/go-gl/gl/v3.2-core/gl
```

If you encounter linking errors with pthread (multiple definition of pthread functions), this is a known issue with CGo on Windows. The project includes a `gl_windows.go` file that adds a linker flag to allow multiple definitions and resolve this issue. Make sure this file is present in your project directory before building.

## Usage

1. Launch the application
2. Enter a magnet link in the text field and click "Add Torrent" or click "Open File" to select a .torrent file
3. The torrent will appear in the list and start downloading automatically
4. To remove a torrent, select it from the list and click "Remove"

## Development

The application is built using:
- [Go](https://golang.org/) as the programming language
- [anacrolix/torrent](https://github.com/anacrolix/torrent) for torrent functionality
- [Fyne](https://fyne.io/) for the GUI

## License

This project is licensed under the MIT License - see the LICENSE file for details.
