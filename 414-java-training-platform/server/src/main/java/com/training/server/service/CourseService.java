package com.training.server.service;

import com.training.common.dto.request.CreateCourseRequest;
import com.training.common.dto.request.UpdateCourseRequest;
import com.training.common.dto.response.CourseDTO;
import com.training.common.enums.CourseStatus;
import com.training.common.enums.CourseType;
import com.training.common.enums.ErrorCode;
import com.training.server.entity.Course;
import com.training.server.entity.Instructor;
import com.training.server.repository.CourseRepository;
import com.training.server.repository.InstructorRepository;
import com.training.server.repository.RegistrationRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Service
public class CourseService {

    @Autowired
    private CourseRepository courseRepository;

    @Autowired
    private InstructorRepository instructorRepository;

    @Autowired
    private RegistrationRepository registrationRepository;

    public CourseDTO createCourse(CreateCourseRequest request) {
        if (!instructorRepository.existsById(request.getInstructorId())) {
            throw new RuntimeException(ErrorCode.INSTRUCTOR_NOT_FOUND.getMessage());
        }

        if (request.getStartTime() != null && request.getEndTime() != null 
                && request.getEndTime().isBefore(request.getStartTime())) {
            throw new RuntimeException("课程结束时间不能早于开始时间");
        }

        Course course = new Course();
        course.setId(UUID.randomUUID().toString());
        course.setName(request.getName());
        course.setDescription(request.getDescription());
        course.setInstructorId(request.getInstructorId());
        course.setDurationMinutes(request.getDurationMinutes());
        course.setMaxCapacity(request.getMaxCapacity());
        course.setStartTime(request.getStartTime());
        course.setEndTime(request.getEndTime());
        course.setCourseType(request.getCourseType() != null ? request.getCourseType() : CourseType.ELECTIVE);
        course.setRequiredDepartments(request.getRequiredDepartments());
        course.setCredits(request.getCredits());
        course.setFee(request.getFee());

        if (request.getStartTime() != null) {
            course.setRegistrationDeadline(request.getStartTime().minusHours(1));
        }

        Course saved = courseRepository.save(course);
        return toDTO(saved);
    }

    public CourseDTO getCourseById(String id) {
        Optional<Course> courseOpt = courseRepository.findById(id);
        if (courseOpt.isEmpty()) {
            throw new RuntimeException(ErrorCode.COURSE_NOT_FOUND.getMessage());
        }
        return toDTO(courseOpt.get());
    }

    public List<CourseDTO> getAllCourses() {
        List<Course> courses = courseRepository.findAll();
        List<CourseDTO> dtoList = new ArrayList<>();
        for (Course course : courses) {
            dtoList.add(toDTO(course));
        }
        return dtoList;
    }

    public List<CourseDTO> getPublishedCourses() {
        List<Course> courses = courseRepository.findAll();
        List<CourseDTO> dtoList = new ArrayList<>();
        for (Course course : courses) {
            if (course.getStatus() == CourseStatus.PUBLISHED || course.getStatus() == CourseStatus.ONGOING) {
                dtoList.add(toDTO(course));
            }
        }
        return dtoList;
    }

    public CourseDTO updateCourse(String id, UpdateCourseRequest request) {
        Optional<Course> courseOpt = courseRepository.findById(id);
        if (courseOpt.isEmpty()) {
            throw new RuntimeException(ErrorCode.COURSE_NOT_FOUND.getMessage());
        }

        Course course = courseOpt.get();

        if (request.getInstructorId() != null && !instructorRepository.existsById(request.getInstructorId())) {
            throw new RuntimeException(ErrorCode.INSTRUCTOR_NOT_FOUND.getMessage());
        }

        if (request.getName() != null) {
            course.setName(request.getName());
        }
        if (request.getDescription() != null) {
            course.setDescription(request.getDescription());
        }
        if (request.getInstructorId() != null) {
            course.setInstructorId(request.getInstructorId());
        }
        if (request.getDurationMinutes() != null) {
            course.setDurationMinutes(request.getDurationMinutes());
        }
        if (request.getMaxCapacity() != null) {
            int currentEnrollment = registrationRepository.countByCourseId(course.getId());
            if (request.getMaxCapacity() < currentEnrollment) {
                throw new RuntimeException("名额上限不能小于当前已报名人数");
            }
            course.setMaxCapacity(request.getMaxCapacity());
        }
        if (request.getStartTime() != null) {
            course.setStartTime(request.getStartTime());
            course.setRegistrationDeadline(request.getStartTime().minusHours(1));
        }
        if (request.getEndTime() != null) {
            course.setEndTime(request.getEndTime());
        }
        if (request.getCourseType() != null) {
            course.setCourseType(request.getCourseType());
        }
        if (request.getRequiredDepartments() != null) {
            course.setRequiredDepartments(request.getRequiredDepartments());
        }
        if (request.getStatus() != null) {
            course.setStatus(request.getStatus());
        }
        if (request.getCredits() != null) {
            course.setCredits(request.getCredits());
        }
        if (request.getFee() != null) {
            course.setFee(request.getFee());
        }

        course.setUpdatedAt(LocalDateTime.now());
        Course saved = courseRepository.save(course);
        return toDTO(saved);
    }

    public void deleteCourse(String id) {
        if (!courseRepository.existsById(id)) {
            throw new RuntimeException(ErrorCode.COURSE_NOT_FOUND.getMessage());
        }
        courseRepository.deleteById(id);
    }

    public List<CourseDTO> getCoursesByInstructor(String instructorId) {
        if (!instructorRepository.existsById(instructorId)) {
            throw new RuntimeException(ErrorCode.INSTRUCTOR_NOT_FOUND.getMessage());
        }
        List<Course> courses = courseRepository.findByInstructorId(instructorId);
        List<CourseDTO> dtoList = new ArrayList<>();
        for (Course course : courses) {
            dtoList.add(toDTO(course));
        }
        return dtoList;
    }

    private CourseDTO toDTO(Course course) {
        CourseDTO dto = new CourseDTO();
        dto.setId(course.getId());
        dto.setName(course.getName());
        dto.setDescription(course.getDescription());
        dto.setInstructorId(course.getInstructorId());
        dto.setDurationMinutes(course.getDurationMinutes());
        dto.setMaxCapacity(course.getMaxCapacity());
        dto.setCurrentEnrollment(registrationRepository.countByCourseId(course.getId()));
        dto.setStartTime(course.getStartTime());
        dto.setEndTime(course.getEndTime());
        dto.setRegistrationDeadline(course.getRegistrationDeadline());
        dto.setCourseType(course.getCourseType());
        dto.setStatus(course.getStatus());
        dto.setRequiredDepartments(course.getRequiredDepartments());
        dto.setCredits(course.getCredits());
        dto.setFee(course.getFee());
        dto.setCreatedAt(course.getCreatedAt());
        dto.setUpdatedAt(course.getUpdatedAt());

        Optional<Instructor> instructorOpt = instructorRepository.findById(course.getInstructorId());
        if (instructorOpt.isPresent()) {
            dto.setInstructorName(instructorOpt.get().getName());
        }

        return dto;
    }
}
