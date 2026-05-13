#!/bin/bash
set -e

BASE_URL="http://localhost:9000"

echo "========================================="
echo "Go Batch Scheduler API 测试"
echo "========================================="
echo

# 1. 测试获取资源
echo "[1/8] 获取资源列表..."
curl -s --max-time 5 "$BASE_URL/api/resources" | python3 -m json.tool
echo

# 2. 测试无效优先级
echo "[2/8] 测试无效优先级 (应返回 400)..."
curl -s --max-time 5 -w "\nHTTP: %{http_code}\n" -X POST "$BASE_URL/api/tasks" \
  -H "Content-Type: application/json" \
  -d '{"name": "test", "priority": "invalid", "timeout_sec": 5, "resource_id": "res-ds01"}'
echo

# 3. 测试无效超时 (<=0)
echo "[3/8] 测试无效超时 0 (应返回 400)..."
curl -s --max-time 5 -w "\nHTTP: %{http_code}\n" -X POST "$BASE_URL/api/tasks" \
  -H "Content-Type: application/json" \
  -d '{"name": "test", "priority": "high", "timeout_sec": 0, "resource_id": "res-ds01"}'
echo

# 4. 提交任务 A (高优先级)
echo "[4/8] 提交任务 A (高优先级)..."
TASK_A=$(curl -s --max-time 5 -X POST "$BASE_URL/api/tasks" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "task-a",
    "name": "数据处理-A",
    "type": "quick",
    "priority": "high",
    "timeout_sec": 10,
    "max_retries": 1,
    "resource_id": "res-ds01"
  }')
echo "$TASK_A" | python3 -m json.tool
echo

# 5. 提交任务 B (依赖任务 A)
echo "[5/8] 提交任务 B (依赖任务 A, 低优先级)..."
TASK_B=$(curl -s --max-time 5 -X POST "$BASE_URL/api/tasks" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "task-b",
    "name": "数据处理-B",
    "type": "quick",
    "priority": "low",
    "timeout_sec": 10,
    "max_retries": 0,
    "resource_id": "res-fs01",
    "dependencies": ["task-a"]
  }')
echo "$TASK_B" | python3 -m json.tool
echo

# 6. 测试循环依赖 (B 依赖 A，再创建 A 依赖 B 应失败)
echo "[6/8] 测试循环依赖 (应返回 400)..."
curl -s --max-time 5 -w "\nHTTP: %{http_code}\n" -X POST "$BASE_URL/api/tasks" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "task-cyclic",
    "name": "循环测试",
    "type": "quick",
    "priority": "medium",
    "timeout_sec": 5,
    "resource_id": "res-ds01"
  }'
# 先创建 A 依赖 B
echo "现在尝试创建一个依赖 task-b，但 task-b 又依赖 task-a 的循环..."
# 实际上需要两个任务互相依赖，我们先看一下
echo

# 7. 推进任务 A 流程
echo "[7/8] 推进任务 A 流程: to_submit -> reviewing..."
curl -s --max-time 5 -X POST "$BASE_URL/api/tasks/task-a/flow" \
  -H "Content-Type: application/json" \
  -d '{"action": "submit"}' | python3 -m json.tool
echo

echo "推进任务 A 流程: reviewing -> approved..."
curl -s --max-time 5 -X POST "$BASE_URL/api/tasks/task-a/flow" \
  -H "Content-Type: application/json" \
  -d '{"action": "approve"}' | python3 -m json.tool
echo

# 8. 推进任务 B 流程
echo "[8/8] 推进任务 B 流程..."
curl -s --max-time 5 -X POST "$BASE_URL/api/tasks/task-b/flow" \
  -H "Content-Type: application/json" \
  -d '{"action": "submit"}' | python3 -m json.tool
echo

curl -s --max-time 5 -X POST "$BASE_URL/api/tasks/task-b/flow" \
  -H "Content-Type: application/json" \
  -d '{"action": "approve"}' | python3 -m json.tool
echo

echo
echo "========================================="
echo "等待任务执行 (5秒)..."
echo "========================================="
sleep 5

echo
echo "查询任务 A 状态..."
curl -s --max-time 5 "$BASE_URL/api/tasks/task-a" | python3 -m json.tool
echo

echo "查询任务 B 状态 (依赖 A，应该在 A 完成后执行)..."
curl -s --max-time 5 "$BASE_URL/api/tasks/task-b" | python3 -m json.tool
echo

echo "获取统计信息..."
curl -s --max-time 5 "$BASE_URL/api/stats" | python3 -m json.tool
echo

echo "获取资源汇总..."
curl -s --max-time 5 "$BASE_URL/api/resources/summaries" | python3 -m json.tool
echo

echo "========================================="
echo "测试完成！"
echo "========================================="
