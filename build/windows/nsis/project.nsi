Unicode true

####
## Convert4Share NSIS installer.
##
## Starts at user level (no UAC). On the Install-mode page, the user
## picks one of:
##   - "Current user only" -> stays unelevated, installs into $LOCALAPPDATA
##                            and writes registry under HKCU.
##   - "All users (admin)" -> if not yet elevated, the installer
##                            re-launches itself with `runas` (UAC), then
##                            exits. The elevated child installs into
##                            $PROGRAMFILES64 and writes registry under
##                            HKLM.
####

# Force REQUEST_EXECUTION_LEVEL=user so wails_tools.nsh sets
# RequestExecutionLevel user. The installer always starts unelevated;
# admin elevation only happens if the user explicitly picks
# "All users" on the install-mode page.
!define REQUEST_EXECUTION_LEVEL "user"

# wails_tools.nsh sets project metadata (${INFO_*}, ARCH, macros).
!include "wails_tools.nsh"

!include "MUI.nsh"
!include "nsDialogs.nsh"
!include "FileFunc.nsh"
!include "LogicLib.nsh"
!include "WordFunc.nsh"

# The version information for these two must consist of 4 parts
VIProductVersion "${INFO_PRODUCTVERSION}.0"
VIFileVersion    "${INFO_PRODUCTVERSION}.0"

VIAddVersionKey "CompanyName"     "${INFO_COMPANYNAME}"
VIAddVersionKey "FileDescription" "${INFO_PRODUCTNAME} Installer"
VIAddVersionKey "ProductVersion"  "${INFO_PRODUCTVERSION}"
VIAddVersionKey "FileVersion"     "${INFO_PRODUCTVERSION}"
VIAddVersionKey "LegalCopyright"  "${INFO_COPYRIGHT}"
VIAddVersionKey "ProductName"     "${INFO_PRODUCTNAME}"

# Enable HiDPI support. https://nsis.sourceforge.io/Reference/ManifestDPIAware
ManifestDPIAware true

# Install mode tracking ($InstallScope is "user" or "all").
Var InstallScope
Var RadioUser
Var RadioAll

# Upgrade-detection state, populated in .onInit by scanning HKLM/HKCU
# uninstall registry. When $ExistingScope is non-empty we treat this run
# as an in-place upgrade: scope and INSTDIR are inherited, and the
# InstallScope / Directory pages are skipped.
Var ExistingScope
Var ExistingDir
Var ExistingVer
Var AllUsersFlag
# Set when both HKLM and HKCU uninstall entries are present at .onInit
# time; used to warn the user about parallel installs.
Var BothScopesExist

# Safer file-association macro: only writes the `_backup` slot when the
# current default class differs from ours. The upstream APP_ASSOCIATE in
# wails_tools.nsh clobbers `_backup` on every reinstall, which means
# repeated installs lose the user's original .heic/.mov handler and a
# later uninstall restores a now-deleted class instead of the original.
!macro SAFE_APP_ASSOCIATE EXT FILECLASS DESCRIPTION ICON COMMANDTEXT COMMAND
  ReadRegStr $R0 SHELL_CONTEXT "Software\Classes\.${EXT}" ""
  ${If} $R0 != "${FILECLASS}"
    WriteRegStr SHELL_CONTEXT "Software\Classes\.${EXT}" "${FILECLASS}_backup" "$R0"
  ${EndIf}

  WriteRegStr SHELL_CONTEXT "Software\Classes\.${EXT}" "" "${FILECLASS}"

  WriteRegStr SHELL_CONTEXT "Software\Classes\${FILECLASS}" "" `${DESCRIPTION}`
  WriteRegStr SHELL_CONTEXT "Software\Classes\${FILECLASS}\DefaultIcon" "" `${ICON}`
  WriteRegStr SHELL_CONTEXT "Software\Classes\${FILECLASS}\shell" "" "open"
  WriteRegStr SHELL_CONTEXT "Software\Classes\${FILECLASS}\shell\open" "" `${COMMANDTEXT}`
  WriteRegStr SHELL_CONTEXT "Software\Classes\${FILECLASS}\shell\open\command" "" `${COMMAND}`
!macroend

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
!define MUI_FINISHPAGE_NOAUTOCLOSE
!define MUI_ABORTWARNING

# Skip the Welcome page on the UAC-elevated child and on detected
# upgrades — the parent already showed it (fresh-install all-users
# path), and seasoned users don't need it again.
!define MUI_PAGE_CUSTOMFUNCTION_PRE WelcomePagePre
!insertmacro MUI_PAGE_WELCOME
Page custom InstallScopePageCreate InstallScopePageLeave

