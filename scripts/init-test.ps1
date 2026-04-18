$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir
$BinDir = Join-Path $ProjectRoot "bin"
$GyatBin = Join-Path $BinDir "gyat.exe"
$TestDir = Join-Path $ProjectRoot "tmp\gyat-test"

Write-Host "Building gyat..."
New-Item -ItemType Directory -Force -Path $BinDir | Out-Null
go build -o $GyatBin .

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

git init "$TestDir" --quiet

Write-Host "Initializing gyat workspace..."
Push-Location $TestDir
try {
    & $GyatBin init
    & $GyatBin track services/auth
    & $GyatBin track services/api
    & $GyatBin track services/web
    & $GyatBin list
} finally {
    Pop-Location
}