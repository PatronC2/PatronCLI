# Patron CLI

This project creates a way to interact with Patron remotely via CLI.

## Install

### Windows
Run `.\install.ps1` from an elevated powershell

### Linux
Run `./install.sh` from an elevated shell

# Build
Docker and the docker buildx plugin must be installed and running 

### Windows
From an elevated powershell, run `make all`

### Linux
From an elevated powershell, run `make all`

## Auth
To start, run `patron auth configure`. This will prompt you to create a profile for your Patron C2 server.
To login to a profile, run `patron auth login --profile my-profile` This will prompt for your password.

## Agents
To list agents, run `patron main.go agents list`
