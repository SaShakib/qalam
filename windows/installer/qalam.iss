; Qalam-Setup-<arch>.exe: a normal Windows installer for Qalam (Inno Setup 6).
;
;   iscc /DAppVersion=1.3.0 /DArch=x64   /DExe=..\..\dist\Qalam-Windows-x64.exe   qalam.iss
;   iscc /DAppVersion=1.3.0 /DArch=arm64 /DExe=..\..\dist\Qalam-Windows-arm64.exe qalam.iss
;
; It installs for the current user only (no admin rights), into the same folder the app's own
; "Install for this user?" uses (%LOCALAPPDATA%\Programs\Qalam), so the in-app updater keeps working.
; Silent install for IT / testing: Qalam-Setup-x64.exe /VERYSILENT /SUPPRESSMSGBOXES /NORESTART

#ifndef AppVersion
  #define AppVersion "0.0.0"
#endif
#ifndef Arch
  #define Arch "x64"
#endif
#ifndef Exe
  #define Exe "..\..\dist\Qalam-Windows-" + Arch + ".exe"
#endif
#ifndef OutDir
  #define OutDir "..\..\dist"
#endif

[Setup]
; Keep this AppId: the app updates "DisplayVersion" under ...\Uninstall\{AppId}_is1 (install_windows.go).
AppId={{6F1C2E7A-6A51-4C1B-9E0B-6D5A3C1B7E41}
AppName=Qalam
AppVersion={#AppVersion}
AppVerName=Qalam {#AppVersion}
AppPublisher=Qalam
AppPublisherURL=https://github.com/SaShakib/qalam
AppSupportURL=https://github.com/SaShakib/qalam/issues
AppUpdatesURL=https://github.com/SaShakib/qalam/releases
AppComments=Phonetic Arabic keyboard with full harakat and Qur'anic marks
VersionInfoVersion={#AppVersion}
PrivilegesRequired=lowest
DefaultDirName={userpf}\Qalam
DisableDirPage=yes
DisableProgramGroupPage=yes
DisableReadyPage=yes
UninstallDisplayName=Qalam (phonetic Arabic keyboard)
UninstallDisplayIcon={app}\Qalam.exe
OutputDir={#OutDir}
OutputBaseFilename=Qalam-Setup-{#Arch}
SetupIconFile=..\assets\qalam-on.ico
WizardStyle=modern
Compression=lzma2/max
SolidCompression=yes
CloseApplications=yes
RestartApplications=no
#if Arch == "arm64"
ArchitecturesAllowed=arm64
ArchitecturesInstallIn64BitMode=arm64
#else
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
#endif

[Languages]
Name: "en"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "startup"; Description: "Start Qalam when I sign in to Windows (recommended)"
Name: "desktopicon"; Description: "Add a desktop shortcut"; Flags: unchecked

[Files]
Source: "{#Exe}"; DestDir: "{app}"; DestName: "Qalam.exe"; Flags: ignoreversion

[Icons]
Name: "{userprograms}\Qalam"; Filename: "{app}\Qalam.exe"; Comment: "Qalam phonetic Arabic keyboard"
Name: "{userdesktop}\Qalam"; Filename: "{app}\Qalam.exe"; Tasks: desktopicon

[Registry]
; The same value the tray menu's "Start with Windows" switches.
Root: HKCU; Subkey: "Software\Microsoft\Windows\CurrentVersion\Run"; ValueType: string; ValueName: "Qalam"; \
  ValueData: """{app}\Qalam.exe"""; Tasks: startup; Flags: uninsdeletevalue

[Run]
Filename: "{app}\Qalam.exe"; Description: "Start Qalam now and set it up"; Flags: nowait postinstall skipifsilent

[UninstallDelete]
; Settings and the guide page Qalam keeps for this user.
Type: filesandordirs; Name: "{userappdata}\Qalam"
Type: files; Name: "{app}\Qalam.old.exe"
Type: files; Name: "{app}\Qalam.new.exe"

[Code]
const
  WM_CLOSE = $0010;

// Ask a running Qalam to quit (its hidden tray window is class "QalamTray") so its exe can be replaced.
procedure CloseQalam();
var
  Wnd: HWND;
  I: Integer;
begin
  Wnd := FindWindowByClassName('QalamTray');
  if Wnd = 0 then
    Exit;
  PostMessage(Wnd, WM_CLOSE, 0, 0);
  for I := 1 to 20 do
  begin
    Sleep(150);
    if FindWindowByClassName('QalamTray') = 0 then
      Exit;
  end;
end;

function PrepareToInstall(var NeedsRestart: Boolean): String;
begin
  CloseQalam();
  Result := '';
end;

function InitializeUninstall(): Boolean;
begin
  CloseQalam();
  Result := True;
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
  // An install made by the app itself (before this installer existed) has its own entry in Apps.
  if CurUninstallStep = usPostUninstall then
    RegDeleteKeyIncludingSubkeys(HKCU, 'Software\Microsoft\Windows\CurrentVersion\Uninstall\Qalam');
end;
