$ErrorActionPreference = 'Stop'

$OwnerRepo = 'pimatis/mavetis'
$Project = 'mavetis'
$Version = $env:MAVETIS_VERSION
if ([string]::IsNullOrWhiteSpace($Version)) {
  $Version = 'latest'
}
$InstallDir = $env:MAVETIS_INSTALL_DIR
if ([string]::IsNullOrWhiteSpace($InstallDir)) {
  $InstallDir = Join-Path $HOME 'AppData\Local\mavetis\bin'
}

$Arch = $env:PROCESSOR_ARCHITECTURE.ToLowerInvariant()
if ($Arch -eq 'x86_64') {
  $Arch = 'amd64'
}
if ($Arch -eq 'amd64') {
  $Arch = 'amd64'
}
if ($Arch -eq 'aarch64') {
  $Arch = 'arm64'
}
if ($Arch -eq 'arm64') {
  $Arch = 'arm64'
}
if ($Arch -ne 'amd64' -and $Arch -ne 'arm64') {
  throw "unsupported architecture: $Arch"
}

$Archive = "${Project}_windows_${Arch}.zip"
$Checksum = "${Archive}.sha256"
$BaseUrl = "https://github.com/$OwnerRepo/releases"
if ($Version -eq 'latest') {
  $AssetUrl = "$BaseUrl/latest/download/$Archive"
  $ChecksumUrl = "$BaseUrl/latest/download/$Checksum"
}
if ($Version -ne 'latest') {
  $CleanVersion = $Version.TrimStart('v')
  $AssetUrl = "$BaseUrl/download/v$CleanVersion/$Archive"
  $ChecksumUrl = "$BaseUrl/download/v$CleanVersion/$Checksum"
}

$TempDir = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $TempDir | Out-Null
try {
  $ArchivePath = Join-Path $TempDir $Archive
  $ChecksumPath = Join-Path $TempDir $Checksum
  Invoke-WebRequest -Uri $AssetUrl -OutFile $ArchivePath
  Invoke-WebRequest -Uri $ChecksumUrl -OutFile $ChecksumPath

  $Expected = (Get-Content $ChecksumPath -Raw).Split(' ', [System.StringSplitOptions]::RemoveEmptyEntries)[0].Trim().ToLowerInvariant()
  $Actual = (Get-FileHash -Algorithm SHA256 -Path $ArchivePath).Hash.ToLowerInvariant()
  if ($Expected -ne $Actual) {
    throw 'checksum verification failed'
  }

  Expand-Archive -Path $ArchivePath -DestinationPath $TempDir -Force
  $BinaryPath = Join-Path $TempDir "$Project.exe"
  if (-not (Test-Path -LiteralPath $BinaryPath)) {
    throw "release archive did not contain $Project.exe"
  }

  New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
  Copy-Item -Path $BinaryPath -Destination (Join-Path $InstallDir "$Project.exe") -Force
  Write-Host "$Project installed to $(Join-Path $InstallDir "$Project.exe")"
  Write-Host "Run '$Project update --check' to verify future releases."
  if (-not ($env:PATH.Split(';') -contains $InstallDir)) {
    Write-Host "Add $InstallDir to PATH if the command is not found."
  }
}
finally {
  Remove-Item -Recurse -Force $TempDir -ErrorAction SilentlyContinue
}
