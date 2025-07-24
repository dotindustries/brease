#!/bin/bash
set -e
echo "Fetching secrets from Infisical..."
export INFISICAL_TOKEN=$(infisical login --method=universal-auth --client-id=$INFISICAL_CLIENT_ID --client-secret=$INFISICAL_CLIENT_SECRET --plain --silent)
infisical run --projectId=$INFISICAL_PROJECT_ID --env=staging brease