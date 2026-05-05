package com.training.server.service;

import com.training.common.dto.request.CreateInstructorRequest;
import com.training.common.dto.response.InstructorDTO;
import com.training.common.enums.ErrorCode;
import com.training.server.entity.Instructor;
import com.training.server.repository.InstructorRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Service
public class InstructorService {

    @Autowired
    private InstructorRepository instructorRepository;

    public InstructorDTO createInstructor(CreateInstructorRequest request) {
        Instructor instructor = new Instructor();
        instructor.setId(UUID.randomUUID().toString());
        instructor.setName(request.getName());
        instructor.setEmail(request.getEmail());
        instructor.setSpecialties(request.getSpecialties());
        instructor.setHourlyRate(request.getHourlyRate());

        Instructor saved = instructorRepository.save(instructor);
        return toDTO(saved);
    }

    public InstructorDTO getInstructorById(String id) {
        Optional<Instructor> instructorOpt = instructorRepository.findById(id);
        if (instructorOpt.isEmpty()) {
            throw new RuntimeException(ErrorCode.INSTRUCTOR_NOT_FOUND.getMessage());
        }
        return toDTO(instructorOpt.get());
    }

    public List<InstructorDTO> getAllInstructors() {
        List<Instructor> instructors = instructorRepository.findAll();
        List<InstructorDTO> dtoList = new ArrayList<>();
        for (Instructor instructor : instructors) {
            dtoList.add(toDTO(instructor));
        }
        return dtoList;
    }

    public InstructorDTO updateInstructor(String id, CreateInstructorRequest request) {
        Optional<Instructor> instructorOpt = instructorRepository.findById(id);
        if (instructorOpt.isEmpty()) {
            throw new RuntimeException(ErrorCode.INSTRUCTOR_NOT_FOUND.getMessage());
        }

        Instructor instructor = instructorOpt.get();
        if (request.getName() != null) {
            instructor.setName(request.getName());
        }
        if (request.getEmail() != null) {
            instructor.setEmail(request.getEmail());
        }
        if (request.getSpecialties() != null) {
            instructor.setSpecialties(request.getSpecialties());
        }
        if (request.getHourlyRate() != null) {
            instructor.setHourlyRate(request.getHourlyRate());
        }

        Instructor saved = instructorRepository.save(instructor);
        return toDTO(saved);
    }

    public void deleteInstructor(String id) {
        if (!instructorRepository.existsById(id)) {
            throw new RuntimeException(ErrorCode.INSTRUCTOR_NOT_FOUND.getMessage());
        }
        instructorRepository.deleteById(id);
    }

    private InstructorDTO toDTO(Instructor instructor) {
        InstructorDTO dto = new InstructorDTO();
        dto.setId(instructor.getId());
        dto.setName(instructor.getName());
        dto.setEmail(instructor.getEmail());
        dto.setSpecialties(instructor.getSpecialties());
        dto.setHourlyRate(instructor.getHourlyRate());
        dto.setActive(instructor.isActive());
        dto.setCreatedAt(instructor.getCreatedAt());
        return dto;
    }
}
