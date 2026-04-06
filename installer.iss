[Setup]
AppName=jpegrm
AppVersion=1.0.0
AppPublisher=shuntaka9576
DefaultDirName={autopf}\jpegrm
DefaultGroupName=jpegrm
OutputDir=dist
OutputBaseFilename=jpegrm-setup
Compression=lzma2
SolidCompression=yes
ChangesEnvironment=yes
ArchitecturesInstallIn64BitMode=x64compatible
PrivilegesRequired=lowest

[Files]
Source: "dist\jpegrm.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "README-windows.txt"; DestDir: "{app}"; Flags: ignoreversion

[Registry]
Root: HKCU; Subkey: "Environment"; ValueType: expandsz; ValueName: "Path"; ValueData: "{olddata};{app}"; Check: NeedsAddPath(ExpandConstant('{app}'))

[Code]
function NeedsAddPath(Param: string): Boolean;
var
  OrigPath: string;
begin
  if not RegQueryStringValue(HKEY_CURRENT_USER, 'Environment', 'Path', OrigPath) then
  begin
    Result := True;
    exit;
  end;
  Result := Pos(';' + Param + ';', ';' + OrigPath + ';') = 0;
end;
