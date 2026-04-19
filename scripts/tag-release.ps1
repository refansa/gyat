# Create a new git tag with an incremented version number.
# Automatically reads the current version from existing tags or cmd/root.go.

param(
    [ValidateSet("patch", "minor", "major")]
    [string]$Part = "patch",

    [string]$SpecificVersion = ""
)

$ErrorActionPreference = "Stop"

function Get-CurrentVersion {
    # Try to get version from latest tag
    $version = git describe --tags --abbrev=0 2>$null
    if ($version) {
        return $version -replace '^v', ''
    }

    # Fall back to reading from cmd/root.go
    $content = Get-Content cmd/root.go -Raw
    if ($content -match 'var Version = "(.+)"') {
        $version = $Matches[1]
        if ($version -and $version -ne "dev") {
            return $version
        }
    }

    return "0.1.0"
}

function Increment-Version {
    param(
        [string]$Version,
        [string]$Part
    )

    $parts = $Version -split '\.'
    $major = [int]$parts[0]
    $minor = [int]$parts[1]
    $patch = [int]$parts[2]

    switch ($Part) {
        "major" { $major++; $minor = 0; $patch = 0 }
        "minor" { $minor++; $patch = 0 }
        "patch" { $patch++ }
    }

    return "$major.$minor.$patch"
}

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir
Set-Location $ProjectRoot

# Fetch latest tags from remote to ensure we have the latest version
Write-Host "Fetching latest tags from origin..."
git fetch --tags origin

# Get current version and compute new version
if ($SpecificVersion) {
    $currentVersion = $SpecificVersion.TrimStart("v")
    $newVersion = $currentVersion
}
else {
    $currentVersion = Get-CurrentVersion
    $newVersion = Increment-Version -Version $currentVersion -Part $Part
}

Write-Host "Current version: v$currentVersion"
Write-Host "New version:     v$newVersion"
Write-Host ""

$response = Read-Host "Create tag v${newVersion}? [y/N]"
if ($response -ne "y" -and $response -ne "Y") {
    Write-Host "Aborted." -ForegroundColor Yellow
    exit 1
}

# Create the tag
git tag -a "v$newVersion" -m "Release v$newVersion"

Write-Host "Created tag v$newVersion" -ForegroundColor Green
Write-Host ""
Write-Host "Push with: git push origin v$newVersion"
