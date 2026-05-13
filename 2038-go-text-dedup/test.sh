#!/bin/bash

BASE_URL="http://localhost:8080"

echo "=== 测试1: 空数组 ==="
RESPONSE=$(curl -s --max-time 5 -X POST "$BASE_URL/dedup" \
  -H "Content-Type: application/json" \
  -d '{"texts": []}')
echo "响应: $RESPONSE"
JOB_ID=$(echo "$RESPONSE" | grep -o '"job_id":"[^"]*"' | cut -d'"' -f4)
echo "Job ID: $JOB_ID"
sleep 1
echo "结果:"
curl -s --max-time 5 "$BASE_URL/result/$JOB_ID"
echo -e "\n"

echo "=== 测试2: 完全相同的文本 ==="
RESPONSE=$(curl -s --max-time 5 -X POST "$BASE_URL/dedup" \
  -H "Content-Type: application/json" \
  -d '{"texts": ["hello world", "hello world", "hello world"]}')
echo "响应: $RESPONSE"
JOB_ID=$(echo "$RESPONSE" | grep -o '"job_id":"[^"]*"' | cut -d'"' -f4)
echo "Job ID: $JOB_ID"
sleep 1
echo "结果:"
curl -s --max-time 5 "$BASE_URL/result/$JOB_ID"
echo -e "\n"

echo "=== 测试3: 大小写不同 ==="
RESPONSE=$(curl -s --max-time 5 -X POST "$BASE_URL/dedup" \
  -H "Content-Type: application/json" \
  -d '{"texts": ["Hello World", "HELLO WORLD", "hello world"]}')
echo "响应: $RESPONSE"
JOB_ID=$(echo "$RESPONSE" | grep -o '"job_id":"[^"]*"' | cut -d'"' -f4)
echo "Job ID: $JOB_ID"
sleep 1
echo "结果:"
curl -s --max-time 5 "$BASE_URL/result/$JOB_ID"
echo -e "\n"

echo "=== 测试4: 空格数量不同 ==="
RESPONSE=$(curl -s --max-time 5 -X POST "$BASE_URL/dedup" \
  -H "Content-Type: application/json" \
  -d '{"texts": ["hello   world", "hello world", "hello     world"]}')
echo "响应: $RESPONSE"
JOB_ID=$(echo "$RESPONSE" | grep -o '"job_id":"[^"]*"' | cut -d'"' -f4)
echo "Job ID: $JOB_ID"
sleep 1
echo "结果:"
curl -s --max-time 5 "$BASE_URL/result/$JOB_ID"
echo -e "\n"

echo "=== 测试5: 全半角标点 ==="
RESPONSE=$(curl -s --max-time 5 -X POST "$BASE_URL/dedup" \
  -H "Content-Type: application/json" \
  -d '{"texts": ["你好世界", "你好世界。", "你好世界."]}')
echo "响应: $RESPONSE"
JOB_ID=$(echo "$RESPONSE" | grep -o '"job_id":"[^"]*"' | cut -d'"' -f4)
echo "Job ID: $JOB_ID"
sleep 1
echo "结果:"
curl -s --max-time 5 "$BASE_URL/result/$JOB_ID"
echo -e "\n"

echo "=== 测试6: 空文本和空白文本 ==="
RESPONSE=$(curl -s --max-time 5 -X POST "$BASE_URL/dedup" \
  -H "Content-Type: application/json" \
  -d '{"texts": ["", "   ", "\t\n", "hello"]}')
echo "响应: $RESPONSE"
JOB_ID=$(echo "$RESPONSE" | grep -o '"job_id":"[^"]*"' | cut -d'"' -f4)
echo "Job ID: $JOB_ID"
sleep 1
echo "结果:"
curl -s --max-time 5 "$BASE_URL/result/$JOB_ID"
echo -e "\n"

echo "=== 测试7: 综合测试 ==="
RESPONSE=$(curl -s --max-time 5 -X POST "$BASE_URL/dedup" \
  -H "Content-Type: application/json" \
  -d '{"texts": ["你好世界", "你好 世界。", "HELLO WORLD", "hello   world", "", "   ", "测试文本", "测试文本!"]}')
echo "响应: $RESPONSE"
JOB_ID=$(echo "$RESPONSE" | grep -o '"job_id":"[^"]*"' | cut -d'"' -f4)
echo "Job ID: $JOB_ID"
sleep 1
echo "结果:"
curl -s --max-time 5 "$BASE_URL/result/$JOB_ID"
echo -e "\n"

echo "=== 测试完成 ==="
