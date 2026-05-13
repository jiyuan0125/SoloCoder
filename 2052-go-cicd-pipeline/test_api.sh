#!/bin/bash
set -e

BASE_URL="http://localhost:6666"

extract_id() {
  echo "$1" | sed 's/.*"ID":"\([^"]*\)".*/\1/'
}

echo "=== 1. 创建流水线 ==="
PIPELINE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/pipelines" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "API测试流水线",
    "description": "用于API测试的流水线",
    "total_amount": 10000,
    "triggers": ["manual"],
    "phases": [
      {
        "type": "build",
        "name": "构建",
        "task_mode": "serial",
        "planned_amount": 5000,
        "tasks": [
          {
            "name": "编译",
            "script": "echo build_start && sleep 1 && echo build_done",
            "timeout_sec": 30,
            "failure_strategy": "stop"
          }
        ]
      },
      {
        "type": "test",
        "name": "测试",
        "task_mode": "serial",
        "planned_amount": 5000,
        "tasks": [
          {
            "name": "单元测试",
            "script": "echo test_start && sleep 1 && echo test_done",
            "timeout_sec": 30,
            "failure_strategy": "stop"
          }
        ]
      }
    ]
  }')

PIPELINE_ID=$(extract_id "$PIPELINE_RESPONSE")
echo "流水线ID: $PIPELINE_ID"

echo ""
echo "=== 2. 获取流水线详情 ==="
curl -s "$BASE_URL/api/pipelines/$PIPELINE_ID" | head -c 300
echo "..."

echo ""
echo "=== 3. 触发执行 ==="
EXEC_RESPONSE=$(curl -s -X POST "$BASE_URL/api/pipelines/$PIPELINE_ID/trigger" \
  -H "Content-Type: application/json" \
  -d '{"trigger_type": "manual"}')

EXEC_ID=$(extract_id "$EXEC_RESPONSE")
echo "执行ID: $EXEC_ID"

echo ""
echo "=== 4. 等待执行完成 (6秒) ==="
sleep 6

echo ""
echo "=== 5. 检查执行状态 ==="
curl -s "$BASE_URL/api/executions/$EXEC_ID"

echo ""
echo ""
echo "=== 6. 测试404 (不存在的流水线) ==="
curl -s -w "\nHTTP状态: %{http_code}\n" "$BASE_URL/api/pipelines/nonexistent"

echo ""
echo "=== 7. 测试金额调整 ==="
curl -s -X POST "$BASE_URL/api/pipelines/$PIPELINE_ID/amount" \
  -H "Content-Type: application/json" \
  -d '{"new_total": 20000}'

echo ""
echo "=== 测试完成 ==="
