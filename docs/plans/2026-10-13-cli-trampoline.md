# CLI Mode and C Trampoline Implementation Plan

> **For Antigravity:** REQUIRED WORKFLOW: Use `.agent/workflows/execute-plan.md` to execute this plan in single-flow mode.

**Goal:** Implement a unified CLI mode for Convert4Share with a minimal C trampoline (`c4s`) so that all features can be used via CLI with zero GUI overhead and negligible binary size increase.

**Architecture:** A lightweight C program (`cmd/c4s/main.c`, ~20KB) operates as the console subsystem trampoline on Windows/Linux, executing `convert4share` with the `--cli` flag and forwarding arguments, stdio, and exit codes. In `cmd/convert4share/main.go`, `--cli` immediately bypasses Wails GUI initialization and invokes `internal/cli`, which executes batch conversions via `internal/batch`.

**Tech Stack:** Go 1.23+, C99 (POSIX/Win32), Wails v3, Taskfile, nfpm, NSIS.

---

### Task 1: C Trampoline Launcher (`cmd/c4s/main.c`)

**Files:**
- Create: `cmd/c4s/main.c`
- Create: `cmd/c4s/Makefile` (or test build script)

**Step 1: Write the C trampoline implementation**

```c
#if defined(_WIN32)
#define UNICODE
#define _UNICODE
#include <windows.h>
#include <stdio.h>
#include <wchar.h>

int wmain(int argc, wchar_t *argv[]) {
    wchar_t exePath[MAX_PATH];
    DWORD len = GetModuleFileNameW(NULL, exePath, MAX_PATH);
    if (len == 0 || len >= MAX_PATH) {
        fwprintf(stderr, L"Failed to resolve executable path\n");
        return 1;
    }

    wchar_t *lastSlash = wcsrchr(exePath, L'\\');
    if (!lastSlash) lastSlash = wcsrchr(exePath, L'/');
    if (!lastSlash) {
        fwprintf(stderr, L"Failed to resolve directory\n");
        return 1;
    }
    *(lastSlash + 1) = L'\0';

    wchar_t targetExe[MAX_PATH];
    _snwprintf(targetExe, MAX_PATH, L"%sconvert4share.exe", exePath);

    // Build command line: "targetExe" --cli argv[1] argv[2] ...
    // Calculate total buffer size
    size_t cmdLineLen = wcslen(targetExe) + 32; // quotes and --cli
    for (int i = 1; i < argc; i++) {
        cmdLineLen += wcslen(argv[i]) * 2 + 4; // space, quotes, escapes
    }

    wchar_t *cmdLine = (wchar_t *)malloc(cmdLineLen * sizeof(wchar_t));
    if (!cmdLine) {
        fwprintf(stderr, L"Out of memory\n");
        return 1;
    }

    _snwprintf(cmdLine, cmdLineLen, L"\"%s\" --cli", targetExe);
    for (int i = 1; i < argc; i++) {
        wcscat(cmdLine, L" \"");
        wcscat(cmdLine, argv[i]);
        wcscat(cmdLine, L"\"");
    }

    STARTUPINFOW si;
    PROCESS_INFORMATION pi;
    ZeroMemory(&si, sizeof(si));
    si.cb = sizeof(si);
    ZeroMemory(&pi, sizeof(pi));

    if (!CreateProcessW(NULL, cmdLine, NULL, NULL, TRUE, 0, NULL, NULL, &si, &pi)) {
        DWORD err = GetLastError();
        fwprintf(stderr, L"Failed to execute Convert4Share backend (error %lu)\n", err);
        free(cmdLine);
        return 1;
    }

    free(cmdLine);
    WaitForSingleObject(pi.hProcess, INFINITE);

    DWORD exitCode = 0;
    GetExitCodeProcess(pi.hProcess, &exitCode);
    CloseHandle(pi.hProcess);
    CloseHandle(pi.hThread);

    return (int)exitCode;
}
#else
#define _GNU_SOURCE
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <limits.h>
#include <libgen.h>

int main(int argc, char *argv[]) {
    char exePath[PATH_MAX];
    ssize_t len = readlink("/proc/self/exe", exePath, sizeof(exePath) - 1);
    if (len == -1) {
        perror("readlink /proc/self/exe");
        return 1;
    }
    exePath[len] = '\0';

    char *dir = dirname(exePath);
    char targetExe[PATH_MAX];
    snprintf(targetExe, sizeof(targetExe), "%s/convert4share", dir);

    char **newArgv = malloc(sizeof(char *) * (argc + 2));
    if (!newArgv) {
        perror("malloc");
        return 1;
    }

    newArgv[0] = targetExe;
    newArgv[1] = "--cli";
    for (int i = 1; i < argc; i++) {
        newArgv[i + 1] = argv[i];
    }
    newArgv[argc + 1] = NULL;

    execv(targetExe, newArgv);
    perror("execv convert4share");
    free(newArgv);
    return 1;
}
#endif
```

