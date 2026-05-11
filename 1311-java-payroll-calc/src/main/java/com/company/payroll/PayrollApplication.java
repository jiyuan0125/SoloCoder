package com.company.payroll;

import com.company.payroll.model.*;
import com.company.payroll.service.*;
import com.company.payroll.service.DepartmentSummaryService;
import com.company.payroll.service.ExportService;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.time.YearMonth;
import java.util.List;

public class PayrollApplication {

    public static void main(String[] args) {
        System.out.println("员工薪资计算系统");
        System.out.println("==============================");

        PayrollDataStore dataStore = new PayrollDataStore();
        PayrollCalculatorService calculatorService = new PayrollCalculatorService(dataStore);
        DepartmentSummaryService summaryService = new DepartmentSummaryService(dataStore);
        ExportService exportService = new ExportService();

        setupDemoData(dataStore);

        System.out.println("\n--- 示例1：计算普通员工1月份薪资 ---");
        SalaryInput janInput = createJanuaryInput();
        PaySlip janSlip = calculatorService.calculatePaySlip(janInput);
        System.out.println(janSlip.formatPaySlip());

        System.out.println("\n--- 示例2：计算同一名员工2月份薪资（累计预扣） ---");
        SalaryInput febInput = createFebruaryInput();
        PaySlip febSlip = calculatorService.calculatePaySlip(febInput);
        System.out.println(febSlip.formatPaySlip());

        System.out.println("\n--- 示例3：计算高收入员工（超过96万速算扣除数上限） ---");
        SalaryInput highIncomeInput = createHighIncomeInput();
        PaySlip highIncomeSlip = calculatorService.calculatePaySlip(highIncomeInput);
        System.out.println(highIncomeSlip.formatPaySlip());

        System.out.println("\n--- 示例4：年中入职员工 ---");
        SalaryInput midYearInput = createMidYearHireInput(dataStore);
        PaySlip midYearSlip = calculatorService.calculatePaySlip(midYearInput);
        System.out.println(midYearSlip.formatPaySlip());

        System.out.println("\n--- 示例5：年终奖计算（两种计税方式对比） ---");
        SalaryInput bonusInput = createYearEndBonusInput(dataStore);
        PaySlip bonusSlip = calculatorService.calculatePaySlip(bonusInput);
        System.out.println(bonusSlip.formatPaySlip());

        System.out.println("\n--- 示例6：试用期员工（80%薪资，但社保按全额算） ---");
        SalaryInput probationInput = createProbationInput(dataStore);
        PaySlip probationSlip = calculatorService.calculatePaySlip(probationInput);
        System.out.println(probationSlip.formatPaySlip());

        System.out.println("\n--- 示例7：离职员工按实际工作天数折算 ---");
        SalaryInput terminatedInput = createTerminatedInput(dataStore);
        PaySlip terminatedSlip = calculatorService.calculatePaySlip(terminatedInput);
        System.out.println(terminatedSlip.formatPaySlip());

        System.out.println("\n--- 示例8：部门薪资汇总 ---");
        DepartmentSalarySummary summary = summaryService.calculateDepartmentSummary("DEPT001", 2026, 1);
        System.out.println(summaryService.formatSummary(summary));

        System.out.println("\n--- 示例9：按员工查询历史工资条 ---");
        List<PaySlip> history = dataStore.getPaySlipsByEmployee("EMP001");
        for (PaySlip slip : history) {
            System.out.println(slip.getYear() + "年" + slip.getMonth() + "月: 税前=" + 
                    slip.getGrossSalary() + ", 个税=" + slip.getCurrentMonthTax() + ", 实发=" + slip.getNetSalary());
        }

        System.out.println("\n系统演示完成！");
    }

