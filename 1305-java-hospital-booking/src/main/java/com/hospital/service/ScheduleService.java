package com.hospital.service;

import com.hospital.dto.ScheduleDTO;
import com.hospital.entity.Doctor;
import com.hospital.entity.DoctorSchedule;
import com.hospital.exception.ResourceNotFoundException;
import com.hospital.repository.DoctorRepository;
import com.hospital.repository.DoctorScheduleRepository;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.stream.Collectors;

@Service
@Transactional(readOnly = true)
public class ScheduleService {

    private final DoctorScheduleRepository scheduleRepository;
    private final DoctorRepository doctorRepository;

    public ScheduleService(DoctorScheduleRepository scheduleRepository, DoctorRepository doctorRepository) {
        this.scheduleRepository = scheduleRepository;
        this.doctorRepository = doctorRepository;
    }

    public List<ScheduleDTO> getSchedulesByDoctor(Long doctorId) {
        if (!doctorRepository.existsById(doctorId)) {
            throw new ResourceNotFoundException("Doctor", doctorId);
        }
        return scheduleRepository.findByDoctorId(doctorId).stream()
                .map(this::toDTO)
                .collect(Collectors.toList());
    }

    public List<ScheduleDTO> getAvailableSchedulesByDoctor(Long doctorId) {
        if (!doctorRepository.existsById(doctorId)) {
            throw new ResourceNotFoundException("Doctor", doctorId);
        }
        return scheduleRepository.findByDoctorIdAndIsAvailableTrue(doctorId).stream()
                .map(this::toDTO)
                .collect(Collectors.toList());
    }

    @Transactional
    public ScheduleDTO createSchedule(ScheduleDTO dto) {
        Doctor doctor = doctorRepository.findById(dto.getDoctorId())
                .orElseThrow(() -> new ResourceNotFoundException("Doctor", dto.getDoctorId()));

        DoctorSchedule schedule = new DoctorSchedule();
        schedule.setDoctor(doctor);
        schedule.setDayOfWeek(dto.getDayOfWeek());
        schedule.setTimeSlot(dto.getTimeSlot());
        schedule.setIsAvailable(dto.getIsAvailable() != null ? dto.getIsAvailable() : true);

        DoctorSchedule saved = scheduleRepository.save(schedule);
        return toDTO(saved);
    }

    @Transactional
    public ScheduleDTO updateSchedule(Long id, ScheduleDTO dto) {
        DoctorSchedule schedule = scheduleRepository.findById(id)
                .orElseThrow(() -> new ResourceNotFoundException("Schedule", id));

        schedule.setIsAvailable(dto.getIsAvailable());

        DoctorSchedule saved = scheduleRepository.save(schedule);
        return toDTO(saved);
    }

    @Transactional
    public void deleteSchedule(Long id) {
        if (!scheduleRepository.existsById(id)) {
            throw new ResourceNotFoundException("Schedule", id);
        }
        scheduleRepository.deleteById(id);
    }

    private ScheduleDTO toDTO(DoctorSchedule schedule) {
        return new ScheduleDTO(
                schedule.getId(),
                schedule.getDoctor().getId(),
                schedule.getDoctor().getName(),
                schedule.getDayOfWeek(),
                schedule.getTimeSlot(),
                schedule.getIsAvailable()
        );
    }
}
