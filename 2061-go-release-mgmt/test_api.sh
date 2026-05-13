#!/bin/bash
set -e

BASE_URL="http://localhost:8080"
CURL="curl -s --max-time 10"

echo "=== 1. 测试创建版本 (semver 验证) ==="
echo "测试 1.1: 有效的 semver (1.0.0)"
$CURL -X POST "$BASE_URL/api/releases" \
  -H "Content-Type: application/json" \
  -d '{"version":"1.0.0","description":"Initial release","submitted_by":"user1"}' | head -c 500
echo ""
echo ""

echo "测试 1.2: 无效的 semver (应该返回 400)"
$CURL -w " [HTTP: %{http_code}]" -X POST "$BASE_URL/api/releases" \
  -H "Content-Type: application/json" \
  -d '{"version":"invalid","description":"Bad version","submitted_by":"user1"}'
echo ""
echo ""

echo "测试 1.3: 重复版本号 (应该返回 409)"
$CURL -w " [HTTP: %{http_code}]" -X POST "$BASE_URL/api/releases" \
  -H "Content-Type: application/json" \
  -d '{"version":"1.0.0","description":"Duplicate","submitted_by":"user1"}'
echo ""
echo ""

echo "测试 1.4: 创建第二个版本 1.1.0"
$CURL -X POST "$BASE_URL/api/releases" \
  -H "Content-Type: application/json" \
  -d '{"version":"1.1.0","description":"Second release","submitted_by":"user1"}' | head -c 500
echo ""
echo ""

echo "=== 2. 测试查询版本 ==="
echo "测试 2.1: 查询存在的版本 1.0.0"
$CURL "$BASE_URL/api/releases/1.0.0" | head -c 500
echo ""
echo ""

echo "测试 2.2: 查询不存在的版本 (应该返回 404)"
$CURL -w " [HTTP: %{http_code}]" "$BASE_URL/api/releases/9.9.9"
echo ""
echo ""

echo "=== 3. 测试添加变更记录 ==="
echo "测试 3.1: 给版本 1.0.0 添加变更 (id=1)"
$CURL -X POST "$BASE_URL/api/releases/1/changes" \
  -H "Content-Type: application/json" \
  -d '{"content":"Fixed critical bug in login flow","type":"fix","operator":"user1"}'
echo ""
echo ""

echo "测试 3.2: 添加另一个变更"
$CURL -X POST "$BASE_URL/api/releases/1/changes" \
  -H "Content-Type: application/json" \
  -d '{"content":"Added new API endpoint","type":"feature","operator":"user1"}'
echo ""
echo ""

echo "=== 4. 测试提交审核 ==="
echo "测试 4.1: 提交版本 1.0.0 进入审核"
$CURL -X POST "$BASE_URL/api/releases/1/submit" \
  -H "Content-Type: application/json" \
  -d '{"operator":"user1"}'
echo ""
echo ""

echo "=== 5. 测试审批流程 ==="
echo "测试 5.1: user1 尝试审批自己的版本 (应该返回 403)"
$CURL -w " [HTTP: %{http_code}]" -X POST "$BASE_URL/api/releases/1/approvals" \
  -H "Content-Type: application/json" \
  -d '{"approver":"user1","approved":true}'
echo ""
echo ""

echo "测试 5.2: user2 审批通过"
$CURL -X POST "$BASE_URL/api/releases/1/approvals" \
  -H "Content-Type: application/json" \
  -d '{"approver":"user2","approved":true}'
echo ""
echo ""

echo "测试 5.3: 审批不通过但没填原因 (应该返回 400)"
$CURL -w " [HTTP: %{http_code}]" -X POST "$BASE_URL/api/releases/1/approvals" \
  -H "Content-Type: application/json" \
  -d '{"approver":"user3","approved":false}'
echo ""
echo ""

echo "测试 5.4: user3 审批通过 (达到 2 人，状态变为 approved)"
$CURL -X POST "$BASE_URL/api/releases/1/approvals" \
  -H "Content-Type: application/json" \
  -d '{"approver":"user3","approved":true}'
echo ""
echo ""

echo "测试 5.5: 再次查询版本 1.0.0 状态"
$CURL "$BASE_URL/api/releases/1.0.0"
echo ""
echo ""

echo "=== 6. 测试部署 staging ==="
echo "测试 6.1: 部署 1.0.0 到 staging (包含冒烟测试)"
STAGING_RESULT=$($CURL -X POST "$BASE_URL/api/releases/1/deploy/staging" \
  -H "Content-Type: application/json" \
  -d '{"operator":"user1"}')
echo "$STAGING_RESULT"
echo ""

SMOKE_PASS=$(echo "$STAGING_RESULT" | grep -o '"smoke_test_pass":true')
if [ -z "$SMOKE_PASS" ]; then
  echo "⚠️  冒烟测试未通过 (随机失败)，为了继续测试生产部署，让我们再试一次..."
  $CURL -X POST "$BASE_URL/api/releases/1/deploy/staging" \
    -H "Content-Type: application/json" \
    -d '{"operator":"user1"}'
  echo ""
fi
echo ""

echo "=== 7. 测试部署 production ==="
echo "测试 7.1: 部署 1.0.0 到 production"
$CURL -X POST "$BASE_URL/api/releases/1/deploy/production" \
  -H "Content-Type: application/json" \
  -d '{"operator":"user1"}'
echo ""
echo ""

echo "=== 8. 测试灰度发布 ==="
echo "测试 8.1: 部署 1.1.0 到 staging"
$CURL -X POST "$BASE_URL/api/releases/2/submit" \
  -H "Content-Type: application/json" \
  -d '{"operator":"user1"}'
echo ""
$CURL -X POST "$BASE_URL/api/releases/2/approvals" \
  -H "Content-Type: application/json" \
  -d '{"approver":"user2","approved":true}'
echo ""
$CURL -X POST "$BASE_URL/api/releases/2/approvals" \
  -H "Content-Type: application/json" \
  -d '{"approver":"user3","approved":true}'
echo ""
$CURL -X POST "$BASE_URL/api/releases/2/deploy/staging" \
  -H "Content-Type: application/json" \
  -d '{"operator":"user1"}'
echo ""
echo ""

echo "测试 8.2: 灰度发布 1.1.0 到 10% 节点"
$CURL -X POST "$BASE_URL/api/releases/2/deploy/production" \
  -H "Content-Type: application/json" \
  -d '{"operator":"user1","gray_release":true,"gray_percent":10}'
echo ""
echo ""

echo "=== 9. 测试操作历史 ==="
echo "测试 9.1: 查询版本 1 的操作历史"
$CURL "$BASE_URL/api/releases/1/history"
echo ""
echo ""

echo "=== 10. 测试查看所有记录 ==="
echo "测试 10.1: 查看所有 releases"
$CURL "$BASE_URL/api/releases"
echo ""
echo ""

echo "✓ 测试完成！"
