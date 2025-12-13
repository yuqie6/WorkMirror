#define AppName "Mirror"
#define AppPublisher "yuqie6"
#define AppURL "https://github.com/yuqie6/mirror"
#define AgentExeName "mirror.exe"
#define UIExeName "mirror-ui.exe"

#define AppVersion GetEnv("MIRROR_APP_VERSION")
#if AppVersion == ""
  #define AppVersion "0.1.0"
#endif

#define SourceDir GetEnv("MIRROR_STAGING_DIR")
#if SourceDir == ""
  #define SourceDir "."
#endif

#define OutputDir GetEnv("MIRROR_OUTPUT_DIR")
#if OutputDir == ""
  #define OutputDir "."
#endif

[Setup]
AppId={{8D63DAF9-9C09-4C95-8E5C-2E1D08D2A8E6}
AppName={#AppName}
AppVersion={#AppVersion}
AppPublisher={#AppPublisher}
AppPublisherURL={#AppURL}
AppSupportURL={#AppURL}
AppUpdatesURL={#AppURL}
DefaultDirName={localappdata}\Programs\{#AppName}
DefaultGroupName={#AppName}
DisableProgramGroupPage=yes
PrivilegesRequired=lowest
Compression=lzma2
SolidCompression=yes
WizardStyle=modern
OutputDir={#OutputDir}
OutputBaseFilename={#AppName}-Setup-{#AppVersion}
UninstallDisplayIcon={app}\{#UIExeName}

[Languages]
Name: "chinesesimplified"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "创建桌面快捷方式"; GroupDescription: "附加任务"; Flags: unchecked

[Files]
Source: "{#SourceDir}\{#AgentExeName}"; DestDir: "{app}"; Flags: ignoreversion
Source: "{#SourceDir}\{#UIExeName}"; DestDir: "{app}"; Flags: ignoreversion
Source: "{#SourceDir}\config\*"; DestDir: "{app}\config"; Flags: ignoreversion recursesubdirs createallsubdirs; Check: DirExists(ExpandConstant('{#SourceDir}\config'))

[Icons]
Name: "{group}\Mirror UI"; Filename: "{app}\{#UIExeName}"
Name: "{group}\卸载 Mirror"; Filename: "{uninstallexe}"
Name: "{userdesktop}\Mirror UI"; Filename: "{app}\{#UIExeName}"; Tasks: desktopicon

[Run]
Filename: "{app}\{#UIExeName}"; Description: "启动 Mirror"; Flags: nowait postinstall skipifsilent

[Code]
var
  PurgeUserData: Boolean;

function InitializeUninstall(): Boolean;
begin
  Result := True;
  PurgeUserData := False;

  // 询问用户是否清理数据
  if MsgBox('是否同时删除所有用户数据（历史记录、数据库）和配置文件？' #13#10 #13#10 '选择“是”将彻底清除所有数据。' #13#10 '选择“否”保留数据以备将来使用。', mbConfirmation, MB_YESNO) = IDYES then
  begin
    PurgeUserData := True;
  end;
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
  if CurUninstallStep = usUninstall then
  begin
    if PurgeUserData then
    begin
      // 1. 清除 Roaming 数据目录 (数据库默认位置)
      // {userappdata} = C:\Users\xxx\AppData\Roaming
      DelTree(ExpandConstant('{userappdata}\Mirror'), True, True, True);
      
      // 2. 清除安装目录下的 config (因为 config.yaml 可能会被修改，导致卸载保留)
      DelTree(ExpandConstant('{app}\config'), True, True, True);
    end;
  end;
end;

