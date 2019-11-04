#!/usr/bin/env bash
set -e
set -u

sudo curl -L ${OPERATOR_SDK} -o /usr/local/bin/operator-sdk
sudo chmod +x /usr/local/bin/operator-sdk