    private static void setupDemoData(PayrollDataStore dataStore) {
        Department techDept = new Department("DEPT001", "技术部");
        Department hrDept = new Department("DEPT002", "人力资源部");
        dataStore.addDepartment(techDept);
        dataStore.addDepartment(hrDept);

        Employee emp1 = new Employee("EMP001", "张三", "DEPT001", new BigDecimal("15000"), 
                false, LocalDate.of(2025, 1, 1));
        emp1.addSpecialDeduction(SpecialDeductionType.CHILDREN_EDUCATION);
        emp1.addSpecialDeduction(SpecialDeductionType.HOUSING_LOAN_INTEREST);
        dataStore.addEmployee(emp1);

        Employee emp2 = new Employee("EMP002", "李四", "DEPT001", new BigDecimal("25000"), 
                false, LocalDate.of(2024, 6, 1));
        emp2.addSpecialDeduction(SpecialDeductionType.ELDERLY_SUPPORT);
        emp2.addSpecialDeduction(SpecialDeductionType.HOUSING_RENT);
        dataStore.addEmployee(emp2);

        Employee emp3 = new Employee("EMP003", "王五", "DEPT002", new BigDecimal("12000"), 
                true, LocalDate.of(2026, 3, 1));
        dataStore.addEmployee(emp3);

        Employee emp4 = new Employee("EMP004", "赵六", "DEPT001", new BigDecimal("100000"), 
                false, LocalDate.of(2020, 1, 1));
        emp4.addSpecialDeduction(SpecialDeductionType.CHILDREN_EDUCATION);
        emp4.addSpecialDeduction(SpecialDeductionType.ELDERLY_SUPPORT);
        emp4.addSpecialDeduction(SpecialDeductionType.HOUSING_LOAN_INTEREST);
        dataStore.addEmployee(emp4);
    }

    private static SalaryInput createJanuaryInput() {
        SalaryInput input = new SalaryInput("EMP001", 2026, 1, new BigDecimal("15000"));
        input.getAdditionalIncomes().add(new AdditionalIncome("AI001", "绩效奖金", new BigDecimal("3000"), IncomeType.PERFORMANCE_BONUS));
        input.getAdditionalIncomes().add(new AdditionalIncome("AI002", "加班费", new BigDecimal("500"), IncomeType.OVERTIME_PAY));
        input.getAdditionalIncomes().add(new AdditionalIncome("AI003", "交通补贴", new BigDecimal("800"), IncomeType.TRANSPORTATION_ALLOWANCE));
        input.getAdditionalIncomes().add(new AdditionalIncome("AI004", "通讯补贴", new BigDecimal("200"), IncomeType.COMMUNICATION_ALLOWANCE));
        input.getAdditionalIncomes().add(new AdditionalIncome("AI005", "餐补", new BigDecimal("600"), IncomeType.MEAL_ALLOWANCE));
        return input;
    }

    private static SalaryInput createFebruaryInput() {
        SalaryInput input = new SalaryInput("EMP001", 2026, 2, new BigDecimal("15000"));
        input.getAdditionalIncomes().add(new AdditionalIncome("AI006", "绩效奖金", new BigDecimal("2500"), IncomeType.PERFORMANCE_BONUS));
        input.getAdditionalIncomes().add(new AdditionalIncome("AI007", "项目奖金", new BigDecimal("5000"), IncomeType.PROJECT_BONUS));
        return input;
    }

    private static SalaryInput createHighIncomeInput() {
        SalaryInput input = new SalaryInput("EMP004", 2026, 12, new BigDecimal("100000"));
        input.getAdditionalIncomes().add(new AdditionalIncome("AI008", "绩效奖金", new BigDecimal("50000"), IncomeType.PERFORMANCE_BONUS));
        return input;
    }

    private static SalaryInput createMidYearHireInput(PayrollDataStore dataStore) {
        Employee emp = new Employee("EMP005", "钱七", "DEPT002", new BigDecimal("18000"), 
                false, LocalDate.of(2026, 4, 1));
        emp.addSpecialDeduction(SpecialDeductionType.HOUSING_RENT);
        dataStore.addEmployee(emp);

        SalaryInput input = new SalaryInput("EMP005", 2026, 6, new BigDecimal("18000"));
        return input;
    }

    private static SalaryInput createYearEndBonusInput(PayrollDataStore dataStore) {
        SalaryInput input = new SalaryInput("EMP002", 2026, 12, new BigDecimal("25000"));
        input.getAdditionalIncomes().add(new AdditionalIncome("AI009", "年终奖", new BigDecimal("60000"), IncomeType.YEAR_END_BONUS));
        return input;
    }

    private static SalaryInput createProbationInput(PayrollDataStore dataStore) {
        SalaryInput input = new SalaryInput("EMP003", 2026, 5, new BigDecimal("12000"));
        return input;
    }

    private static SalaryInput createTerminatedInput(PayrollDataStore dataStore) {
        Employee emp = new Employee("EMP006", "孙八", "DEPT001", new BigDecimal("20000"), 
                false, LocalDate.of(2025, 1, 1));
        emp.setTerminationDate(LocalDate.of(2026, 3, 10));
        dataStore.addEmployee(emp);

        SalaryInput input = new SalaryInput("EMP006", 2026, 3, new BigDecimal("20000"));
        input.setActualWorkDays(8);
        input.setTotalWorkDaysInMonth(21);
        return input;
    }
}
