#define MyAppName "Gecko"
#define MyAppVersion "1.0.0"
#define MyAppPublisher "ShiWildy"
#define MyAppURL "https://github.com/shiwildy/Gecko.git"
#define MyAppExeName "gecko.exe"

[Setup]
AppId={{527E0227-F3E6-4C74-861F-3A64C85AF581}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}
DefaultDirName=C:\{#MyAppName}
DisableDirPage=yes
DisableProgramGroupPage=yes
OutputDir=..\output
OutputBaseFilename=gecko-setup
SetupIconFile=..\etc\logo\logo.ico
Compression=lzma
SolidCompression=yes
WizardStyle=modern
PrivilegesRequired=admin
ChangesEnvironment=yes

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Files]
Source: "..\gecko.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\bin\*"; DestDir: "{app}\bin\"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\etc\*"; DestDir: "{app}\etc\"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\www\*"; DestDir: "{app}\www\"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\logs\*"; DestDir: "{app}\logs\"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\tmp\*"; DestDir: "{app}\tmp\"; Flags: ignoreversion recursesubdirs createallsubdirs

[Icons]
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"
Name: "{commondesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"

[Registry]
Root: HKCU; Subkey: "Environment"; ValueType: string; ValueName: "Path"; ValueData: "{olddata};{app}\bin\php;{app}\bin\mysql\bin;{app}\bin\ngrok";
