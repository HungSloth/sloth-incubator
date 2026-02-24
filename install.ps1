# Sloth Incubator installer for Windows
# Usage: iwr -useb https://raw.githubusercontent.com/HungSloth/sloth-incubator/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

$repo = "HungSloth/sloth-incubator"
$binaryName = "incubator"
$installDir = "$env:LOCALAPPDATA\Programs\incubator"

# Detect architecture
$arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }

Write-Host "Detected: windows/$arch"

# Get latest release
$release = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest"
$version = $release.tag_name
Write-Host "Latest version: $version"

# Download
$assetName = "${binaryName}_windows_${arch}.tar.gz"
$asset = $release.assets | Where-Object { $_.name -eq $assetName }

if (-not $asset) {
    Write-Error "No release asset found for windows/$arch"
    exit 1
}

$downloadUrl = $asset.browser_download_url
Write-Host "Downloading $downloadUrl..."

$tmpDir = New-Item -ItemType Directory -Path "$env:TEMP\incubator-install-$(Get-Random)"
$tmpFile = Join-Path $tmpDir $assetName

Invoke-WebRequest -Uri $downloadUrl -OutFile $tmpFile

# Extract (requires tar, available in Windows 10+)
tar -xzf $tmpFile -C $tmpDir

# Install
New-Item -ItemType Directory -Path $installDir -Force | Out-Null
Move-Item -Path (Join-Path $tmpDir "$binaryName.exe") -Destination (Join-Path $installDir "$binaryName.exe") -Force

# Clean up
Remove-Item -Recurse -Force $tmpDir

Write-Host "Installed $binaryName to $installDir"

# Check PATH
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$installDir*") {
    Write-Host ""
    Write-Host "Adding $installDir to your PATH..."
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$installDir", "User")
    $env:Path = "$env:Path;$installDir"
    Write-Host "PATH updated. You may need to restart your terminal."
}

# Verify
Write-Host ""
& (Join-Path $installDir "$binaryName.exe") version
Write-Host "Installation complete!"
