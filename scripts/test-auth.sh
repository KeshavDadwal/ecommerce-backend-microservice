#!/usr/bin/env bash
# Test auth-service gRPC: RegisterAdmin -> LoginAdmin -> ValidateToken
# Requires: grpcurl (go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest)
# Run from repo root with auth-service up: docker compose up -d auth-service

set -e
ADDR="${1:-localhost:8080}"

echo "=== 1. RegisterAdmin ==="
REGISTER_RESP=$(grpcurl -plaintext -d '{
  "email": "admin@example.com",
  "password": "secret123",
  "full_name": "Admin User"
}' "$ADDR" auth.v1.AuthService/RegisterAdmin)
echo "$REGISTER_RESP"
ACCESS_TOKEN=$(echo "$REGISTER_RESP" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
if [ -z "$ACCESS_TOKEN" ]; then
  echo "Failed to get access_token from RegisterAdmin"
  exit 1
fi
echo "Got access_token (length ${#ACCESS_TOKEN})"
echo ""

echo "=== 2. LoginAdmin (same user) ==="
LOGIN_RESP=$(grpcurl -plaintext -d '{
  "email": "admin@example.com",
  "password": "secret123"
}' "$ADDR" auth.v1.AuthService/LoginAdmin)
echo "$LOGIN_RESP"
echo ""

echo "=== 3. ValidateToken (using token from step 1) ==="
grpcurl -plaintext -d "{\"accessToken\": \"$ACCESS_TOKEN\"}" "$ADDR" auth.v1.AuthService/ValidateToken
echo ""
echo "Done."
