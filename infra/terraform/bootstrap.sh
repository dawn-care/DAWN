#!/bin/bash
set -euo pipefail
export DEBIAN_FRONTEND=noninteractive

# The Elastic IP is attached shortly after the instance starts.
apt-get -o Acquire::Retries=10 update
apt-get -o Acquire::Retries=10 install -y docker.io docker-compose-v2 git
systemctl enable --now docker

# Deploy the application separately after configuring production TLS and secrets.
install -d -o ubuntu -g ubuntu -m 0750 /opt/dawn
