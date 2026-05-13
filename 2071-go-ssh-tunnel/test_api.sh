#!/bin/bash

BASE_URL="http://localhost:8082"

echo "=== Test 1: Create Tunnel ==="
curl -s -X POST "$BASE_URL/tunnels" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Local Tunnel",
    "type": "local",
    "ssh_server": "example.com:22",
    "ssh_user": "testuser",
    "auth_type": "password",
    "auth_data": "testpass123",
    "local_port": 9999,
    "remote_host": "localhost",
    "remote_port": 22,
    "operator": "test-operator"
  }' | python3 -m json.tool

echo ""
echo ""

echo "=== Test 2: List Tunnels ==="
curl -s "$BASE_URL/tunnels" | python3 -m json.tool

echo ""
echo ""

echo "=== Test 3: List Records ==="
curl -s "$BASE_URL/records" | python3 -m json.tool

echo ""
echo ""

echo "=== Test 4: Get Record 1 ==="
curl -s "$BASE_URL/records/1" | python3 -m json.tool

echo ""
echo ""

echo "=== Test 5: Submit Record 1 for Review ==="
curl -s -X POST "$BASE_URL/records/1/submit" \
  -H "Content-Type: application/json" \
  -d '{"operator": "reviewer"}' | python3 -m json.tool

echo ""
echo ""

echo "=== Test 6: Approve Record 1 ==="
curl -s -X POST "$BASE_URL/records/1/approve" \
  -H "Content-Type: application/json" \
  -d '{"operator": "admin"}' | python3 -m json.tool

echo ""
echo ""

echo "=== Test 7: Get Record 1 History ==="
curl -s "$BASE_URL/records/1/history" | python3 -m json.tool

echo ""
echo ""

echo "=== Test 8: Query non-existent tunnel ==="
curl -s -w "\nHTTP Status: %{http_code}\n" "$BASE_URL/tunnels/9999"

echo ""

echo "=== Test 9: Invalid auth type ==="
curl -s -X POST "$BASE_URL/tunnels" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Invalid Auth",
    "type": "local",
    "ssh_server": "example.com:22",
    "ssh_user": "testuser",
    "auth_type": "invalid_type",
    "auth_data": "test",
    "local_port": 9998,
    "remote_host": "localhost",
    "remote_port": 22
  }' -w "\nHTTP Status: %{http_code}\n"

echo ""
