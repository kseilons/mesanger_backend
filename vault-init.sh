#!/bin/sh

# Wait for Vault to be ready
echo "Waiting for Vault to be ready..."
sleep 5

# Set Vault address and token
export VAULT_ADDR="${VAULT_ADDR:-http://localhost:8200}"
export VAULT_TOKEN="${VAULT_TOKEN:-messenger-token}"

# Wait for Vault to be available
until vault status > /dev/null 2>&1; do
  echo "Waiting for Vault to be available..."
  sleep 2
done

echo "Initializing Vault secrets..."

# Enable KV secrets engine
vault secrets enable -path=secret kv-v2 || echo "KV secrets engine already enabled"

# Note: Namespaces are enterprise-only feature, skipping for dev mode
# export VAULT_NAMESPACE="messenger"

# Create secrets for database
vault kv put secret/database/password value="messenger_password" || echo "Database password secret already exists"

# Create secrets for Redis
vault kv put secret/redis/password value="redis_password" || echo "Redis password secret already exists"

# Create secrets for JWT
vault kv put secret/jwt/secret value="your-super-secret-jwt-key-change-this-in-production" || echo "JWT secret already exists"

echo "Vault initialization completed!"
echo "Vault is ready at: $VAULT_ADDR"
echo "Token: $VAULT_TOKEN"