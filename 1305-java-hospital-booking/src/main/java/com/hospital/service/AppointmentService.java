package com.hospital.service;

import com.hospital.dto.AppointmentDTO;
import com.hospital.dto.BookingRequestDTO;
import com.hospital.dto.SlotAvailabilityDTO;
import com.hospital.entity.Appointment;
import com.hospital.entity.Doctor;
import com.hospital.entity.DoctorSchedule;
import com.hospital.enums.AppointmentStatus;
import com.hospital.enums.DayOfWeek;
import com.hospital.enums.TimeSlot;
import com.hospital.exception.BusinessException;
import com.hospital.exception.ResourceNotFoundException;
import com.hospital.repository.AppointmentRepository;
import com.hospital.repository.DoctorRepository;
import com.hospital.repository.DoctorScheduleRepository;
import com.hospital.util.DateUtil;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.Arrays;
import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;

@Service
@Transactional(readOnly = true)
public class AppointmentService {

    private final AppointmentRepository appointmentRepository;
    private final DoctorRepository doctorRepository;
    private final DoctorScheduleRepository scheduleRepository;

    private static final List<AppointmentStatus> ACTIVE_STATUSES = Arrays.asList(AppointmentStatus.PENDING);

    public AppointmentService(AppointmentRepository appointmentRepository,
                              DoctorRepository doctorRepository,
                              DoctorScheduleRepository scheduleRepository) {
        this.appointmentRepository = appointmentRepository;
        this.doctorRepository = doctorRepository;
        this.scheduleRepository = scheduleRepository;
    }

    @Transactional
    public AppointmentDTO bookAppointment(BookingRequestDTO request) {
        validateBookingRequest(request);

        Doctor doctor = doctorRepository.findById(request.getDoctorId())
                .orElseThrow(() -> new ResourceNotFoundException("Doctor", request.getDoctorId()));

        LocalDate appointmentDate = request.getAppointmentDate();
        if (appointmentDate.isBefore(LocalDate.now())) {
            throw new BusinessException("不能预约过去的日期");
        }

        DayOfWeek dayOfWeek = DateUtil.getDayOfWeek(appointmentDate);

        DoctorSchedule schedule = scheduleRepository
                .findByDoctorIdAndDayOfWeekAndTimeSlot(doctor.getId(), dayOfWeek, request.getTimeSlot())
                .orElseThrow(() -> new BusinessException("该医生在预约日期的该时段不上班"));

        if (!Boolean.TRUE.equals(schedule.getIsAvailable())) {
            throw new BusinessException("该医生在预约日期的该时段不上班");
        }

        int totalSlots = getTotalSlots(doctor, request.getTimeSlot());
        long bookedCount = appointmentRepository.countByDoctorIdAndAppointmentDateAndTimeSlotAndStatusIn(
                doctor.getId(), appointmentDate, request.getTimeSlot(), ACTIVE_STATUSES);

        if (bookedCount >= totalSlots) {
            throw new BusinessException("该时段号源已满");
        }

        Optional<Appointment> existingAppointment = appointmentRepository
                .findByDoctorIdAndPatientPhoneAndAppointmentDateAndStatusIn(
                        doctor.getId(), request.getPatientPhone(), appointmentDate, ACTIVE_STATUSES);

        if (existingAppointment.isPresent()) {
            throw new BusinessException("您今天已经预约过这位医生，如需改时间请先取消再重新预约");
        }

        Integer maxSerial = appointmentRepository.findMaxSerialNumber(
                doctor.getId(), appointmentDate, request.getTimeSlot());
        int serialNumber = (maxSerial == null ? 0 : maxSerial) + 1;

        Appointment appointment = new Appointment();
        appointment.setDoctor(doctor);
        appointment.setAppointmentDate(appointmentDate);
        appointment.setTimeSlot(request.getTimeSlot());
        appointment.setSerialNumber(serialNumber);
        appointment.setPatientName(request.getPatientName());
        appointment.setPatientPhone(request.getPatientPhone());
        appointment.setPatientIdCard(request.getPatientIdCard());
        appointment.setStatus(AppointmentStatus.PENDING);

        Appointment saved = appointmentRepository.save(appointment);
        return toDTO(saved);
    }

    @Transactional
    public AppointmentDTO cancelAppointment(Long appointmentId) {
        Appointment appointment = appointmentRepository.findById(appointmentId)
                .orElseThrow(() -> new ResourceNotFoundException("Appointment", appointmentId));

        if (appointment.getStatus() == AppointmentStatus.CANCELLED) {
            throw new BusinessException("该挂号记录已取消，不能重复取消");
        }

        if (appointment.getStatus() == AppointmentStatus.EXPIRED) {
            throw new BusinessException("该挂号记录已过期");
        }

        appointment.setStatus(AppointmentStatus.CANCELLED);
        appointment.setCancelTime(LocalDateTime.now());

        Appointment saved = appointmentRepository.save(appointment);
        return toDTO(saved);
    }

