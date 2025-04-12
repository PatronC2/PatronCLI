param (
    [string]$Tag = "latest"
)

$Platform = "windows"
$InstallPath = "$env:Windows\System32\Patron.exe"


$Image = "patronc2/cli:$Platform-$Tag"
$BinaryName = if ($Platform -eq "windows") { "patron.exe" } else { "patron" }

Write-Host "Pulling image $Image..."
docker pull $Image

$containerId = docker create $Image
Write-Host "Extracting $BinaryName to $InstallPath"

if (-not (Test-Path $InstallPath)) {
    New-Item -ItemType Directory -Force -Path $InstallPath | Out-Null
}

docker cp "${containerId}:/${BinaryName}" "${InstallPath}\${BinaryName}"
docker rm $containerId | Out-Null

Write-Host ""
Write-Host "Installed $BinaryName to $InstallPath"
