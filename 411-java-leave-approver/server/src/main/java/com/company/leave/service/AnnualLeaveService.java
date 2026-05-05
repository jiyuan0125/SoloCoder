package com.company.leave.service;

import com.company.leave.entity.Employee;
import org.springframework.stereotype.Service;

import java.time.LocalDate;
import java.time.Month;
import java.time.Period;

@Service
public class AnnualLeaveService {

    public int calculateAnnualLeaveQuota(int yearsOfService) {
        if (yearsOfService < 1) {
            return 0;
        } else if (yearsOfService < 3) {
            return 5;
        } else if (yearsOfService < 10) {
            return 10;
        } else {
            return 15;
        }
    }

    public int calculateYearsOfService(LocalDate joinDate, LocalDate asOfDate) {
        Period period = Period.between(joinDate, asOfDate);
        return period.getYears();
    }

    public void recalculateAnnualLeaveOnNewYear(Employee employee, int year) {
        LocalDate januaryFirst = LocalDate.of(year, Month.JANUARY, 1);
        int yearsOfService = calculateYearsOfService(employee.getJoinDate(), januaryFirst);
        int newQuota = calculateAnnualLeaveQuota(yearsOfService);
        
        int carriedOver = 0;
        LocalDate march31LastYear = LocalDate.of(year - 1, Month.MARCH, 31);
        if (januaryFirst.isBefore(march31LastYear.plusYears(1))) {
            carriedOver = employee.getAnnualLeaveRemaining();
        }
        
        employee.setAnnualLeaveQuota(newQuota);
        employee.setAnnualLeaveRemaining(newQuota);
        employee.setCarriedOverLeave(carriedOver);
    }

    public void clearCarriedOverLeaveIfExpired(Employee employee, LocalDate currentDate) {
        int currentYear = currentDate.getYear();
        LocalDate march31 = LocalDate.of(currentYear, Month.MARCH, 31);
        
        if (currentDate.isAfter(march31)) {
            employee.setCarriedOverLeave(0);
        }
    }

    public int getTotalAvailableAnnualLeave(Employee employee) {
        return employee.getAnnualLeaveRemaining() + employee.getCarriedOverLeave();
    }

    public void deductAnnualLeave(Employee employee, int days) {
        int remaining = employee.getAnnualLeaveRemaining();
        int carriedOver = employee.getCarriedOverLeave();
        
        if (carriedOver >= days) {
            employee.setCarriedOverLeave(carriedOver - days);
        } else {
            int fromCarried = carriedOver;
            int fromCurrent = days - fromCarried;
            employee.setCarriedOverLeave(0);
            employee.setAnnualLeaveRemaining(remaining - fromCurrent);
        }
    }

    public void refundAnnualLeave(Employee employee, int days) {
        int currentQuota = employee.getAnnualLeaveQuota();
        int currentRemaining = employee.getAnnualLeaveRemaining();
        
        int canAddToCurrent = currentQuota - currentRemaining;
        if (days <= canAddToCurrent) {
            employee.setAnnualLeaveRemaining(currentRemaining + days);
        } else {
            employee.setAnnualLeaveRemaining(currentQuota);
            int extra = days - canAddToCurrent;
            employee.setCarriedOverLeave(employee.getCarriedOverLeave() + extra);
        }
    }
}
