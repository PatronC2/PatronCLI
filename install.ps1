param (
    [string]$Platform = $IsWindows ? "windows" : "linux",
    [string]$Tag = "latest",
    [string]$InstallPath = ""
)

if (-not $InstallPath) {
    if ($IsWindows) {
        $InstallPath = "$env:ProgramFiles\Patron"
    } else {
        $InstallPath = "/usr/local/bin"
    }
}

$Image = "patronc2/cli:$Platform-$Tag"
$BinaryName = if ($Platform -eq "windows") { "patron.exe" } else { "patron" }

Write-Host "Pulling $Image..."
docker pull $Image

$containerId = docker create $Image
Write-Host "Extracting $BinaryName to $InstallPath"

if (-not (Test-Path $InstallPath)) {
    New-Item -ItemType Directory -Force -Path $InstallPath | Out-Null
}

docker cp "$containerId:/$BinaryName" "$InstallPath/$BinaryName"
docker rm $containerId | Out-Null

Write-Host "`n✅ Installed $BinaryName to $InstallPath"
