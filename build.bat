@echo off
setlocal enabledelayedexpansion

REM Build the Docker image for Windows
docker build --build-arg GOOS=windows --build-arg GOARCH=amd64 -t go-app-builder .

REM Run the container to copy the Windows binary
docker run --rm -v "%cd%/output:/output" go-app-builder cp /root/main /output/patron.exe

REM Optionally, move the binary to a directory in PATH
move output\patron.exe %windir%\System32
