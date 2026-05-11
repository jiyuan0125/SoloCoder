package com.hospital.repository;

import com.hospital.entity.DoctorSchedule;
import com.hospital.enums.DayOfWeek;
import com.hospital.enums.TimeSlot;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

@Repository
public interface DoctorScheduleRepository extends JpaRepository<DoctorSchedule, Long> {
    
    List<DoctorSchedule> findByDoctorId(Long doctorId);
    
    Optional<DoctorSchedule> findByDoctorIdAndDayOfWeekAndTimeSlot(Long doctorId, DayOfWeek dayOfWeek, TimeSlot timeSlot);
    
    List<DoctorSchedule> findByDoctorIdAndDayOfWeek(Long doctorId, DayOfWeek dayOfWeek);
    
    List<DoctorSchedule> findByDoctorIdAndIsAvailableTrue(Long doctorId);
}