# Skip the directory page on detected upgrades — the existing install
# directory has already been adopted in .onInit, and letting the user
# pick a fresh path on upgrade silently orphans the previous copy.
!define MUI_PAGE_CUSTOMFUNCTION_PRE DirectoryPagePre
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "English"

Name "${INFO_PRODUCTNAME}"
# OutFile is resolved relative to this file (build/windows/nsis/).
OutFile "..\..\..\bin\${INFO_PROJECTNAME}-${ARCH}-installer.exe"
ShowInstDetails show

# Read CLI flags from the parent (e.g. self-relaunch passes /AllUsers).
Function .onInit
    StrCpy $InstallScope "user"
    StrCpy $INSTDIR "$LOCALAPPDATA\Programs\${INFO_PRODUCTNAME}"
    StrCpy $AllUsersFlag "0"
    StrCpy $ExistingScope ""
    StrCpy $ExistingDir ""
    StrCpy $ExistingVer ""
    StrCpy $BothScopesExist "0"

    ${GetParameters} $R0
    ClearErrors
    ${GetOptions} $R0 "/AllUsers" $R1
    ${IfNot} ${Errors}
        StrCpy $AllUsersFlag "1"
        StrCpy $InstallScope "all"
        StrCpy $INSTDIR "$PROGRAMFILES64\${INFO_PRODUCTNAME}"
    ${EndIf}

    # Detect an existing installation. HKLM takes precedence so an
    # all-users install is always preferred over a per-user one when
    # both exist — re-running this installer should adopt the elevated
    # copy rather than silently spawning a parallel per-user install.
    SetRegView 64
    ReadRegStr $R2 HKLM "${UNINST_KEY}" "InstallLocation"
    ${If} $R2 != ""
        StrCpy $ExistingScope "all"
        StrCpy $ExistingDir $R2
        ReadRegStr $ExistingVer HKLM "${UNINST_KEY}" "DisplayVersion"
        # Note: a parallel HKCU install may exist alongside the HKLM
        # one. Record it so we can warn the user — the upcoming
        # registry writes will only refresh the HKLM tree, leaving the
        # per-user copy untouched and confusing.
        ReadRegStr $R3 HKCU "${UNINST_KEY}" "InstallLocation"
        ${If} $R3 != ""
            StrCpy $BothScopesExist "1"
        ${EndIf}
    ${Else}
        ReadRegStr $R2 HKCU "${UNINST_KEY}" "InstallLocation"
        ${If} $R2 != ""
            StrCpy $ExistingScope "user"
            StrCpy $ExistingDir $R2
            ReadRegStr $ExistingVer HKCU "${UNINST_KEY}" "DisplayVersion"
        ${EndIf}
    ${EndIf}

    # Warn once (in the unelevated parent) about parallel installs so
    # the user can bail out and uninstall the orphaned scope before
    # continuing.
    ${If} $AllUsersFlag == "0"
    ${AndIf} $BothScopesExist == "1"
        MessageBox MB_YESNO|MB_ICONEXCLAMATION \
            "${INFO_PRODUCTNAME} is installed for both All Users and the current user on this PC.$\nThis installer will upgrade the All Users install only. The per-user install must be removed separately.$\n$\nContinue anyway?" \
            IDYES +2
        Quit
    ${EndIf}

    # Adopt the existing scope/dir unless /AllUsers forces all-users
    # mode (e.g. the elevated child of a fresh install).
    ${If} $AllUsersFlag == "0"
    ${AndIf} $ExistingScope != ""
        StrCpy $InstallScope $ExistingScope
        StrCpy $INSTDIR $ExistingDir
    ${EndIf}

    ${If} $InstallScope == "all"
        SetShellVarContext all
    ${Else}
        SetShellVarContext current
    ${EndIf}

    # Downgrade warning. Only show in the un-elevated parent so we don't
    # double-prompt across the UAC re-launch.
    ${If} $AllUsersFlag == "0"
    ${AndIf} $ExistingVer != ""
        ${VersionCompare} "${INFO_PRODUCTVERSION}" "$ExistingVer" $R0
        ${If} $R0 == "2"
            MessageBox MB_YESNO|MB_ICONEXCLAMATION \
                "A newer version ($ExistingVer) is already installed.$\nDowngrade to version ${INFO_PRODUCTVERSION}?" \
                IDYES +2
            Quit
        ${EndIf}
    ${EndIf}

    # If we resolved to an all-users install but are still running
    # un-elevated (either picked-up from an existing HKLM install, or
    # never got the InstallScope page yet), re-launch ourselves with
    # `runas` to elevate. The child re-enters .onInit with /AllUsers,
    # skips this branch, and proceeds.
    ${If} $InstallScope == "all"
    ${AndIf} $AllUsersFlag == "0"
        ExecShell "runas" "$EXEPATH" "/AllUsers"
        SetErrorLevel 0
        Quit
    ${EndIf}

    # Refuse to upgrade in place if the app is currently running, since
    # `wails.files` would otherwise fail with a sharing violation and
    # leave the user staring at a generic NSIS error. Probe via an
    # atomic rename — works for both per-user and all-users layouts and
    # needs no extra plugins or process enumeration.
    ${If} $ExistingScope != ""
    ${AndIf} ${FileExists} "$INSTDIR\${PRODUCT_EXECUTABLE}"
        ClearErrors
        Rename "$INSTDIR\${PRODUCT_EXECUTABLE}" "$INSTDIR\${PRODUCT_EXECUTABLE}.upgradeprobe"
        ${If} ${Errors}
            MessageBox MB_OK|MB_ICONEXCLAMATION \
                "${INFO_PRODUCTNAME} is currently running.$\nClose the app and run setup again."
            SetErrorLevel 1
            Quit
        ${Else}
            Rename "$INSTDIR\${PRODUCT_EXECUTABLE}.upgradeprobe" "$INSTDIR\${PRODUCT_EXECUTABLE}"
        ${EndIf}
    ${EndIf}

    !insertmacro wails.checkArchitecture
