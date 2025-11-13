# Convert4Share

`Convert4Share` is a simple command-line utility for Windows designed to convert `.mov` and `.heic` files into the more widely compatible `.mp4` and `.jpg` formats.

It is built to be seamlessly integrated with the Windows Shell's "Open with" or "Send to" context menus, allowing for quick and easy file conversions directly from the file explorer.

## Key Features

- **File Conversion**:
  - Converts `.mov` (QuickTime Video) files to `.mp4` (H.264/AAC).
  - Converts `.heic` (High-Efficiency Image Format) files to `.jpg`.
- **Hardware Acceleration**: Automatically detects AMD GPUs on Windows and utilizes the `h264_amf` hardware encoder for faster video conversion. Falls back to the `libx264` software encoder if an AMD GPU is not found.
- **Concurrent Processing**: Boosts performance by processing multiple image conversions in parallel. Video conversions are processed one at a time to ensure stability.
- **Smart Output Path**:
  - If a source file is located within a path containing `Photos`, the converted file is saved to your user's `Downloads` folder.
  - Otherwise, the converted file is saved in the same directory as the original file.
- **Single Instance Execution**: Ensures that only one instance of the application runs at a time. If you select multiple files to convert, they are queued and processed by the single master instance, providing a smooth and predictable user experience.

## Prerequisites

For `Convert4Share` to function correctly, the following third-party software must be installed on your system and accessible via the system's `PATH` environment variable:

- **FFmpeg**: Required for video conversion.
- **ImageMagick**: Required for image conversion.

## Building from Source

To build the application from source, you need to have Go installed. Run the following command in the project's root directory:

```shell
go build -o convert4share.exe
```

This command compiles the application as a GUI program. This is crucial to prevent the console window from flashing when the application is launched from the context menu.

## How to Use

The tool is designed to be used from the command line or, more conveniently, through the Windows File Explorer. You can convert a single file or multiple files at once.

### Command Line

Run the executable with the paths to the files you want to convert as arguments:

```shell
# Convert a single file
convert4share.exe "C:\path\to\your\video.mov"

# Convert multiple files at once
convert4share.exe "C:\path\to\video.mov" "C:\path\to\image.heic" "D:\another\video.mov"
```

### Windows Explorer Integration (Recommended)

The application can be integrated directly into the Windows context menu for `.mov` and `.heic` files, making conversions quick and easy.

**To install the context menu:**

1.  Place the `convert4share.exe` file in a permanent location on your computer (e.g., `C:\Program Files\Convert4Share`).
2.  Open a Command Prompt or PowerShell and navigate to the directory where you placed the executable.
3.  Run the following command:
    ```shell
    convert4share.exe install
    ```
   The application will automatically request administrator privileges (UAC prompt) to add a "Convert with Convert4Share" option to the right-click menu.

   **Note for Windows 11 users:** Due to changes in the context menu, the "Convert with Convert4Share" option may appear in the "Show more options" submenu. You can also access the classic context menu directly by holding `Shift` while right-clicking the file.

**To uninstall the context menu:**

Simply run the `uninstall` command from the same directory:
   ```shell
   convert4share.exe uninstall
   ```

## How It Works

When you run `convert4share.exe` with a file path, it checks if an instance of the program is already running.

- **If it's the first instance**: It starts a master process, initializes worker pools for video and image processing, and begins processing the file. It then stays running for a short period (10 seconds) to listen for more conversion requests.
- **If an instance is already running**: It sends the new file path to the master instance for processing and then immediately exits.

The master instance will automatically shut down after being idle for 10 seconds to free up system resources.

## License

This project is licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)**. The full license text can be found in the `LICENSE.md` file.

Please note that some files may be licensed differently. In such cases, the specific license is specified at the top of the source file. For example, the single-instance logic is based on code from the Wails project and is licensed under the MIT License.
