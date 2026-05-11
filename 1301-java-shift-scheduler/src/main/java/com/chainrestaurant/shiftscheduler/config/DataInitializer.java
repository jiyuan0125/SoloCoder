package com.chainrestaurant.shiftscheduler.config;

import com.chainrestaurant.shiftscheduler.entity.Employee;
import com.chainrestaurant.shiftscheduler.entity.Shift;
import com.chainrestaurant.shiftscheduler.repository.EmployeeRepository;
import com.chainrestaurant.shiftscheduler.repository.ShiftRepository;
import org.springframework.boot.CommandLineRunner;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import java.time.LocalTime;

@Configuration
public class DataInitializer {
    
    @Bean
    CommandLineRunner initDatabase(EmployeeRepository employeeRepository, ShiftRepository shiftRepository) {
        return args -> {
            if (shiftRepository.count() == 0) {
                shiftRepository.save(new Shift(null, "早班", LocalTime.of(6, 0), LocalTime.of(14, 0), false, 8));
                shiftRepository.save(new Shift(null, "中班", LocalTime.of(10, 0), LocalTime.of(18, 0), false, 8));
                shiftRepository.save(new Shift(null, "晚班", LocalTime.of(18, 0), LocalTime.of(2, 0), true, 8));
                shiftRepository.save(new Shift(null, "夜班", LocalTime.of(22, 0), LocalTime.of(6, 0), true, 8));
                shiftRepository.save(new Shift(null, "正常班", LocalTime.of(8, 0), LocalTime.of(17, 0), false, 9));
                System.out.println("初始班次数据已创建");
            }
            
            if (employeeRepository.count() == 0) {
                employeeRepository.save(new Employee(null, "张三", "EMP001", "北京朝阳门店"));
                employeeRepository.save(new Employee(null, "李四", "EMP002", "北京朝阳门店"));
                employeeRepository.save(new Employee(null, "王五", "EMP003", "上海浦东店"));
                employeeRepository.save(new Employee(null, "赵六", "EMP004", "广州天河店"));
                System.out.println("初始员工数据已创建");
            }
        };
    }
}
