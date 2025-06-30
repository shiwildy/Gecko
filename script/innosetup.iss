#define MyAppName "Gecko"
#define MyAppVersion "1.0.2"
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

[Tasks]
Name: "modifypath"; Description: "Add to PATH Environment"; Flags: checkedonce

[Files]
Source: "..\gecko.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\bin\*"; DestDir: "{app}\bin\"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\etc\*"; DestDir: "{app}\etc\"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\www\*"; DestDir: "{app}\www\"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\logs\*"; DestDir: "{app}\logs\"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\tmp\*"; DestDir: "{app}\tmp\"; Flags: ignoreversion recursesubdirs createallsubdirs

[Icons]
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; IconFilename: "{app}\etc\logo\logo.ico"
Name: "{commondesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; IconFilename: "{app}\etc\logo\logo.ico"

[Registry]
Root: HKCU; Subkey: "Environment"; ValueType: string; ValueName: "Path"; ValueData: "{olddata};{app}\bin\php\php;{app}\bin\mysql\bin;{app}\bin\ngrok;{app}\bin\pgsql\bin;{app}\bin\cloudflared;{app}\bin\composer"; Tasks: modifypath

[Code]
procedure SplitString(const S, Delim: string; var A: TArrayOfString);
var
  I, P: Integer;
  Temp: string;
begin
  SetArrayLength(A, 0);
  Temp := S;
  I := 0;
  while Length(Temp) > 0 do
  begin
    P := Pos(Delim, Temp);
    if P = 0 then
    begin
      SetArrayLength(A, I + 1);
      A[I] := Temp;
      Break;
    end
    else
    begin
      SetArrayLength(A, I + 1);
      A[I] := Copy(Temp, 1, P - 1);
      Temp := Copy(Temp, P + Length(Delim), Length(Temp));
      I := I + 1;
    end;
  end;
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
var
  AppPath: string;
  CurrentPath, NewPath: string;
  Paths: TArrayOfString;
  I: Integer;
begin
  if CurUninstallStep = usPostUninstall then
  begin
    AppPath := Lowercase(ExpandConstant('{app}'));
    if RegQueryStringValue(HKEY_CURRENT_USER, 'Environment', 'Path', CurrentPath) then
    begin
      SplitString(CurrentPath, ';', Paths);
      NewPath := '';
      for I := 0 to GetArrayLength(Paths) - 1 do
      begin
        if (Trim(Paths[I]) <> '') and (Pos(AppPath, Lowercase(Paths[I])) = 0) then
        begin
          if NewPath <> '' then
            NewPath := NewPath + ';';
          NewPath := NewPath + Paths[I];
        end;
      end;
      RegWriteStringValue(HKEY_CURRENT_USER, 'Environment', 'Path', NewPath);
    end;
  end;
end;
