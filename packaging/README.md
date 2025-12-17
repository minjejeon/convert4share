# MSIX Packaging and Share Target

This directory contains resources for packaging **Convert4Share** as an MSIX application and implementing the Windows Share Target contract.

## Prerequisites

1.  **Windows 10 SDK** (for `MakeAppx.exe` and `SignTool.exe`).
2.  A valid code-signing certificate.
3.  **Go WinRT Bindings**: You must generate the required Go bindings for WinRT APIs.

## 1. Generate WinRT Bindings

The Go backend uses `saltosystems/winrt-go` to interact with Windows Share APIs. Since these bindings depend on your Windows SDK version, they must be generated on your machine.

Using `Taskfile`:
```bash
task gen-winrt
```
Or manually:
```bash
go generate ./internal/winrt
```

## 2. Build with Share Target Support

To enable the Share Target logic, build with the `winrt` tag.

Using `Taskfile`:
```bash
task build-winrt
```
Or manually:
```bash
wails build -tags winrt
```

## 3. Package as MSIX

To package the app manually:

1.  Create a directory (e.g., `MyPackage`).
2.  Copy the built `convert4share.exe` to `MyPackage`.
3.  Create an `Assets` subdirectory in `MyPackage` and add the required logo images:
    -   `StoreLogo.png`
    -   `Square150x150Logo.png`
    -   `Square44x44Logo.png`
4.  Copy `packaging/Package.appxmanifest` to `MyPackage/AppxManifest.xml`.
5.  Run the packaging commands:

```powershell
MakeAppx pack /d "path\to\MyPackage" /p "Convert4share.msix"
SignTool sign /fd SHA256 /a /f "MyCertificate.pfx" /p "password" "Convert4share.msix"
```

## How it Works

When the app is activated via the Share UI:
1.  The app launches.
2.  `share.CheckActivation()` calls `AppInstance.GetActivatedEventArgs`.
3.  If it's a `ShareTarget` activation, it retrieves the files (WinRT Async).
4.  The files are appended to `os.Args` so the standard startup logic processes them.
