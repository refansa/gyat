# This script always reinitializes the test workspace from a clean state.
# Any existing tmp/gyat-test directory will be removed and recreated.
#
# This is intended for testing commands that require a staged workspace, such as `gyat commit`.

$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir
$BinDir = Join-Path $ProjectRoot "bin"
$GyatBin = Join-Path $BinDir "gyat.exe"
$TestDir = Join-Path $ProjectRoot "tmp\gyat-test"

Write-Host "Building gyat..."
New-Item -ItemType Directory -Force -Path $BinDir | Out-Null
go build -o $GyatBin .

Write-Host "Cleaning up existing test directory..."
Remove-Item -Recurse -Force "$TestDir" -ErrorAction SilentlyContinue

Write-Host "Creating test umbrella repository at $TestDir..."

New-Item -ItemType Directory -Force -Path "$TestDir\services\auth" | Out-Null
New-Item -ItemType Directory -Force -Path "$TestDir\services\api" | Out-Null
New-Item -ItemType Directory -Force -Path "$TestDir\services\web" | Out-Null

"# Auth Service" | Out-File -FilePath "$TestDir\services\auth\README.md" -Encoding utf8
"# API Service" | Out-File -FilePath "$TestDir\services\api\README.md" -Encoding utf8
"# Web Service" | Out-File -FilePath "$TestDir\services\web\README.md" -Encoding utf8

git -C "$TestDir\services\auth" init --quiet
git -C "$TestDir\services\api" init --quiet
git -C "$TestDir\services\web" init --quiet

git -C "$TestDir\services\auth" add README.md
git -C "$TestDir\services\api" add README.md
git -C "$TestDir\services\web" add README.md

git -C "$TestDir\services\auth" commit -m "Initial commit" --no-gpg-sign --quiet
git -C "$TestDir\services\api" commit -m "Initial commit" --no-gpg-sign --quiet
git -C "$TestDir\services\web" commit -m "Initial commit" --no-gpg-sign --quiet

git init "$TestDir" --quiet

# Add services to .gitignore (source repos to track, not commit)
"/services" | Out-File -FilePath "$TestDir\.gitignore" -Append -Encoding utf8

Write-Host "Initializing gyat workspace..."
Push-Location $TestDir
try {
    & $GyatBin init
    & $GyatBin track services/auth
    & $GyatBin track services/api
    & $GyatBin track services/web
    & $GyatBin update
} finally {
    Pop-Location
}

"# Auth Service (modified)" | Out-File -FilePath "$TestDir\auth\README.md" -Encoding utf8
"# API Service (modified)" | Out-File -FilePath "$TestDir\api\README.md" -Encoding utf8
"# Web Service (modified)" | Out-File -FilePath "$TestDir\web\README.md" -Encoding utf8

Write-Host "Staging changes in gyat workspace..."
Push-Location $TestDir
try {
    & $GyatBin exec -- git add .
    & $GyatBin list
} finally {
    Pop-Location
}