**Step 2: Compile and test C trampoline locally**

Run: `gcc -O2 -s -o bin/c4s cmd/c4s/main.c`
Expected: `bin/c4s` is compiled cleanly with size < 30KB.

**Step 3: Commit**

```bash
git add cmd/c4s/main.c
git commit -m "feat(cli): add minimal C trampoline launcher (c4s)"
```

---

### Task 2: Shared Batch Conversion Engine (`internal/batch`)

**Files:**
- Create: `internal/batch/runner.go`
- Create: `internal/batch/runner_test.go`
- Modify: `internal/services/jobs/service.go`

**Step 1: Write failing test for batch runner**

Create `internal/batch/runner_test.go`:
```go
package batch_test

import (
    "context"
    "testing"

    "github.com/minjejeon/convert4share/internal/batch"
    "github.com/minjejeon/convert4share/internal/converter"
)

func testBatchRunner(t *testing.T) {
    runner := batch.NewRunner(&converter.Config{}, batch.Options{
        CollisionOption: "rename",
        AutoLivePhoto:   true,
    })
    if runner == nil {
        t.Fatal("expected runner to not be nil")
    }
}
```

**Step 2: Run test to verify failure**

Run: `go test ./internal/batch/...`
Expected: Compilation failure because package `batch` does not exist yet.

**Step 3: Implement `internal/batch/runner.go`**

Implement `Runner`, `Options`, `JobProgress`, and `RunBatch(ctx context.Context, files []string, progressCb func(JobProgress)) BatchResult`.
Support destination path resolution, collision handling, live photo pairing, semaphores, and error tracking.

**Step 4: Run test to verify pass**

Run: `go test -v ./internal/batch/...`
Expected: PASS

**Step 5: Refactor `internal/services/jobs/service.go` to use `internal/batch`**

Delegate `runBatch` in `Service` to the shared `batch.Runner` logic so GUI and CLI behavior remain 100% consistent.

**Step 6: Run existing tests**

Run: `go test ./...`
Expected: PASS

**Step 7: Commit**

```bash
git add internal/batch internal/services/jobs/service.go
git commit -m "refactor(jobs): extract shared batch runner for CLI and GUI"
```

---

### Task 3: CLI Package & Flag Parsing (`internal/cli`)

**Files:**
- Create: `internal/cli/cli.go`
- Create: `internal/cli/cli_test.go`

**Step 1: Write failing test for CLI flag parsing and help display**

In `internal/cli/cli_test.go`:
```go
package cli_test

import (
    "bytes"
    "testing"

    "github.com/minjejeon/convert4share/internal/cli"
)

func TestParseFlagsHelp(t *testing.T) {
    var stdout bytes.Buffer
    exitCode := cli.RunWithIO([]string{"--help"}, &stdout, &stdout)
    if exitCode != 0 {
        t.Fatalf("expected exit code 0 for --help, got %d", exitCode)
    }
    if !bytes.Contains(stdout.Bytes(), []byte("Usage:")) {
        t.Fatalf("expected usage in output: %s", stdout.String())
    }
}
```

**Step 2: Run test to verify failure**

Run: `go test ./internal/cli/...`
Expected: Compilation failure (package `cli` does not exist).

**Step 3: Implement `internal/cli/cli.go`**

