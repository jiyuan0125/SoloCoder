#!/bin/bash

BASE_URL="http://localhost:8888/api"

echo "=== 测试JSON序列化修复 ==="
echo ""

echo "1. 等待服务启动..."
for i in {1..30}; do
    if curl -s "$BASE_URL/categories" > /dev/null 2>&1; then
        echo "服务已启动！"
        break
    fi
    sleep 1
done

echo ""
echo "2. 测试 GET /api/categories ..."
curl -s "$BASE_URL/categories" | head -c 2000
echo ""

echo ""
echo "3. 测试 GET /api/questions ..."
curl -s "$BASE_URL/questions" | head -c 3000
echo ""

echo ""
echo "4. 测试 GET /api/questions/1 ..."
curl -s "$BASE_URL/questions/1" | head -c 3000
echo ""

echo ""
echo "=== 测试完成 ==="