    public SlotAvailabilityDTO checkSlotAvailability(Long doctorId, LocalDate date, TimeSlot timeSlot) {
        Doctor doctor = doctorRepository.findById(doctorId)
                .orElseThrow(() -> new ResourceNotFoundException("Doctor", doctorId));

        DayOfWeek dayOfWeek = DateUtil.getDayOfWeek(date);

        Optional<DoctorSchedule> scheduleOpt = scheduleRepository
                .findByDoctorIdAndDayOfWeekAndTimeSlot(doctorId, dayOfWeek, timeSlot);

        boolean isDoctorAvailable = scheduleOpt
                .map(s -> Boolean.TRUE.equals(s.getIsAvailable()))
                .orElse(false);

        int totalSlots = getTotalSlots(doctor, timeSlot);
        long bookedCount = 0;
        int remainingSlots = 0;
        boolean isAvailable = false;

        if (isDoctorAvailable) {
            bookedCount = appointmentRepository.countByDoctorIdAndAppointmentDateAndTimeSlotAndStatusIn(
                    doctorId, date, timeSlot, ACTIVE_STATUSES);
            remainingSlots = totalSlots - (int) bookedCount;
            isAvailable = remainingSlots > 0 && !date.isBefore(LocalDate.now());
        }

        return new SlotAvailabilityDTO(
                doctorId,
                doctor.getName(),
                date,
                timeSlot,
                totalSlots,
                (int) bookedCount,
                remainingSlots,
                isAvailable
        );
    }

    public List<AppointmentDTO> getPatientAppointments(String patientPhone) {
        List<Appointment> appointments = appointmentRepository
                .findByPatientPhoneOrderByBookingTimeDesc(patientPhone);
        return appointments.stream()
                .map(this::toDTOWithExpiryCheck)
                .collect(Collectors.toList());
    }

    public List<AppointmentDTO> getPatientAppointmentsByNameAndPhone(String patientName, String patientPhone) {
        List<Appointment> appointments = appointmentRepository
                .findByPatientNameAndPatientPhoneOrderByBookingTimeDesc(patientName, patientPhone);
        return appointments.stream()
                .map(this::toDTOWithExpiryCheck)
                .collect(Collectors.toList());
    }

    public List<AppointmentDTO> getDoctorDailyAppointments(Long doctorId, LocalDate date) {
        List<Appointment> appointments = appointmentRepository
                .findByDoctorIdAndAppointmentDateOrderByTimeSlotAscSerialNumberAsc(doctorId, date);
        return appointments.stream()
                .map(this::toDTOWithExpiryCheck)
                .collect(Collectors.toList());
    }

    public AppointmentDTO getAppointmentById(Long id) {
        Appointment appointment = appointmentRepository.findById(id)
                .orElseThrow(() -> new ResourceNotFoundException("Appointment", id));
        return toDTOWithExpiryCheck(appointment);
    }

    private int getTotalSlots(Doctor doctor, TimeSlot timeSlot) {
        return switch (timeSlot) {
            case MORNING -> doctor.getTitle().getMorningSlots();
            case AFTERNOON -> doctor.getTitle().getAfternoonSlots();
        };
    }

    private void validateBookingRequest(BookingRequestDTO request) {
        if (request.getDoctorId() == null) {
            throw new BusinessException("请选择医生");
        }
        if (request.getAppointmentDate() == null) {
            throw new BusinessException("请选择预约日期");
        }
        if (request.getTimeSlot() == null) {
            throw new BusinessException("请选择时段");
        }
        if (request.getPatientName() == null || request.getPatientName().trim().isEmpty()) {
            throw new BusinessException("请输入患者姓名");
        }
        if (request.getPatientPhone() == null || request.getPatientPhone().trim().isEmpty()) {
            throw new BusinessException("请输入联系电话");
        }
    }

    private AppointmentDTO toDTO(Appointment appointment) {
        Doctor doctor = appointment.getDoctor();
        return new AppointmentDTO(
                appointment.getId(),
                doctor.getId(),
                doctor.getName(),
                doctor.getTitle().name(),
                doctor.getDepartment().getId(),
                doctor.getDepartment().getName(),
                appointment.getAppointmentDate(),
                appointment.getTimeSlot(),
                appointment.getSerialNumber(),
                appointment.getPatientName(),
                appointment.getPatientPhone(),
                appointment.getPatientIdCard(),
                appointment.getStatus(),
                appointment.getBookingTime(),
                appointment.getCancelTime()
        );
    }

    private AppointmentDTO toDTOWithExpiryCheck(Appointment appointment) {
        if (appointment.getStatus() == AppointmentStatus.PENDING 
                && appointment.getAppointmentDate().isBefore(LocalDate.now())) {
            appointment.setStatus(AppointmentStatus.EXPIRED);
            appointmentRepository.save(appointment);
        }
        return toDTO(appointment);
    }
}
