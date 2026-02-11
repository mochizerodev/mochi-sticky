param(
  [string]$Version = "latest",
  [string]$Repo = "mochizerodev/mochi-sticky",
  [string]$InstallDir = "",
  [string]$FromFile = "",
  [string]$ChecksumFile = "",
  [switch]$SkipChecksum,
  [switch]$Force,
  [switch]$AddToPath
)

$ErrorActionPreference = "Stop"
$BinName = "mochi-sticky.exe"

function Get-Arch {
  $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToLowerInvariant()
  switch ($arch) {
    "x64" { return "amd64" }
    "amd64" { return "amd64" }
    "arm64" { return "arm64" }
    default { throw "Unsupported architecture: $arch (expected amd64 or arm64)." }
  }
}

function Get-ExpectedHash {
  param([string]$Path)
  $raw = (Get-Content -Path $Path -Raw).Trim()
  if ([string]::IsNullOrWhiteSpace($raw)) {
    throw "Checksum file is empty: $Path"
  }
  return (($raw -split "\s+")[0]).ToLowerInvariant()
}

function Resolve-InstallDir {
  if (-not [string]::IsNullOrWhiteSpace($InstallDir)) {
    return $InstallDir
  }
  return Join-Path $env:LOCALAPPDATA "Programs\mochi-sticky\bin"
}

$tempDir = Join-Path ([System.IO.Path]::GetTempPath()) ("mochi-sticky-install-" + [Guid]::NewGuid().ToString("N"))
New-Item -Path $tempDir -ItemType Directory -Force | Out-Null

try {
  if (-not [string]::IsNullOrWhiteSpace($FromFile)) {
    $archivePath = $FromFile
    if (-not (Test-Path -Path $archivePath -PathType Leaf)) {
      throw "Archive not found: $archivePath"
    }
  } else {
    $arch = Get-Arch
    $assetName = "mochi-sticky-windows-$arch.zip"
    if ($Version -eq "latest") {
      $baseUrl = "https://github.com/$Repo/releases/latest/download"
    } else {
      $baseUrl = "https://github.com/$Repo/releases/download/$Version"
    }
    $archivePath = Join-Path $tempDir $assetName
    Invoke-WebRequest -Uri "$baseUrl/$assetName" -OutFile $archivePath
    if ([string]::IsNullOrWhiteSpace($ChecksumFile)) {
      $downloadedChecksum = Join-Path $tempDir "$assetName.sha256"
      Invoke-WebRequest -Uri "$baseUrl/$assetName.sha256" -OutFile $downloadedChecksum
      $ChecksumFile = $downloadedChecksum
    }
  }

  if (-not $SkipChecksum.IsPresent) {
    if ([string]::IsNullOrWhiteSpace($ChecksumFile)) {
      $fallback = "$archivePath.sha256"
      if (Test-Path -Path $fallback -PathType Leaf) {
        $ChecksumFile = $fallback
      } else {
        throw "Checksum file is required (use -ChecksumFile or -SkipChecksum)."
      }
    }

    $expected = Get-ExpectedHash -Path $ChecksumFile
    $actual = (Get-FileHash -Path $archivePath -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($expected -ne $actual) {
      throw "Checksum mismatch for $archivePath (expected $expected, got $actual)."
    }
  }

  $extractDir = Join-Path $tempDir "extract"
  New-Item -Path $extractDir -ItemType Directory -Force | Out-Null
  Expand-Archive -Path $archivePath -DestinationPath $extractDir -Force

  $binaryPath = Join-Path $extractDir $BinName
  if (-not (Test-Path -Path $binaryPath -PathType Leaf)) {
    $candidate = Get-ChildItem -Path $extractDir -Recurse -File -Filter $BinName | Select-Object -First 1
    if ($null -eq $candidate) {
      throw "Failed to find $BinName in archive $archivePath"
    }
    $binaryPath = $candidate.FullName
  }

  $targetDir = Resolve-InstallDir
  New-Item -Path $targetDir -ItemType Directory -Force | Out-Null
  $installPath = Join-Path $targetDir $BinName

  if ((Test-Path -Path $installPath -PathType Leaf) -and (-not $Force.IsPresent)) {
    throw "Binary already exists at $installPath. Re-run with -Force to replace."
  }

  Copy-Item -Path $binaryPath -Destination $installPath -Force
  Write-Host "Installed $BinName to $installPath"

  $sessionPath = $env:PATH
  $normalizedSession = ";$sessionPath;"
  if ($normalizedSession -notlike "*;$targetDir;*") {
    if ($AddToPath.IsPresent) {
      $userPath = [Environment]::GetEnvironmentVariable("PATH", [EnvironmentVariableTarget]::User)
      $normalizedUser = ";" + ($userPath ?? "") + ";"
      if ($normalizedUser -notlike "*;$targetDir;*") {
        $newUserPath = if ([string]::IsNullOrWhiteSpace($userPath)) { $targetDir } else { "$userPath;$targetDir" }
        [Environment]::SetEnvironmentVariable("PATH", $newUserPath, [EnvironmentVariableTarget]::User)
      }
      $env:PATH = "$targetDir;$env:PATH"
      Write-Host "Added $targetDir to user PATH."
    } else {
      Write-Host "Add $targetDir to PATH to run mochi-sticky from any shell."
    }
  }

  Write-Host "Verify with: mochi-sticky --version"
} finally {
  if (Test-Path -Path $tempDir) {
    Remove-Item -Path $tempDir -Recurse -Force
  }
}
