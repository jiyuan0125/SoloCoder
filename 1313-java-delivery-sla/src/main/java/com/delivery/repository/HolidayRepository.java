package com.delivery.repository;

import com.delivery.entity.Holiday;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;
import java.time.LocalDate;
import java.util.Optional;

@Repository
public interface HolidayRepository extends JpaRepository<Holiday, Long> {
    Optional<Holiday> findByHolidayDate(LocalDate date);
    
    @Query("SELECT h FROM Holiday h WHERE h.holidayDate = :date")
    Optional<Holiday> checkHoliday(@Param("date") LocalDate date);
}
