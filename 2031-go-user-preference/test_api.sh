#!/bin/bash

BASE_URL="http://localhost:8080"
TIMEOUT=5

echo "=========================================="
echo "用户偏好系统 API 测试"
echo "=========================================="

echo ""
echo "1. 创建测试用户"
echo "------------------------------------------"
USER1=$(curl -s -X POST "$BASE_URL/users" \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser"}' \
  --connect-timeout $TIMEOUT \
  --max-time $TIMEOUT)
echo "创建用户响应: $USER1"
USER_ID=$(echo $USER1 | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
echo "用户 ID: $USER_ID"

echo ""
echo "2. 获取所有偏好项定义"
echo "------------------------------------------"
curl -s "$BASE_URL/preferences" --connect-timeout $TIMEOUT --max-time $TIMEOUT | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/preferences" --connect-timeout $TIMEOUT --max-time $TIMEOUT

echo ""
echo "3. 获取默认偏好"
echo "------------------------------------------"
curl -s "$BASE_URL/preferences/defaults" --connect-timeout $TIMEOUT --max-time $TIMEOUT | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/preferences/defaults" --connect-timeout $TIMEOUT --max-time $TIMEOUT

echo ""
echo "4. 获取用户偏好（未设置，应该使用默认值）"
echo "------------------------------------------"
curl -s "$BASE_URL/users/$USER_ID/preferences" --connect-timeout $TIMEOUT --max-time $TIMEOUT | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/users/$USER_ID/preferences" --connect-timeout $TIMEOUT --max-time $TIMEOUT

echo ""
echo "5. 批量更新用户偏好"
echo "------------------------------------------"
curl -s -X PATCH "$BASE_URL/users/$USER_ID/preferences" \
  -H "Content-Type: application/json" \
  -d '{
    "preferences": [
      {"key": "theme", "value": "dark"},
      {"key": "email_notification", "value": false},
      {"key": "language", "value": "en-US"}
    ]
  }' \
  --connect-timeout $TIMEOUT \
  --max-time $TIMEOUT | python3 -m json.tool 2>/dev/null

echo ""
echo "6. 更新后获取用户偏好（应该显示 user 来源）"
echo "------------------------------------------"
curl -s "$BASE_URL/users/$USER_ID/preferences" --connect-timeout $TIMEOUT --max-time $TIMEOUT | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/users/$USER_ID/preferences" --connect-timeout $TIMEOUT --max-time $TIMEOUT

echo ""
echo "7. 更新单个偏好项"
echo "------------------------------------------"
curl -s -X PUT "$BASE_URL/users/$USER_ID/preferences/font_size" \
  -H "Content-Type: application/json" \
  -d '{"value": "large"}' \
  --connect-timeout $TIMEOUT \
  --max-time $TIMEOUT | python3 -m json.tool 2>/dev/null

echo ""
echo "8. 设置 Web 客户端覆盖"
echo "------------------------------------------"
curl -s -X PATCH "$BASE_URL/users/$USER_ID/preferences?client=web" \
  -H "Content-Type: application/json" \
  -d '{
    "preferences": [
      {"key": "theme", "value": "light"},
      {"key": "push_notification", "value": false}
    ]
  }' \
  --connect-timeout $TIMEOUT \
  --max-time $TIMEOUT | python3 -m json.tool 2>/dev/null

echo ""
echo "9. 获取 Web 客户端偏好（应该显示 client_web 来源）"
echo "------------------------------------------"
curl -s "$BASE_URL/users/$USER_ID/preferences?client=web" --connect-timeout $TIMEOUT --max-time $TIMEOUT | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/users/$USER_ID/preferences?client=web" --connect-timeout $TIMEOUT --max-time $TIMEOUT

echo ""
echo "10. 测试类型错误验证（应该返回 400）"
echo "------------------------------------------"
curl -s -X PUT "$BASE_URL/users/$USER_ID/preferences/email_notification" \
  -H "Content-Type: application/json" \
  -d '{"value": "invalid_string"}' \
  --connect-timeout $TIMEOUT \
  --max-time $TIMEOUT | python3 -m json.tool 2>/dev/null

echo ""
echo "11. 获取变更历史"
echo "------------------------------------------"
curl -s "$BASE_URL/users/$USER_ID/history" --connect-timeout $TIMEOUT --max-time $TIMEOUT | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/users/$USER_ID/history" --connect-timeout $TIMEOUT --max-time $TIMEOUT

echo ""
echo "12. 测试不存在用户（应该返回 404）"
echo "------------------------------------------"
curl -s "$BASE_URL/users/9999/preferences" --connect-timeout $TIMEOUT --max-time $TIMEOUT | python3 -m json.tool 2>/dev/null

echo ""
echo "=========================================="
echo "测试完成！"
echo "=========================================="
