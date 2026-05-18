Unicode true

####
## Convert4Share NSIS installer.
##
## Supports both per-user (no UAC) and all-users (admin via UAC) installs.
## A MULTIUSER_PAGE_INSTALLMODE page lets the user pick mid-install; if
## "All users" is chosen and the current process is not elevated, NSIS
## relaunches itself with the admin token via Windows UAC.
####

# Force REQUEST_EXECUTION_LEVEL=user so wails_tools.nsh sets
# RequestExecutionLevel user. MultiUser then either runs as-is (per-user)
# or re-launches the installer with UAC elevation (all-users).
!define REQUEST_EXECUTION_LEVEL "user"

# wails_tools.nsh sets project metadata (${INFO_*}, ARCH, macros).
!include "wails_tools.nsh"

# ---- MultiUser configuration ------------------------------------------------
!define MULTIUSER_EXECUTIONLEVEL Highest
!define MULTIUSER_MUI
!define MULTIUSER_INSTALLMODE_COMMANDLINE
!define MULTIUSER_INSTALLMODE_INSTDIR "${INFO_PRODUCTNAME}"
!define MULTIUSER_USE_PROGRAMFILES64
# Remember the last chosen mode and install dir so an upgrade run defaults
# to whatever was used previously.
!define MULTIUSER_INSTALLMODE_INSTDIR_REGISTRY_KEY "Software\${UNINST_KEY_NAME}"
!define MULTIUSER_INSTALLMODE_INSTDIR_REGISTRY_VALUENAME "InstallLocation"
!define MULTIUSER_INSTALLMODE_DEFAULT_REGISTRY_KEY "Software\${UNINST_KEY_NAME}"
!define MULTIUSER_INSTALLMODE_DEFAULT_REGISTRY_VALUENAME "InstallMode"
!include "MultiUser.nsh"
# -----------------------------------------------------------------------------

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

!include "MUI.nsh"

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
!define MUI_FINISHPAGE_NOAUTOCLOSE
!define MUI_ABORTWARNING

!insertmacro MUI_PAGE_WELCOME
!insertmacro MULTIUSER_PAGE_INSTALLMODE
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "English"

Name "${INFO_PRODUCTNAME}"
# OutFile is resolved relative to this file (build/windows/nsis/).
OutFile "..\..\..\bin\${INFO_PROJECTNAME}-${ARCH}-installer.exe"
ShowInstDetails show

Function .onInit
   !insertmacro MULTIUSER_INIT
   !insertmacro wails.checkArchitecture
FunctionEnd

Function un.onInit
   !insertmacro MULTIUSER_UNINIT
FunctionEnd

Section
    # MultiUser already called SetShellVarContext in .onInit based on
    # $MultiUser.InstallMode, so SHELL_CONTEXT routes to HKLM (AllUsers)
    # or HKCU (CurrentUser) automatically. Do NOT call wails.setShellContext.

    !insertmacro wails.webview2runtime

    SetOutPath $INSTDIR

    !insertmacro wails.files

    CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    CreateShortCut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"

    !insertmacro wails.associateFiles

    # Custom uninstaller registration: use SHELL_CONTEXT (HKLM or HKCU) so
    # both AllUsers and CurrentUser installs leave a working Add/Remove
    # Programs entry. The wails_tools.nsh `wails.writeUninstaller` macro
    # hard-codes HKLM and would fail for per-user installs.
    WriteUninstaller "$INSTDIR\uninstall.exe"

    SetRegView 64
    WriteRegStr   SHELL_CONTEXT "${UNINST_KEY}" "Publisher"         "${INFO_COMPANYNAME}"
    WriteRegStr   SHELL_CONTEXT "${UNINST_KEY}" "DisplayName"       "${INFO_PRODUCTNAME}"
    WriteRegStr   SHELL_CONTEXT "${UNINST_KEY}" "DisplayVersion"    "${INFO_PRODUCTVERSION}"
    WriteRegStr   SHELL_CONTEXT "${UNINST_KEY}" "DisplayIcon"       "$INSTDIR\${PRODUCT_EXECUTABLE}"
    WriteRegStr   SHELL_CONTEXT "${UNINST_KEY}" "UninstallString"   "$\"$INSTDIR\uninstall.exe$\""
    WriteRegStr   SHELL_CONTEXT "${UNINST_KEY}" "QuietUninstallString" "$\"$INSTDIR\uninstall.exe$\" /S"
    WriteRegStr   SHELL_CONTEXT "${UNINST_KEY}" "InstallLocation"   "$INSTDIR"
    WriteRegStr   SHELL_CONTEXT "${UNINST_KEY}" "InstallMode"       "$MultiUser.InstallMode"

    ${GetSize} "$INSTDIR" "/S=0K" $0 $1 $2
    IntFmt $0 "0x%08X" $0
    WriteRegDWORD SHELL_CONTEXT "${UNINST_KEY}" "EstimatedSize"     "$0"
SectionEnd

Section "uninstall"
    # MultiUser already set context in un.onInit based on the install mode
    # recorded in the registry.

    RMDir /r "$AppData\${PRODUCT_EXECUTABLE}"

    RMDir /r $INSTDIR

    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"

    !insertmacro wails.unassociateFiles

    Delete "$INSTDIR\uninstall.exe"

    SetRegView 64
    DeleteRegKey SHELL_CONTEXT "${UNINST_KEY}"
SectionEnd
