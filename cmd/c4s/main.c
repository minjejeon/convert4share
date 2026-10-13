#if defined(_WIN32)
#define UNICODE
#define _UNICODE
#include <windows.h>
#include <stdio.h>
#include <stdlib.h>
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

    // Calculate total buffer size for command line:
    // "targetExe" --cli [escaped args...]
    size_t cmdLineLen = wcslen(targetExe) + 32;
    for (int i = 1; i < argc; i++) {
        cmdLineLen += wcslen(argv[i]) * 2 + 8;
    }

    wchar_t *cmdLine = (wchar_t *)malloc(cmdLineLen * sizeof(wchar_t));
    if (!cmdLine) {
        fwprintf(stderr, L"Out of memory\n");
        return 1;
    }

    _snwprintf(cmdLine, cmdLineLen, L"\"%s\" --cli", targetExe);
    for (int i = 1; i < argc; i++) {
        wcscat(cmdLine, L" \"");
        // Escape quotes if needed
        for (const wchar_t *p = argv[i]; *p; p++) {
            if (*p == L'"') {
                wcscat(cmdLine, L"\\\"");
            } else {
                size_t cl = wcslen(cmdLine);
                cmdLine[cl] = *p;
                cmdLine[cl + 1] = L'\0';
            }
        }
        wcscat(cmdLine, L"\"");
    }

    STARTUPINFOW si;
    PROCESS_INFORMATION pi;
    ZeroMemory(&si, sizeof(si));
    si.cb = sizeof(si);
    ZeroMemory(&pi, sizeof(pi));

    if (!CreateProcessW(NULL, cmdLine, NULL, NULL, TRUE, 0, NULL, NULL, &si, &pi)) {
        DWORD err = GetLastError();
        fwprintf(stderr, L"Failed to execute Convert4Share backend (%ls): error %lu\n", targetExe, err);
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
