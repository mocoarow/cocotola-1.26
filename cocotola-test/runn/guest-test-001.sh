#!/bin/bash
set -euo pipefail

export ORGANIZATION_NAME="${ORGANIZATION_NAME:-cocotola}"

runn run --debug guest-test-001.yml
