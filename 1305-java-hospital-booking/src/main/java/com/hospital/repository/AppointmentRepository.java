package com.hospital.repository;

import com.hospital.entity.Appointment;
import com.hospital.enums.AppointmentStatus;
import com.hospital.enums.TimeSlot;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.time.LocalDate;
import java.util.List;
import java.util.Optional;

@Repository
public interface AppointmentRepository extends JpaRepository<Appointment, Long> {
    
    List<Appointment> findByDoctorIdAndAppointmentDateAndTimeSlotAndStatusIn(
            Long doctorId, LocalDate appointmentDate, TimeSlot timeSlot, List<AppointmentStatus> statuses);
    
    long countByDoctorIdAndAppointmentDateAndTimeSlotAndStatusIn(
            Long doctorId, LocalDate appointmentDate, TimeSlot timeSlot, List<AppointmentStatus> statuses);
    
    List<Appointment> findByPatientPhoneOrderByBookingTimeDesc(String patientPhone);
    
    List<Appointment> findByPatientNameAndPatientPhoneOrderByBookingTimeDesc(String patientName, String patientPhone);
    
    Optional<Appointment> findByDoctorIdAndPatientPhoneAndAppointmentDateAndStatusIn(
            Long doctorId, String patientPhone, LocalDate appointmentDate, List<AppointmentStatus> statuses);
    
    List<Appointment> findByDoctorIdAndAppointmentDateOrderByTimeSlotAscSerialNumberAsc(
            Long doctorId, LocalDate appointmentDate);
    
    @Query("SELECT MAX(a.serialNumber) FROM Appointment a WHERE a.doctor.id = :doctorId AND a.appointmentDate = :date AND a.timeSlot = :timeSlot")
    Integer findMaxSerialNumber(@Param("doctorId") Long doctorId, @Param("date") LocalDate date, @Param("timeSlot") TimeSlot timeSlot);
    
    List<Appointment> findByAppointmentDateBeforeAndStatus(LocalDate date, AppointmentStatus status);
}
