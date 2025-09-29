#!/bin/sh
set -e

# Wait for Hydra to be ready
until curl -s http://hydra:4445/health/ready; do
  echo "Waiting for Hydra to be ready..."
  sleep 5
done

# Create an OAuth2 client
ory create oauth2-client \
  --endpoint http://hydra:4445 \
  --grant-type client_credentials \
  --response-type token \
  --token-endpoint-auth-method client_secret_post \
  --name "Claude Connector Client" \
  --scope "profile"

echo "OAuth2 client created successfully."
