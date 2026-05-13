#!/bin/bash
set -x

cd /home/baru/Work/Private/work01/SoloCoder/2024-go-data-export

pkill -f data-export || true
sleep 1

rm -f server.log

./data-export > server.log 2>&1 &
SERVER_PID=$!
echo "Server PID: $SERVER_PID"

sleep 3

echo "=== Test 1: Get tables ==="
curl -s http://127.0.0.1:8080/api/tables
echo ""

echo "=== Test 2: Get users table structure ==="
curl -s http://127.0.0.1:8080/api/tables/users
echo ""

wait $SERVER_PID 2>/dev/null
