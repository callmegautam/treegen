#Requires -Version 5.1
[CmdletBinding()]
param(
	[string]$Version = $env:TREEGEN_VERSION,
	[string]$InstallDir = $env:TREEGEN_INSTALL_DIR,
	[string]$BaseUrl = $env:TREEGEN_BASE_URL
)

$ErrorActionPreference = "Stop"

$repo = "callmegautam/treegen"
$binary = "treegen"

function Get-Arch {
	$machine = $env:PROCESSOR_ARCHITEW6432
	if (-not $machine) { $machine = $env:PROCESSOR_ARCHITECTURE }

	switch ($machine) {
		"AMD64" { return "amd64" }
		"ARM64" { return "arm64" }
		default { throw "unsupported architecture: $machine" }
	}
}

if (-not $InstallDir) {
	$InstallDir = Join-Path $env:LOCALAPPDATA "Programs\treegen"
}
if (-not $Version) {
	$Version = "latest"
}
if (-not $BaseUrl) {
	if ($Version -eq "latest") {
		$BaseUrl = "https://github.com/$repo/releases/latest/download"
	} else {
		$BaseUrl = "https://github.com/$repo/releases/download/$Version"
	}
}

$arch = Get-Arch
$archive = "${binary}_windows_${arch}.zip"
$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ([System.IO.Path]::GetRandomFileName())
New-Item -ItemType Directory -Path $tmp -Force | Out-Null

try {
	Write-Host "Downloading $binary (windows/$arch)..."
	$archivePath = Join-Path $tmp $archive
	Invoke-WebRequest -Uri "$BaseUrl/$archive" -OutFile $archivePath -UseBasicParsing

	$sumsPath = Join-Path $tmp "checksums.txt"
	$haveSums = $true
	try {
		Invoke-WebRequest -Uri "$BaseUrl/checksums.txt" -OutFile $sumsPath -UseBasicParsing
	} catch {
		$haveSums = $false
		Write-Warning "checksums.txt not available, skipping verification"
	}

	if ($haveSums) {
		$actual = (Get-FileHash -Path $archivePath -Algorithm SHA256).Hash.ToLower()
		$line = Select-String -Path $sumsPath -Pattern ([regex]::Escape($archive) + '$') | Select-Object -First 1
		if (-not $line) { throw "no checksum listed for $archive" }
		$expected = ($line.Line -split '\s+')[0].ToLower()
		if ($actual -ne $expected) { throw "checksum mismatch for $archive" }
		Write-Host "Checksum verified."
	}

	Expand-Archive -Path $archivePath -DestinationPath $tmp -Force
	$exe = Join-Path $tmp "$binary.exe"
	if (-not (Test-Path $exe)) { throw "archive did not contain $binary.exe" }

	New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
	Move-Item -Path $exe -Destination (Join-Path $InstallDir "$binary.exe") -Force
	Write-Host "Installed $binary to $InstallDir\$binary.exe"

	$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
	$entries = @()
	if ($userPath) {
		$entries = $userPath -split ';' | Where-Object { $_ -ne "" }
	}

	if ($entries -notcontains $InstallDir) {
		$updated = (@($entries) + $InstallDir) -join ';'
		[Environment]::SetEnvironmentVariable("Path", $updated, "User")
		Write-Host "Added $InstallDir to your PATH, restart your terminal to pick it up."
	}
} finally {
	Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}
