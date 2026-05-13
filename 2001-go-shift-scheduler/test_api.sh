#!/bin/bash

BASE_URL="http://localhost:8081"

echo "=== 排班管理系统测试 ==="
echo ""

echo "1. 测试基本数据创建..."
curl -s -X POST "$BASE_URL/api/employees" -H "Content-Type: application/json" -d '{"name":"员工A","department_id":1}' | python3 -m json.tool
echo ""

echo "2. 测试创建排班（周一 9:00-18:00，9小时）..."
RESP=$(curl -s -X POST "$BASE_URL/api/shifts" -H "Content-Type: application/json" -d '{"employee_id":1,"shift_date":"2026-05-11","start_time":"09:00","end_time":"18:00","position":"岗位1"}')
echo "$RESP" | python3 -m json.tool
echo ""
SHIFT_ID1=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])")

echo "3. 测试同一天重复排班（应返回409）..."
curl -s -w "\nHTTP_CODE=%{http_code}\n" -X POST "$BASE_URL/api/shifts" -H "Content-Type: application/json" -d '{"employee_id":1,"shift_date":"2026-05-11","start_time":"14:00","end_time":"22:00","position":"岗位1"}'
echo ""

echo "4. 测试11小时间隔规则（周一22:00下班，周二最早9:00上班）..."
RESP=$(curl -s -X POST "$BASE_URL/api/shifts" -H "Content-Type: application/json" -d '{"employee_id":2,"shift_date":"2026-05-10","start_time":"14:00","end_time":"22:00","position":"岗位2"}')
echo "员工2 5月10日 14:00-22:00: $RESP"

echo "测试5月11日 8:00上班（间隔10小时，应失败）..."
curl -s -w "\nHTTP_CODE=%{http_code}\n" -X POST "$BASE_URL/api/shifts" -H "Content-Type: application/json" -d '{"employee_id":2,"shift_date":"2026-05-11","start_time":"08:00","end_time":"17:00","position":"岗位2"}'

echo "测试5月11日 9:00上班（间隔11小时，应成功）..."
curl -s -w "\nHTTP_CODE=%{http_code}\n" -X POST "$BASE_URL/api/shifts" -H "Content-Type: application/json" -d '{"employee_id":2,"shift_date":"2026-05-11","start_time":"09:00","end_time":"18:00","position":"岗位2"}'
echo ""

echo "5. 测试周工时40小时上限..."
echo "为员工3排满40小时..."
curl -s -X POST "$BASE_URL/api/employees" -H "Content-Type: application/json" -d '{"name":"员工C","department_id":1}' > /dev/null

for day in 11 12 13 14; do
  curl -s -X POST "$BASE_URL/api/shifts" -H "Content-Type: application/json" -d "{\"employee_id\":4,\"shift_date\":\"2026-05-${day}\",\"start_time\":\"09:00\",\"end_time\":\"19:00\",\"position\":\"岗位3\"}" > /dev/null
  echo "5月${day}日 9:00-19:00（10小时）已排"
done

echo "尝试再排1小时（40+1=41>40，应失败）..."
curl -s -w "\nHTTP_CODE=%{http_code}\n" -X POST "$BASE_URL/api/shifts" -H "Content-Type: application/json" -d '{"employee_id":4,"shift_date":"2026-05-15","start_time":"09:00","end_time":"10:00","position":"岗位3"}'
echo ""

echo "6. 测试流程状态推进（待提交→审核中→已通过→执行中→已完成）..."
echo "当前排班状态:"
curl -s "$BASE_URL/api/shifts/$SHIFT_ID1" | python3 -m json.tool

echo "推进到审核中..."
curl -s -X POST "$BASE_URL/api/shifts/$SHIFT_ID1/advance" -H "Content-Type: application/json" -d '{"operator_id":1,"note":"提交审核"}'
echo ""

echo "推进到已通过..."
curl -s -X POST "$BASE_URL/api/shifts/$SHIFT_ID1/advance" -H "Content-Type: application/json" -d '{"operator_id":1,"note":"审核通过"}'
echo ""

echo "推进到执行中..."
curl -s -X POST "$BASE_URL/api/shifts/$SHIFT_ID1/advance" -H "Content-Type: application/json" -d '{"operator_id":1,"note":"开始执行"}'
echo ""

echo "推进到已完成..."
curl -s -X POST "$BASE_URL/api/shifts/$SHIFT_ID1/advance" -H "Content-Type: application/json" -d '{"operator_id":1,"note":"执行完成"}'
echo ""

echo "最终状态:"
curl -s "$BASE_URL/api/shifts/$SHIFT_ID1" | python3 -m json.tool
echo ""

echo "7. 测试操作历史..."
echo "查看排班 $SHIFT_ID1 的操作历史:"
curl -s "$BASE_URL/api/shifts/$SHIFT_ID1/history" | python3 -m json.tool
echo ""

echo "8. 测试节假日双倍工时..."
echo "添加节假日 5月1日:"
curl -s -X POST "$BASE_URL/api/holidays" -H "Content-Type: application/json" -d '{"date":"2026-05-01","name":"劳动节"}' | python3 -m json.tool
echo ""

echo "=== 测试完成 ==="