FunctionEnd

Function un.onInit
    SetShellVarContext current
    # If the uninstaller is running with admin token (called from the
    # Add/Remove Programs entry that was written under HKLM), restore
    # all-users context so we delete the right registry tree.
    System::Call "advapi32::GetUserName(t .r0, *i ${NSIS_MAX_STRLEN}) i.r1"
    ${If} $0 == "SYSTEM"
        SetShellVarContext all
    ${Else}
        # Best-effort: detect AllUsers install by checking HKLM uninstall key.
        SetRegView 64
        ReadRegStr $0 HKLM "${UNINST_KEY}" "InstallLocation"
        ${If} $0 != ""
            SetShellVarContext all
        ${EndIf}
    ${EndIf}
FunctionEnd

# ---- Install-mode custom page ----------------------------------------------
Function InstallScopePageCreate
    # If we were re-launched with /AllUsers we already know the scope; skip.
    # Also skip on detected upgrades — the previous scope wins, and asking
    # again would let users silently spawn a parallel install in the other
    # scope.
    ${If} $InstallScope == "all"
    ${OrIf} $ExistingScope != ""
        Abort
    ${EndIf}

    !insertmacro MUI_HEADER_TEXT "Choose Install Mode" "Pick who will see this installation."

    nsDialogs::Create 1018
    Pop $0
    ${If} $0 == error
        Abort
    ${EndIf}

    ${NSD_CreateLabel} 0 0 100% 24u "Install ${INFO_PRODUCTNAME} for the current user only, or for all users on this PC. Installing for all users adds the .mov and .heic context-menu entries for everyone and requires administrator approval."
    Pop $0

    ${NSD_CreateRadioButton} 0 32u 100% 12u "&Current user only (no admin needed)"
    Pop $RadioUser
    ${NSD_AddStyle} $RadioUser ${WS_GROUP}
    ${NSD_OnClick} $RadioUser InstallScopeSelectUser

    ${NSD_CreateRadioButton} 0 48u 100% 12u "&All users (requires administrator approval)"
    Pop $RadioAll
    ${NSD_OnClick} $RadioAll InstallScopeSelectAll

    # Default = current user (preselect).
    ${NSD_Check} $RadioUser

    nsDialogs::Show
FunctionEnd

Function InstallScopeSelectUser
    StrCpy $InstallScope "user"
FunctionEnd

Function InstallScopeSelectAll
    StrCpy $InstallScope "all"
FunctionEnd

Function InstallScopePageLeave
    ${If} $InstallScope == "all"
        # Re-launch ourselves with UAC so the rest of the install runs
        # with administrative privileges. The child uses /AllUsers to
        # skip this page and start in all-users mode.
        ExecShell "runas" "$EXEPATH" "/AllUsers"
        SetErrorLevel 0
        Quit
    ${Else}
        SetShellVarContext current
        StrCpy $INSTDIR "$LOCALAPPDATA\Programs\${INFO_PRODUCTNAME}"
    ${EndIf}
FunctionEnd

Function DirectoryPagePre
    # On detected upgrades, hide the directory page entirely so users
    # can't redirect the upgrade into a fresh folder and orphan the old
    # install + its uninstall.exe entry.
    ${If} $ExistingScope != ""
        Abort
    ${EndIf}
