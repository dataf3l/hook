#!/bin/bash
GOOS=linux GOARCH=amd64 go build -o ./github_actions_linux
upx github_actions_linux
