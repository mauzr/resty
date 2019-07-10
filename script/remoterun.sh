#!/usr/bin/bash

set -eu

arch=$1  # Host architecture
host=$2  # Hostname to connect to
shift 2

GOARCH=$arch go build -v ./cmd/mauzr  # Build binary
ssh $host "killall mauzr" || true  # Remove target location
scp mauzr $host:mauzr  # Copy binary there
ssh $host ./mauzr $@ # Execute binary remotely