FunctionEnd

Function WelcomePagePre
    # Skip the welcome page on the elevated /AllUsers child (the parent
    # already greeted the user) and on detected upgrades (they know
    # what this is — straight to the install action).
    ${If} $AllUsersFlag == "1"
    ${OrIf} $ExistingScope != ""
        Abort
    ${EndIf}
FunctionEnd
# -----------------------------------------------------------------------------

Section
    # SetShellVarContext is already set by .onInit (current or all). All
    # registry writes below use SHELL_CONTEXT which routes to HKLM or HKCU
    # automatically.

    !insertmacro wails.webview2runtime

    SetOutPath $INSTDIR

    !insertmacro wails.files

    CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    CreateShortCut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"

    # Clean up legacy v2-era registry keys written by the old in-app
    # context-menu installer (windows/registry_windows.go, removed in
    # 69f85e0). Idempotent — DeleteRegKey is a no-op when the key is
    # absent. HKCU is always cleaned; HKLM only when running elevated.
    SetRegView 64
    DeleteRegKey HKCU "Software\Classes\Convert4Share.File"
    DeleteRegKey HKCU "Software\Classes\*\shell\Convert4Share"
    ${If} $InstallScope == "all"
        DeleteRegKey HKLM "Software\Classes\Convert4Share.File"
        DeleteRegKey HKLM "Software\Classes\*\shell\Convert4Share"
    ${EndIf}

    # File associations — use SAFE_APP_ASSOCIATE (defined above) instead
    # of wails.associateFiles, which calls APP_ASSOCIATE and clobbers
    # the `_backup` slot on every reinstall. The macro lives in this
    # file so it survives wails_tools.nsh regeneration.
    !insertmacro SAFE_APP_ASSOCIATE "mov" "QuickTime Video" "QuickTime Video (handled by Convert4Share)" "$INSTDIR\icon.ico" "Open with ${INFO_PRODUCTNAME}" "$INSTDIR\${PRODUCT_EXECUTABLE} $\"%1$\""
    !insertmacro SAFE_APP_ASSOCIATE "heic" "HEIC Image" "High Efficiency Image (handled by Convert4Share)" "$INSTDIR\icon.ico" "Open with ${INFO_PRODUCTNAME}" "$INSTDIR\${PRODUCT_EXECUTABLE} $\"%1$\""
    File "..\icon.ico"

    # Uninstaller registration via SHELL_CONTEXT so per-user installs land
    # in HKCU and all-users installs land in HKLM. The wails_tools.nsh
    # `wails.writeUninstaller` helper hardcodes HKLM and would silently
    # fail under per-user.
    WriteUninstaller "$INSTDIR\uninstall.exe"

    SetRegView 64
    WriteRegStr   SHELL_CONTEXT "${UNINST_KEY}" "Publisher"            "${INFO_COMPANYNAME}"
    WriteRegStr   SHELL_CONTEXT "${UNINST_KEY}" "DisplayName"          "${INFO_PRODUCTNAME}"
    WriteRegStr   SHELL_CONTEXT "${UNINST_KEY}" "DisplayVersion"       "${INFO_PRODUCTVERSION}"
    WriteRegStr   SHELL_CONTEXT "${UNINST_KEY}" "DisplayIcon"          "$INSTDIR\${PRODUCT_EXECUTABLE}"
    WriteRegStr   SHELL_CONTEXT "${UNINST_KEY}" "UninstallString"      "$\"$INSTDIR\uninstall.exe$\""
    WriteRegStr   SHELL_CONTEXT "${UNINST_KEY}" "QuietUninstallString" "$\"$INSTDIR\uninstall.exe$\" /S"
    WriteRegStr   SHELL_CONTEXT "${UNINST_KEY}" "InstallLocation"      "$INSTDIR"
    WriteRegStr   SHELL_CONTEXT "${UNINST_KEY}" "InstallScope"         "$InstallScope"

    ${GetSize} "$INSTDIR" "/S=0K" $0 $1 $2
    IntFmt $0 "0x%08X" $0
    WriteRegDWORD SHELL_CONTEXT "${UNINST_KEY}" "EstimatedSize"        "$0"
SectionEnd

Section "uninstall"
    # un.onInit set the shell context based on which scope was used.

    RMDir /r "$AppData\${PRODUCT_EXECUTABLE}"

    RMDir /r $INSTDIR

    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"

    !insertmacro wails.unassociateFiles

    Delete "$INSTDIR\uninstall.exe"

    SetRegView 64
    DeleteRegKey SHELL_CONTEXT "${UNINST_KEY}"
SectionEnd
