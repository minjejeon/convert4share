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

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
!define MUI_FINISHPAGE_NOAUTOCLOSE
!define MUI_ABORTWARNING

!insertmacro MUI_PAGE_WELCOME
Page custom InstallScopePageCreate InstallScopePageLeave
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

    ${GetParameters} $R0
    ClearErrors
    ${GetOptions} $R0 "/AllUsers" $R1
    ${IfNot} ${Errors}
        StrCpy $InstallScope "all"
        StrCpy $INSTDIR "$PROGRAMFILES64\${INFO_PRODUCTNAME}"
        SetShellVarContext all
    ${Else}
        SetShellVarContext current
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
    ${If} $InstallScope == "all"
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

    !insertmacro wails.associateFiles

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
