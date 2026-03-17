#!/bin/bash

# Test Authentication Flow
# This script tests the complete authentication flow

BASE_URL="http://localhost:8080/v1"
EMAIL="test@example.com"
PASSWORD="testpassword123"
NAME="Test User"

echo "🔐 Testing Ascend API Authentication Flow"
echo "=========================================="
echo ""

# Test 1: Register a new user
echo "1️⃣  Testing user registration..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$EMAIL\",
    \"password\": \"$PASSWORD\",
    \"name\": \"$NAME\",
    \"body_weight\": 75.5
  }")

echo "$REGISTER_RESPONSE" | jq '.'

# Extract access token from registration
ACCESS_TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.access_token')
REFRESH_TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.refresh_token')

if [ "$ACCESS_TOKEN" != "null" ] && [ -n "$ACCESS_TOKEN" ]; then
  echo "✅ Registration successful! Access token received."
else
  echo "❌ Registration failed!"
  exit 1
fi

echo ""
echo "2️⃣  Testing login with correct credentials..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$EMAIL\",
    \"password\": \"$PASSWORD\"
  }")

echo "$LOGIN_RESPONSE" | jq '.'

LOGIN_ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token')

if [ "$LOGIN_ACCESS_TOKEN" != "null" ] && [ -n "$LOGIN_ACCESS_TOKEN" ]; then
  echo "✅ Login successful!"
else
  echo "❌ Login failed!"
  exit 1
fi

echo ""
echo "3️⃣  Testing login with incorrect password..."
WRONG_LOGIN=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$EMAIL\",
    \"password\": \"wrongpassword\"
  }")

ERROR_CODE=$(echo "$WRONG_LOGIN" | jq -r '.error.code')

if [ "$ERROR_CODE" = "AUTHENTICATION_FAILED" ]; then
  echo "✅ Correctly rejected invalid credentials"
else
  echo "❌ Should have rejected invalid credentials"
fi

echo ""
echo "4️⃣  Testing protected endpoint /auth/me with valid token..."
ME_RESPONSE=$(curl -s -X GET "$BASE_URL/auth/me" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "$ME_RESPONSE" | jq '.'

USER_EMAIL=$(echo "$ME_RESPONSE" | jq -r '.email')

if [ "$USER_EMAIL" = "$EMAIL" ]; then
  echo "✅ Successfully accessed protected endpoint!"
else
  echo "❌ Failed to access protected endpoint"
  exit 1
fi

echo ""
echo "5️⃣  Testing protected endpoint without token..."
NO_AUTH_RESPONSE=$(curl -s -X GET "$BASE_URL/auth/me")

ERROR_CODE=$(echo "$NO_AUTH_RESPONSE" | jq -r '.error.code')

if [ "$ERROR_CODE" = "UNAUTHORIZED" ]; then
  echo "✅ Correctly rejected request without token"
else
  echo "❌ Should have rejected request without token"
fi

echo ""
echo "6️⃣  Testing token refresh..."
REFRESH_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/refresh" \
  -H "Content-Type: application/json" \
  -d "{
    \"refresh_token\": \"$REFRESH_TOKEN\"
  }")

echo "$REFRESH_RESPONSE" | jq '.'

NEW_ACCESS_TOKEN=$(echo "$REFRESH_RESPONSE" | jq -r '.access_token')

if [ "$NEW_ACCESS_TOKEN" != "null" ] && [ -n "$NEW_ACCESS_TOKEN" ]; then
  echo "✅ Token refresh successful!"
else
  echo "❌ Token refresh failed!"
  exit 1
fi

echo ""
echo "=========================================="
echo "✅ All authentication tests passed!"
echo "=========================================="
