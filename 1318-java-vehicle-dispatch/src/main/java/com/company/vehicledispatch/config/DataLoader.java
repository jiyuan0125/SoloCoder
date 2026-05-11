package com.company.vehicledispatch.config;

import com.company.vehicledispatch.entity.Department;
import com.company.vehicledispatch.entity.Vehicle;
import com.company.vehicledispatch.enums.FuelType;
import com.company.vehicledispatch.enums.VehicleStatus;
import com.company.vehicledispatch.repository.DepartmentRepository;
import com.company.vehicledispatch.repository.VehicleRepository;
import com.company.vehicledispatch.service.VehicleService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.CommandLineRunner;
import org.springframework.stereotype.Component;

import java.time.LocalDate;

@Component
public class DataLoader implements CommandLineRunner {

    @Autowired
    private DepartmentRepository departmentRepository;

    @Autowired
    private VehicleRepository vehicleRepository;

    @Autowired
    private VehicleService vehicleService;

    @Override
    public void run(String... args) {
        if (departmentRepository.count() == 0) {
            loadDepartments();
        }
        if (vehicleRepository.count() == 0) {
            loadVehicles();
        }
    }

    private void loadDepartments() {
        String[][] departments = {
                {"技术部", "负责公司技术研发"},
                {"市场部", "负责公司市场推广"},
                {"销售部", "负责公司产品销售"},
                {"人力资源部", "负责公司人力资源管理"},
                {"财务部", "负责公司财务管理"},
                {"行政部", "负责公司行政事务"}
        };

        for (String[] dept : departments) {
            Department department = new Department();
            department.setName(dept[0]);
            department.setDescription(dept[1]);
            departmentRepository.save(department);
        }
    }

    private void loadVehicles() {
        Object[][] vehicles = {
                {"京A12345", "丰田凯美瑞", 5, FuelType.GASOLINE, 50.0, 60.0, 15000.0, 10000.0},
                {"京B67890", "大众帕萨特", 5, FuelType.GASOLINE, 15.0, 60.0, 25000.0, 20000.0},
                {"京C11111", "别克GL8", 7, FuelType.GASOLINE, 70.0, 80.0, 30000.0, 25000.0},
                {"京D22222", "本田雅阁", 5, FuelType.GASOLINE, 3.0, 60.0, 18000.0, 15000.0},
                {"京E33333", "特斯拉Model 3", 5, FuelType.ELECTRIC, 400.0, 500.0, 8000.0, 5000.0},
                {"京F44444", "奥迪A6L", 5, FuelType.GASOLINE, 45.0, 70.0, 35000.0, 30000.0},
                {"京G55555", "奔驰E级", 5, FuelType.GASOLINE, 20.0, 65.0, 42000.0, 40000.0},
                {"京H66666", "宝马5系", 5, FuelType.GASOLINE, 55.0, 68.0, 12000.0, 10000.0},
                {"京I77777", "比亚迪汉EV", 5, FuelType.ELECTRIC, 80.0, 600.0, 5000.0, 5000.0},
                {"京J88888", "大众迈腾", 5, FuelType.GASOLINE, 12.0, 66.0, 28000.0, 25000.0}
        };

        LocalDate today = LocalDate.now();
        LocalDate insuranceExpiry = today.plusMonths(6);
        LocalDate inspectionExpiry = today.plusMonths(8);

        for (Object[] v : vehicles) {
            Vehicle vehicle = new Vehicle();
            vehicle.setPlateNumber((String) v[0]);
            vehicle.setModel((String) v[1]);
            vehicle.setSeats((Integer) v[2]);
            vehicle.setFuelType((FuelType) v[3]);
            vehicle.setFuelLevel((Double) v[4]);
            vehicle.setMaxFuelLevel((Double) v[5]);
            vehicle.setCurrentMileage((Double) v[6]);
            vehicle.setLastMaintenanceMileage((Double) v[7]);
            vehicle.setInsuranceExpiryDate(insuranceExpiry);
            vehicle.setAnnualInspectionExpiryDate(inspectionExpiry);
            vehicle.setStatus(VehicleStatus.AVAILABLE);
            vehicle.setAverageFuelConsumption(8.0);

            vehicleService.updateVehicleStatuses(vehicle);
            vehicleRepository.save(vehicle);
        }
    }
}
