package com.company.leave.scheduler;

import com.company.leave.entity.Employee;
import com.company.leave.service.AnnualLeaveService;
import com.company.leave.service.EmployeeService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

import java.time.LocalDate;
import java.time.Month;
import java.util.List;

@Component
public class AnnualLeaveScheduler {
    private static final Logger logger = LoggerFactory.getLogger(AnnualLeaveScheduler.class);
    
    private final EmployeeService employeeService;
    private final AnnualLeaveService annualLeaveService;

    public AnnualLeaveScheduler(EmployeeService employeeService, AnnualLeaveService annualLeaveService) {
        this.employeeService = employeeService;
        this.annualLeaveService = annualLeaveService;
    }

    @Scheduled(cron = "0 0 0 1 1 *")
    public void recalculateAnnualLeaveOnNewYear() {
        int year = LocalDate.now().getYear();
        logger.info("开始执行年度年假重算任务，年份: {}", year);
        
        List<Employee> employees = employeeService.getAllEmployees();
        for (Employee employee : employees) {
            annualLeaveService.recalculateAnnualLeaveOnNewYear(employee, year);
            employeeService.saveEmployee(employee);
            logger.info("员工 {} 年假已重算，新配额: {}天", employee.getName(), employee.getAnnualLeaveQuota());
        }
        
        logger.info("年度年假重算任务完成");
    }

    @Scheduled(cron = "0 0 0 1 4 *")
    public void clearCarriedOverLeave() {
        logger.info("开始执行结转年假清零任务");
        
        LocalDate now = LocalDate.now();
        List<Employee> employees = employeeService.getAllEmployees();
        for (Employee employee : employees) {
            int carriedOver = employee.getCarriedOverLeave();
            if (carriedOver > 0) {
                employee.setCarriedOverLeave(0);
                employeeService.saveEmployee(employee);
                logger.info("员工 {} 结转年假已清零: {}天", employee.getName(), carriedOver);
            }
        }
        
        logger.info("结转年假清零任务完成");
    }
}
