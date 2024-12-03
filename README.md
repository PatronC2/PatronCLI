# Patron CLI

This project creates a way to interact with Patron remotely via CLI.

## Build
Docker must be installed and running
Run build.sh

## Auth
To start, run `patron auth configure`. This will prompt you to create a profile for your Patron C2 server.
To login to a profile, run `patron auth login --profile my-profile` This will prompt for your password.

## Agents
To list agents, run `patron main.go agents list`
