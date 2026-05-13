#!/bin/bash

BASE_URL="http://localhost:8080"

echo "========================================"
echo "配置中心系统测试"
echo "========================================"

echo ""
echo "=== 1. 创建配置（开发环境）==="
echo "--- 创建字符串配置 ---"
curl -s -X POST "$BASE_URL/configs/user-service/dev" \
  -H "Content-Type: application/json" \
  -d '{"key": "db_host", "value": "localhost"}'
echo ""

echo "--- 创建数字配置 ---"
curl -s -X POST "$BASE_URL/configs/user-service/dev" \
  -H "Content-Type: application/json" \
  -d '{"key": "db_port", "value": 3306}'
echo ""

echo "--- 创建布尔配置 ---"
curl -s -X POST "$BASE_URL/configs/user-service/dev" \
  -H "Content-Type: application/json" \
  -d '{"key": "debug", "value": true}'
echo ""

echo "--- 创建 JSON 对象配置 ---"
curl -s -X POST "$BASE_URL/configs/user-service/dev" \
  -H "Content-Type: application/json" \
  -d '{"key": "features", "value": {"auth": true, "cache": false}}'
echo ""

echo ""
echo "=== 2. 读取开发环境配置 ==="
curl -s "$BASE_URL/configs/user-service/dev" | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/configs/user-service/dev"
echo ""

echo ""
echo "=== 3. 为测试环境添加覆盖配置 ==="
curl -s -X POST "$BASE_URL/configs/user-service/test" \
  -H "Content-Type: application/json" \
  -d '{"key": "db_host", "value": "test-db.example.com"}'
echo ""

echo ""
echo "=== 4. 读取测试环境配置（应继承 dev 的配置并覆盖 db_host）==="
echo "预期：db_host=test-db.example.com, db_port=3306, debug=true, features={auth:true,cache:false}"
curl -s "$BASE_URL/configs/user-service/test" | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/configs/user-service/test"
echo ""

echo ""
echo "=== 5. 测试类型验证（修改 db_port 为字符串，应该失败）==="
echo "预期：返回 400 错误，提示类型不匹配"
curl -s -w "\nHTTP状态码: %{http_code}\n" -X POST "$BASE_URL/configs/user-service/dev" \
  -H "Content-Type: application/json" \
  -d '{"key": "db_port", "value": "invalid"}'

echo ""
echo "=== 6. 修改配置（更新 db_port 为 3307）==="
curl -s -X POST "$BASE_URL/configs/user-service/dev" \
  -H "Content-Type: application/json" \
  -d '{"key": "db_port", "value": 3307}'
echo ""

echo ""
echo "=== 7. 验证配置已更新 ==="
curl -s "$BASE_URL/configs/user-service/dev" | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/configs/user-service/dev"
echo ""

echo ""
echo "=== 8. 查看 db_port 的历史记录 ==="
curl -s "$BASE_URL/configs/history/user-service/dev/db_port" | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/configs/history/user-service/dev/db_port"
echo ""

echo ""
echo "=== 9. 测试回滚到版本 1 ==="
curl -s -X POST "$BASE_URL/configs/rollback/user-service/dev/db_port" \
  -H "Content-Type: application/json" \
  -d '{"version": 1, "comment": "测试回滚"}'
echo ""

echo ""
echo "=== 10. 验证回滚后 db_port 应为 3306 ==="
curl -s "$BASE_URL/configs/user-service/dev" | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/configs/user-service/dev"
echo ""

echo ""
echo "=== 11. 再次查看历史记录（应有一条 rollback 记录）==="
curl -s "$BASE_URL/configs/history/user-service/dev/db_port" | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/configs/history/user-service/dev/db_port"
echo ""

echo ""
echo "=== 12. 测试不存在的服务（应返回 404）==="
echo "预期：返回 404"
curl -s -w "\nHTTP状态码: %{http_code}\n" "$BASE_URL/configs/nonexistent-service/dev"

echo ""
echo "=== 13. 创建主实体 ==="
MASTER_RESULT=$(curl -s -X POST "$BASE_URL/masters/user-service" \
  -H "Content-Type: application/json" \
  -d '{"name": "配置文档", "content": "初始版本内容 v1"}')
echo "$MASTER_RESULT"
MASTER_ID=$(echo "$MASTER_RESULT" | grep -o '"master_id":[0-9]*' | grep -o '[0-9]*')
echo "主实体 ID: $MASTER_ID"

echo ""
echo "=== 14. 更新主实体（添加新版本明细）==="
curl -s -X PUT "$BASE_URL/masters/user-service/$MASTER_ID" \
  -H "Content-Type: application/json" \
  -d '{"content": "更新后的内容 v2", "change_note": "添加了新功能"}'
echo ""

echo ""
echo "=== 15. 查看主实体列表 ==="
curl -s "$BASE_URL/masters/user-service" | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/masters/user-service"
echo ""

echo ""
echo "=== 16. 查看明细记录历史 ==="
curl -s "$BASE_URL/masters/user-service/details/$MASTER_ID" | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/masters/user-service/details/$MASTER_ID"
echo ""

echo ""
echo "========================================"
echo "测试完成！"
echo "========================================"
