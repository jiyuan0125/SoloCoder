export function showHelp(): void {
  const help = `
薪资计算器 CLI

用法: salary-cli <命令> [选项]

命令:
  calculate  计算薪资
  save       保存薪资记录（需先计算）
  query      查询薪资记录
  comparison 查询同比环比对比

选项:
  --server <url>  服务器地址 (默认: http://127.0.0.1:3000)
  --help, -h       显示帮助信息

calculate 命令选项:
  --employeeId <id>              员工ID
  --employeeName <name>           员工姓名
  --month <YYYY-MM>               月份
  --baseSalary <分>               基本工资（整数分）
  --positionAllowance <分>        岗位津贴（整数分）
  --performanceCoefficient <0-2>  绩效系数
  --socialInsuranceRate <0-1>     社保比例
  --housingFundRate <0-1>         公积金比例
  --taxThreshold <分>              个税起征点
  --leaveDeduction <分>            请假扣款
  --overtimeWeekday <小时>         工作日加班小时
  --overtimeWeekend <小时>         周末加班小时
  --overtimeHoliday <小时>         法定节假日加班小时
  --yearEndBonus <分>              年终奖

query 命令选项:
  --employeeId <id>  按员工ID筛选
  --month <YYYY-MM>  按月份筛选

comparison 命令选项:
  --employeeId <id>  员工ID
  --month <YYYY-MM>  月份

示例:
  salary-cli calculate --employeeId E001 --employeeName "张三" --month 2024-01 \
    --baseSalary 1000000 --positionAllowance 200000 --performanceCoefficient 1.0 \
    --socialInsuranceRate 0.105 --housingFundRate 0.12 --taxThreshold 500000 \
    --leaveDeduction 0 --overtimeWeekday 10 --overtimeWeekend 8 --overtimeHoliday 0 \
    --yearEndBonus 0

  salary-cli query --employeeId E001
  salary-cli query --month 2024-01
  salary-cli comparison --employeeId E001 --month 2024-01
`;
  console.log(help);
}
