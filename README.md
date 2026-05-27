![Convert4Share Logo](logo.png)

# Convert4Share

`Convert4Share` is a Windows desktop application designed to convert `.mov` and `.heic` files into the more widely compatible `.mp4` and `.jpg` formats.

It features a modern GUI and is built to be seamlessly integrated with the Windows Shell's "Open with" or "Send to" context menus, allowing for quick and easy file conversions directly from the file explorer.

## Key Features

- **File Conversion**:
  - Converts `.mov` (QuickTime Video) files to `.mp4` (H.264/AAC).
  - Converts `.heic` (High-Efficiency Image Format) files to `.jpg`.
  - **HEIC Max Resolution**: Configurable limit for the longest side of converted images (default: 2560px).
  - **Pass-through (Copy Only)**: Optionally skip conversion and copy files directly for specific extensions (e.g., `.jpg`, `.mp4`).
- **Live Photo Detection**:
  - Automatically detects and skips the `.mov` component of Apple Live Photos if the corresponding `.heic` file is in the same batch.
- **Drag & Drop Interface**:
  - Simply drag files onto the application window to add them to the conversion queue.
- **Hardware Acceleration**:
  - Automatically detects AMD/NVIDIA GPUs on Windows (during installation) and utilizes hardware encoders (`h264_amf`, `h264_nvenc`) for faster video conversion.
- **Quality Presets**:
  - Supports 'High', 'Medium', and 'Low' quality presets for video conversion, dynamically adjusting bitrates (5Mbps, 2.5Mbps, 1Mbps) and hardware flags.
- **Concurrent Processing**:
  - Boosts performance by processing multiple image conversions in parallel (configurable limit). Video conversions are processed one at a time to ensure stability.
- **Smart Output Path**:
  - Configurable "exclude patterns" allow you to divert output to a specific directory (e.g., `Pictures`) if the source path contains certain keywords (e.g., `Cloud`).
  - Otherwise, the converted file is saved in the same directory as the original file.
- **Theme Support**:
  - Fully supports Light and Dark modes (defaults to Dark), matching your system preference or manual toggle.
- **Single Instance Execution**:
  - Ensures that only one instance of the application runs at a time. If you select multiple files to convert, they are queued and processed by the single master instance.

## Prerequisites

For `Convert4Share` to function correctly, the following tools are required:

- **FFmpeg**: Required for video conversion.
- **ImageMagick**: Required for image conversion.

The application automatically attempts to detect these binaries in your system `PATH`, as well as standard `WinGet` installation locations. You can also manually configure the paths in the Settings if they are not detected.

## Building from Source

You need **Go** (1.25+), **Node.js** with **npm**, and **go-task** installed. The Wails v3 CLI is fetched on demand by Taskfile targets.

```shell
# Development build (frontend + Go binary -> bin/convert4share.exe)
task build

# Production build (-tags production, stripped, -H windowsgui)
task build PRODUCTION=true

# Run the built executable
task run

# Live dev mode (Vite + wails3 dev)
task dev

# Regenerate Wails v3 frontend bindings after changing internal/services/* method signatures
task generate:bindings

# Regenerate build assets (manifest, .syso, NSIS templates) after editing build/config.yml
task common:update:build-assets
```

The Go entry point is `cmd/convert4share/main.go` and embeds `cmd/convert4share/dist/` (populated by `task common:build:frontend` from `frontend/dist`).

### Packaging (NSIS installer)

`task package` builds a Windows installer. It requires `makensis` on `PATH` (NSIS 3+). The installer registers Convert4Share as the handler for `.mov` and `.heic` (declared in `build/config.yml`) so no separate install step is needed at runtime.

## How to Use

### Windows Explorer Integration (Recommended)

The NSIS installer (`task package`) registers Convert4Share as a handler for `.mov` and `.heic` files at install time. After installation, right-click a `.mov` or `.heic` in Explorer and choose **Open with → Convert4Share** to launch a conversion. No in-app install button or admin CLI is needed — the file associations are declared in `build/config.yml`.

If you prefer not to use the installer, you can still launch the app and drop files into the window (see Drag & Drop) or invoke it from the command line.

### Command Line

Pass file paths as arguments. The running instance (if any) receives them via the Wails v3 single-instance lock and queues them automatically:

```shell
convert4share.exe "C:\path\to\your\video.mov" "C:\path\to\your\photo.heic"
```

### Drag & Drop

Drag files onto the application window; the drop target is the main DropZone (highlighted via Wails v3's `.file-drop-target-active` class).

## License

This project is licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)**. The full license text can be found in the `LICENSE` file.