- Flags: `-o / --output`, `-q / --quality`, `--hwaccel`, `--max-size`, `--max-image-size`, `--collision`, `--live-photo`, `-j / --concurrency`, `-v / --verbose`, `-h / --help`, `-V / --version`.
- Config loading from user config directory via `viper`.
- Tool detection via `tools.Service`.
- Batch conversion dispatch with formatted terminal output.
- Signal handling (`os.Interrupt`, `syscall.SIGTERM`).

**Step 4: Run test to verify pass**

Run: `go test -v ./internal/cli/...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/cli
git commit -m "feat(cli): implement CLI flag parser, runner, and terminal output"
```

---

### Task 4: Main Entry Point Interception (`cmd/convert4share/main.go`)

**Files:**
- Modify: `cmd/convert4share/main.go`

**Step 1: Write CLI detection logic and integration in `main.go`**

Inspect `os.Args`:
```go
func hasCLIFlag(args []string) bool {
    for _, arg := range args {
        if arg == "--cli" {
            return true
        }
    }
    return false
}

func filterCLIFlag(args []string) []string {
    var out []string
    for _, arg := range args {
        if arg != "--cli" {
            out = append(out, arg)
        }
    }
    return out
}
```
In `main()`:
```go
if hasCLIFlag(os.Args) {
    os.Exit(cli.Run(filterCLIFlag(os.Args[1:])))
}
```

**Step 2: Build and verify `--cli` execution without GUI**

Run: `go build -o bin/convert4share ./cmd/convert4share && ./bin/convert4share --cli --help`
Expected: Prints CLI help output without launching GUI window.

**Step 3: Commit**

```bash
git add cmd/convert4share/main.go
git commit -m "feat(app): add --cli mode interception to bypass Wails runtime"
```

---

### Task 5: Build System & Packaging (`Taskfile` & `nfpm.yaml`)

**Files:**
- Modify: `build/Taskfile.linux.yml`
- Modify: `build/Taskfile.windows.yml`
- Modify: `build/linux/nfpm/nfpm.yaml`
- Modify: `build/windows/nsis/project.nsi`

**Step 1: Update `build/Taskfile.linux.yml`**

Add task:
```yaml
build:c4s:
  summary: Compiles the minimal C trampoline for Linux
  cmds:
    - '{{.CC | default "gcc"}} -O2 -s -o {{.BIN_DIR}}/c4s ./cmd/c4s/main.c'
```
Add `build:c4s` as a dependency of `linux:build`.

**Step 2: Update `build/Taskfile.windows.yml`**

Add task:
```yaml
build:c4s:
  summary: Compiles the minimal C trampoline for Windows
  cmds:
    - '{{.CC | default "gcc"}} -O2 -s -municode -mconsole -o {{.BIN_DIR}}/c4s.exe ./cmd/c4s/main.c'
```
Add `build:c4s` as a dependency of `windows:build`.

**Step 3: Update `build/linux/nfpm/nfpm.yaml`**

Include `bin/c4s` in `/usr/local/bin/c4s`.

**Step 4: Update `build/windows/nsis/project.nsi`**

Add `File "${ROOT_DIR}\bin\c4s.exe"` in install section.

**Step 5: Verify build**

Run: `task build`
Expected: Both `bin/convert4share` and `bin/c4s` are built successfully.

**Step 6: Commit**

```bash
git add build/
git commit -m "build: integrate c4s trampoline into Taskfile, nfpm, and NSIS installer"
```

---

### Task 6: End-to-End Verification

**Step 1: Run all Go tests**

Run: `go test -v ./...`
Expected: All tests PASS.

**Step 2: End-to-end execution of trampoline `bin/c4s`**

Run: `./bin/c4s --version`
Run: `./bin/c4s --help`
Expected: Properly invokes `convert4share --cli` and exits with code 0.

**Step 3: Test CLI file conversion**

Run with test image/video file to verify end-to-end conversion.

**Step 4: Commit any remaining refinements and update task tracker**

```bash
git add -A
git commit -m "test: complete end-to-end verification for CLI mode and trampoline"
```
