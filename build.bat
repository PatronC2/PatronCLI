@echo off
setlocal enabledelayedexpansion

REM Build the Docker image
docker build -t go-app-builder .

REM Run the container to copy the built binary
docker run --rm -v "%cd%/output:/output" go-app-builder cp /root/main /output/patron.exe

REM Optionally, move the binary to a folder in PATH
move output\patron.exe %windir%\System32
