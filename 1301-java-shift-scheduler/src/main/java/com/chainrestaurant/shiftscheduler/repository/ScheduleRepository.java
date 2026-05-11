package com.chainrestaurant.shiftscheduler.repository;

import com.chainrestaurant.shiftscheduler.entity.Schedule;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.time.LocalDate;
import java.util.List;
import java.util.Optional;

@Repository
public interface ScheduleRepository extends JpaRepository<Schedule, Long> {
    
    Optional<Schedule> findByEmployeeIdAndScheduleDate(Long employeeId, LocalDate scheduleDate);
    
    List<Schedule> findByEmployeeIdAndScheduleDateBetween(Long employeeId, LocalDate startDate, LocalDate endDate);
    
    List<Schedule> findByScheduleDateBetweenOrderByScheduleDateAsc(LocalDate startDate, LocalDate endDate);
    
    @Query("SELECT s FROM Schedule s JOIN FETCH s.employee e WHERE e.storeName = :storeName AND s.scheduleDate BETWEEN :startDate AND :endDate ORDER BY s.scheduleDate ASC")
    List<Schedule> findByStoreNameAndDateRange(@Param("storeName") String storeName, 
                                                  @Param("startDate") LocalDate startDate, 
                                                  @Param("endDate") LocalDate endDate);
    
    @Query("SELECT s FROM Schedule s JOIN FETCH s.employee e JOIN FETCH s.shift sh WHERE s.employee.id = :employeeId AND s.scheduleDate BETWEEN :startDate AND :endDate ORDER BY s.scheduleDate ASC")
    List<Schedule> findByEmployeeIdAndDateRangeWithShift(@Param("employeeId") Long employeeId, 
                                                          @Param("startDate") LocalDate startDate, 
                                                          @Param("endDate") LocalDate endDate);
    
    boolean existsByEmployeeIdAndScheduleDate(Long employeeId, LocalDate scheduleDate);
}
