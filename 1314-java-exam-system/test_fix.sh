#!/bin/bash

BASE_URL="http://localhost:8888/api"

echo "=== 测试考试系统修复 ==="

echo ""
echo "1. 创建试卷..."
CREATE_PAPER_RESPONSE=$(curl -s -X POST "$BASE_URL/exam-papers" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Java基础测试",
    "description": "测试自动判分和重考功能",
    "totalScore": 85,
    "passScore": 51,
    "durationMinutes": 30,
    "canRetake": true,
    "maxRetakeCount": 5,
    "rules": [
      {
        "categoryId": 1,
        "questionType": "SINGLE_CHOICE",
        "questionCount": 3,
        "scorePerQuestion": 15,
        "minDifficultyLevel": 1,
        "maxDifficultyLevel": 5
      },
      {
        "categoryId": 1,
        "questionType": "MULTIPLE_CHOICE",
        "questionCount": 2,
        "scorePerQuestion": 20,
        "minDifficultyLevel": 1,
        "maxDifficultyLevel": 5
      }
    ]
  }')

echo "$CREATE_PAPER_RESPONSE"

PAPER_ID=$(echo "$CREATE_PAPER_RESPONSE" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
echo ""
echo "试卷ID: $PAPER_ID"

if [ -z "$PAPER_ID" ]; then
    echo "试卷创建失败，退出测试"
    exit 1
fi

echo ""
echo "2. 开始第一次考试 (userId=3)..."
START_RESPONSE=$(curl -s -X POST "$BASE_URL/exams/start" \
  -H "Content-Type: application/json" \
  -d "{
    \"examPaperId\": $PAPER_ID,
    \"userId\": 3
  }")
echo "$START_RESPONSE"

EXAM_RECORD_ID=$(echo "$START_RESPONSE" | grep -o '"examRecordId":[0-9]*' | cut -d':' -f2)
echo ""
echo "考试记录ID: $EXAM_RECORD_ID"

if [ -z "$EXAM_RECORD_ID" ]; then
    echo "考试开始失败，退出测试"
    exit 1
fi

echo ""
echo "3. 获取考试题目..."
EXAM_DETAIL=$(curl -s "$BASE_URL/exams/$EXAM_RECORD_ID")
echo "$EXAM_DETAIL"

echo ""
echo "4. 提取题目ID并提交答案..."
QUESTION_IDS=$(echo "$EXAM_DETAIL" | grep -o '"examQuestionId":[0-9]*' | cut -d':' -f2)

count=0
for qid in $QUESTION_IDS; do
    count=$((count+1))
    echo "提交题目 $count 答案 (qid=$qid)..."
    ANSWER_RESPONSE=$(curl -s -X POST "$BASE_URL/exams/answer" \
      -H "Content-Type: application/json" \
      -d "{
        \"examRecordId\": $EXAM_RECORD_ID,
        \"examQuestionId\": $qid,
        \"userAnswer\": \"A\"
      }")
    echo "$ANSWER_RESPONSE"
    echo ""
done

echo ""
echo "5. 提交考试..."
SUBMIT_RESPONSE=$(curl -s -X POST "$BASE_URL/exams/$EXAM_RECORD_ID/submit")
echo "$SUBMIT_RESPONSE"

echo ""
echo "6. 获取考试报告..."
REPORT_RESPONSE=$(curl -s "$BASE_URL/exams/$EXAM_RECORD_ID/report")
echo "$REPORT_RESPONSE"

echo ""
echo "检查报告中的关键字段..."
ANSWERED=$(echo "$REPORT_RESPONSE" | grep -o '"answeredQuestions":[0-9]*' | cut -d':' -f2)
SCORE=$(echo "$REPORT_RESPONSE" | grep -o '"obtainedScore":[0-9]*' | head -1 | cut -d':' -f2)

echo "已答题数: $ANSWERED"
echo "得分: $SCORE"

if [ "$ANSWERED" = "5" ]; then
    echo "✓ 判分功能正常工作！"
else
    echo "✗ 判分功能可能仍有问题"
fi

echo ""
echo "7. 测试重考功能 (userId=3)..."
RETAKE_RESPONSE=$(curl -s -X POST "$BASE_URL/exams/start" \
  -H "Content-Type: application/json" \
  -d "{
    \"examPaperId\": $PAPER_ID,
    \"userId\": 3
  }")
echo "$RETAKE_RESPONSE"

if echo "$RETAKE_RESPONSE" | grep -q "不允许重考\|已达到最大重考次数"; then
    echo ""
    echo "✗ 重考功能仍然有问题！"
else
    RETAKE_EXAM_ID=$(echo "$RETAKE_RESPONSE" | grep -o '"examRecordId":[0-9]*' | cut -d':' -f2)
    if [ -n "$RETAKE_EXAM_ID" ]; then
        echo ""
        echo "✓ 重考成功！新的考试记录ID: $RETAKE_EXAM_ID"
    fi
fi

echo ""
echo "=== 测试完成 ==="
