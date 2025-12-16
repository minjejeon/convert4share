# Convert4Share

`Convert4Share` is a Windows desktop application designed to convert `.mov` and `.heic` files into the more widely compatible `.mp4` and `.jpg` formats.

It features a modern GUI and is built to be seamlessly integrated with the Windows Shell's "Open with" or "Send to" context menus, allowing for quick and easy file conversions directly from the file explorer.

## Key Features

- **File Conversion**:
  - Converts `.mov` (QuickTime Video) files to `.mp4` (H.264/AAC).
  - Converts `.heic` (High-Efficiency Image Format) files to `.jpg`.
- **Hardware Acceleration**: Automatically detects AMD/NVIDIA GPUs on Windows (during installation) and utilizes hardware encoders (`h264_amf`, `h264_nvenc`) for faster video conversion.
- **Concurrent Processing**: Boosts performance by processing multiple image conversions in parallel. Video conversions are processed one at a time to ensure stability.
- **Smart Output Path**:
  - Configurable "exclude patterns" allow you to divert output to a specific directory (e.g., `Pictures`) if the source path contains certain keywords (e.g., `Cloud`).
  - Otherwise, the converted file is saved in the same directory as the original file.
- **Single Instance Execution**: Ensures that only one instance of the application runs at a time. If you select multiple files to convert, they are queued and processed by the single master instance.

## Prerequisites

For `Convert4Share` to function correctly, the following tools are required:

- **FFmpeg**: Required for video conversion.
- **ImageMagick**: Required for image conversion.

The application allows you to configure the paths to these binaries in the Settings if they are not in your system `PATH`.

## Building from Source

To build the application from source, you need **Go** and **Node.js** (with **npm**) installed.

1.  Install frontend dependencies:
    ```shell
    cd frontend
    npm install
    cd ..
    ```

2.  Build the application using Wails:
    ```shell
    wails build
    ```

## How to Use

The tool is designed to be used from the command line or, more conveniently, through the Windows File Explorer.

### Command Line

Run the executable with the paths to the files you want to convert as arguments:

```shell
# Convert a single file
convert4share.exe "C:\path\to\your\video.mov"
```

### Windows Explorer Integration (Recommended)

The application can be integrated directly into the Windows context menu for `.mov` and `.heic` files.

**To install the context menu:**

1.  Run `convert4share.exe`.
2.  Go to the **Settings** page.
3.  Under the "Windows Integration" section, click the **Install** button.

Alternatively, via CLI:
```shell
convert4share.exe install
```

**To uninstall the context menu:**

You can uninstall it from the **Settings** page in the application, or by running `convert4share.exe uninstall`.

## License

This project is licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)**. The full license text can be found in the `LICENSE` file.
