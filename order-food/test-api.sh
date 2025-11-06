#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

BASE_URL="${1:-http://localhost:8080}"

echo -e "${BLUE}Testing Order Food API at ${BASE_URL}${NC}\n"

# Test health endpoint
echo -e "${YELLOW}1. Testing Health Endpoint${NC}"
curl -s "${BASE_URL}/health" | jq '.'
echo -e "\n"

# Test ready endpoint
echo -e "${YELLOW}2. Testing Ready Endpoint${NC}"
curl -s "${BASE_URL}/ready" | jq '.'
echo -e "\n"

# Test list products
echo -e "${YELLOW}3. Testing List Products${NC}"
curl -s "${BASE_URL}/api/product" | jq '.'
echo -e "\n"

# Test get specific product
echo -e "${YELLOW}4. Testing Get Product by ID (ID=1)${NC}"
curl -s "${BASE_URL}/api/product/1" | jq '.'
echo -e "\n"

# Test get non-existent product
echo -e "${YELLOW}5. Testing Get Non-existent Product (ID=999)${NC}"
curl -s "${BASE_URL}/api/product/999" | jq '.'
echo -e "\n"

# Test place order without API key (should fail)
echo -e "${YELLOW}6. Testing Place Order WITHOUT API Key (should fail with 401)${NC}"
curl -s -X POST "${BASE_URL}/api/order" \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
      {"productId": "1", "quantity": 2}
    ]
  }' | jq '.'
echo -e "\n"

# Test place order with wrong API key (should fail)
echo -e "${YELLOW}7. Testing Place Order with WRONG API Key (should fail with 403)${NC}"
curl -s -X POST "${BASE_URL}/api/order" \
  -H "Content-Type: application/json" \
  -H "api_key: wrongkey" \
  -d '{
    "items": [
      {"productId": "1", "quantity": 2}
    ]
  }' | jq '.'
echo -e "\n"

# Test place order with valid API key
echo -e "${YELLOW}8. Testing Place Order with VALID API Key${NC}"
curl -s -X POST "${BASE_URL}/api/order" \
  -H "Content-Type: application/json" \
  -H "api_key: apitest" \
  -d '{
    "items": [
      {"productId": "1", "quantity": 2},
      {"productId": "3", "quantity": 1}
    ],
    "couponCode": "SAVE10"
  }' | jq '.'
echo -e "\n"

# Test place order with invalid product
echo -e "${YELLOW}9. Testing Place Order with INVALID Product (should fail)${NC}"
curl -s -X POST "${BASE_URL}/api/order" \
  -H "Content-Type: application/json" \
  -H "api_key: apitest" \
  -d '{
    "items": [
      {"productId": "999", "quantity": 1}
    ]
  }' | jq '.'
echo -e "\n"

# Test place order with invalid request body
echo -e "${YELLOW}10. Testing Place Order with INVALID Request Body (should fail)${NC}"
curl -s -X POST "${BASE_URL}/api/order" \
  -H "Content-Type: application/json" \
  -H "api_key: apitest" \
  -d '{
    "items": []
  }' | jq '.'
echo -e "\n"

echo -e "${GREEN}All tests completed!${NC}"
