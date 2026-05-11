package com.company.expense.repository;

import com.company.expense.entity.Employee;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.Optional;

@Repository
public interface EmployeeRepository extends JpaRepository<Employee, Long> {

    Optional<Employee> findByEmployeeId(String employeeId);

    @Query("SELECT e FROM Employee e WHERE e.isGeneralManager = true")
    Optional<Employee> findGeneralManager();

    @Query("SELECT e FROM Employee e WHERE e.isManager = true AND e.department = :department")
    Optional<Employee> findManagerByDepartment(@Param("department") String department);
}
