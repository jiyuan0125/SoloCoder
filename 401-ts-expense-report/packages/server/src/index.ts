import http from 'http';
import { SERVER_PORT } from '@expense-report/shared';
import { initDefaultUsers } from './store';
import { routeRequest } from './router';

function main(): void {
  initDefaultUsers();

  const server = http.createServer((req, res) => {
    res.setHeader('Access-Control-Allow-Origin', '*');
    res.setHeader('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS');
    res.setHeader('Access-Control-Allow-Headers', 'Content-Type, X-User-Id');

    if (req.method === 'OPTIONS') {
      res.statusCode = 200;
      res.end();
      return;
    }

    routeRequest(req, res);
  });

  server.on('error', (err) => {
    console.error('Server error:', err);
    process.exit(1);
  });

  server.listen(SERVER_PORT, () => {
    console.log(`Expense Report API Server running on http://localhost:${SERVER_PORT}`);
    console.log('');
    console.log('Available endpoints:');
    console.log('  POST   /api/expenses          - 创建报销单');
    console.log('  PUT    /api/expenses          - 更新报销单');
    console.log('  POST   /api/expenses/submit   - 提交报销单');
    console.log('  POST   /api/expenses/approve  - 审批通过');
    console.log('  POST   /api/expenses/reject   - 驳回');
    console.log('  POST   /api/expenses/batch-approve - 批量审批');
    console.log('  POST   /api/expenses/pay      - 打款确认');
    console.log('  POST   /api/expenses/resubmit - 重新提交');
    console.log('  DELETE /api/expenses          - 删除报销单');
    console.log('  GET    /api/expenses/get      - 获取单个报销单 (query: id)');
    console.log('  GET    /api/expenses/list     - 列表 (query: startDate, endDate, status, employeeId, page, pageSize)');
    console.log('  GET    /api/logs/operations    - 操作日志');
    console.log('  GET    /api/logs/audit        - 审计日志');
    console.log('  GET    /api/users/get          - 获取用户 (query: id)');
    console.log('');
    console.log('Default users:');
    console.log('  emp1   - 张三 (员工)');
    console.log('  emp2   - 李四 (员工)');
    console.log('  mgr1   - 王经理 (部门经理)');
    console.log('  mgr2   - 刘经理 (部门经理)');
    console.log('  fd1    - 陈总监 (财务总监)');
    console.log('  admin1 - 系统管理员 (管理员)');
    console.log('');
    console.log('Usage example:');
    console.log('  curl -X POST http://localhost:3000/api/expenses \\');
    console.log('    -H "X-User-Id: emp1" \\');
    console.log('    -H "Content-Type: application/json" \\');
    console.log('    -d \'{"employeeId":"emp1","employeeName":"张三","date":"2026-05-05","amount":50000,"category":"transport","reason":"出差","voucherFileName":"receipt.pdf"}\'');
  });
}

main();
