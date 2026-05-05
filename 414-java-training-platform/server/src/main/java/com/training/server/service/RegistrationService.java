package com.training.server.service;

import com.training.common.dto.request.ApproveCancellationRequest;
import com.training.common.dto.request.CancelRegistrationRequest;
import com.training.common.dto.request.EvaluationRequest;
import com.training.common.dto.request.RegisterRequest;
import com.training.common.dto.request.ScoreRequest;
import com.training.common.dto.response.CourseDTO;
import com.training.common.dto.response.RegistrationDTO;
import com.training.common.enums.CourseStatus;
import com.training.common.enums.CourseType;
import com.training.common.enums.ErrorCode;
import com.training.common.enums.RegistrationStatus;
import com.training.server.entity.Course;
import com.training.server.entity.Department;
import com.training.server.entity.Employee;
import com.training.server.entity.Instructor;
import com.training.server.entity.Registration;
import com.training.server.repository.CourseRepository;
import com.training.server.repository.DepartmentRepository;
import com.training.server.repository.EmployeeRepository;
import com.training.server.repository.InstructorRepository;
import com.training.server.repository.RegistrationRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.time.Duration;
import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Service
public class RegistrationService {

    @Autowired
    private RegistrationRepository registrationRepository;

    @Autowired
    private CourseRepository courseRepository;

    @Autowired
    private EmployeeRepository employeeRepository;

    @Autowired
    private InstructorRepository instructorRepository;

    @Autowired
    private DepartmentRepository departmentRepository;

    public RegistrationDTO register(RegisterRequest request) {
        Optional<Employee> employeeOpt = employeeRepository.findById(request.getEmployeeId());
        if (employeeOpt.isEmpty()) {
            throw new RuntimeException(ErrorCode.EMPLOYEE_NOT_FOUND.getMessage());
        }

        Optional<Course> courseOpt = courseRepository.findById(request.getCourseId());
        if (courseOpt.isEmpty()) {
            throw new RuntimeException(ErrorCode.COURSE_NOT_FOUND.getMessage());
        }

        Employee employee = employeeOpt.get();
        Course course = courseOpt.get();

        if (course.getStatus() != CourseStatus.PUBLISHED) {
            throw new RuntimeException("课程未发布，无法报名");
        }

        LocalDateTime now = LocalDateTime.now();
        if (course.getRegistrationDeadline() != null && now.isAfter(course.getRegistrationDeadline())) {
            throw new RuntimeException(ErrorCode.DEADLINE_PASSED.getMessage());
        }

        Optional<Registration> existingOpt = registrationRepository.findByEmployeeIdAndCourseId(
                employee.getId(), course.getId());
        if (existingOpt.isPresent() && existingOpt.get().getStatus() != RegistrationStatus.CANCELLED) {
            throw new RuntimeException(ErrorCode.ALREADY_REGISTERED.getMessage());
        }

        int currentEnrollment = registrationRepository.countByCourseId(course.getId());
        if (currentEnrollment >= course.getMaxCapacity()) {
            throw new RuntimeException(ErrorCode.REGISTRATION_FULL.getMessage());
        }

        if (course.getCourseType() == CourseType.REQUIRED && course.getRequiredDepartments() != null) {
            if (!course.getRequiredDepartments().contains(employee.getDepartmentId())) {
                throw new RuntimeException("该课程是必修课，仅对指定部门开放");
            }
        }

        Registration registration = new Registration();
        registration.setId(UUID.randomUUID().toString());
        registration.setEmployeeId(employee.getId());
        registration.setCourseId(course.getId());

        Registration saved = registrationRepository.save(registration);
        return toDTO(saved);
    }

    public RegistrationDTO requestCancel(CancelRegistrationRequest request) {
        Optional<Registration> registrationOpt = registrationRepository.findById(request.getRegistrationId());
        if (registrationOpt.isEmpty()) {
            throw new RuntimeException(ErrorCode.REGISTRATION_NOT_FOUND.getMessage());
        }

        Registration registration = registrationOpt.get();
        if (registration.getStatus() != RegistrationStatus.REGISTERED) {
            throw new RuntimeException("当前状态不允许取消");
        }

        Optional<Course> courseOpt = courseRepository.findById(registration.getCourseId());
        if (courseOpt.isEmpty()) {
            throw new RuntimeException(ErrorCode.COURSE_NOT_FOUND.getMessage());
        }

        Course course = courseOpt.get();
        LocalDateTime now = LocalDateTime.now();

        if (course.getStartTime() != null && now.isAfter(course.getStartTime())) {
            throw new RuntimeException(ErrorCode.COURSE_ALREADY_STARTED.getMessage());
        }

        if (course.getStartTime() != null) {
            Duration duration = Duration.between(now, course.getStartTime());
            if (duration.toHours() < 24) {
                registration.setStatus(RegistrationStatus.CANCEL_PENDING);
                registration.setCancelReason(request.getReason());
                registration.setCancelRequestedAt(now);
            } else {
                registration.setStatus(RegistrationStatus.CANCELLED);
                registration.setCancelReason(request.getReason());
                registration.setCancelledAt(now);
            }
        } else {
            registration.setStatus(RegistrationStatus.CANCELLED);
            registration.setCancelReason(request.getReason());
            registration.setCancelledAt(now);
        }

        Registration saved = registrationRepository.save(registration);
        return toDTO(saved);
    }

    public RegistrationDTO approveCancellation(ApproveCancellationRequest request) {
        Optional<Registration> registrationOpt = registrationRepository.findById(request.getRegistrationId());
        if (registrationOpt.isEmpty()) {
            throw new RuntimeException(ErrorCode.REGISTRATION_NOT_FOUND.getMessage());
        }

        Registration registration = registrationOpt.get();
        if (registration.getStatus() != RegistrationStatus.CANCEL_PENDING) {
            throw new RuntimeException("取消申请状态不正确");
        }

        if (request.isApproved()) {
            registration.setStatus(RegistrationStatus.CANCELLED);
            registration.setCancelledAt(LocalDateTime.now());
        } else {
            registration.setStatus(RegistrationStatus.REGISTERED);
        }
        registration.setCancellationApprovalComment(request.getComment());

        Registration saved = registrationRepository.save(registration);
        return toDTO(saved);
    }

    public RegistrationDTO submitScore(ScoreRequest request) {
        if (request.getScore() < 0 || request.getScore() > 100) {
            throw new RuntimeException(ErrorCode.INVALID_SCORE.getMessage());
        }

        Optional<Registration> registrationOpt = registrationRepository.findById(request.getRegistrationId());
        if (registrationOpt.isEmpty()) {
            throw new RuntimeException(ErrorCode.REGISTRATION_NOT_FOUND.getMessage());
        }

        Registration registration = registrationOpt.get();
        if (registration.getStatus() != RegistrationStatus.REGISTERED) {
            throw new RuntimeException("当前状态不允许打分");
        }

        Optional<Course> courseOpt = courseRepository.findById(registration.getCourseId());
        if (courseOpt.isEmpty()) {
            throw new RuntimeException(ErrorCode.COURSE_NOT_FOUND.getMessage());
        }

        Course course = courseOpt.get();
        if (course.getStatus() != CourseStatus.COMPLETED) {
            throw new RuntimeException("课程未完成，无法打分");
        }

        registration.setScore(request.getScore());
        registration.setScoreComment(request.getComment());

        if (request.getScore() >= 60) {
            registration.setStatus(RegistrationStatus.COMPLETED);
        } else {
            registration.setStatus(RegistrationStatus.FAILED);
        }
        registration.setCompletedAt(LocalDateTime.now());

        Registration saved = registrationRepository.save(registration);
        return toDTO(saved);
    }

    public RegistrationDTO submitEvaluation(EvaluationRequest request) {
        Optional<Registration> registrationOpt = registrationRepository.findById(request.getRegistrationId());
        if (registrationOpt.isEmpty()) {
            throw new RuntimeException(ErrorCode.REGISTRATION_NOT_FOUND.getMessage());
        }

        Registration registration = registrationOpt.get();
        if (registration.getStatus() != RegistrationStatus.COMPLETED && 
            registration.getStatus() != RegistrationStatus.FAILED) {
            throw new RuntimeException("课程未完成，无法评价");
        }

        if (request.getRating() < 1 || request.getRating() > 5) {
            throw new RuntimeException("评分范围为1-5星");
        }

        registration.setRating(request.getRating());
        registration.setEvaluationComment(request.getComment());
        registration.setEvaluatedAt(LocalDateTime.now());

        Registration saved = registrationRepository.save(registration);
        return toDTO(saved);
    }

    public RegistrationDTO getRegistrationById(String id) {
        Optional<Registration> registrationOpt = registrationRepository.findById(id);
        if (registrationOpt.isEmpty()) {
            throw new RuntimeException(ErrorCode.REGISTRATION_NOT_FOUND.getMessage());
        }
        return toDTO(registrationOpt.get());
    }

    public List<RegistrationDTO> getRegistrationsByEmployee(String employeeId) {
        if (!employeeRepository.existsById(employeeId)) {
            throw new RuntimeException(ErrorCode.EMPLOYEE_NOT_FOUND.getMessage());
        }
        List<Registration> registrations = registrationRepository.findByEmployeeId(employeeId);
        List<RegistrationDTO> dtoList = new ArrayList<>();
        for (Registration registration : registrations) {
            dtoList.add(toDTO(registration));
        }
        return dtoList;
    }

    public List<RegistrationDTO> getRegistrationsByCourse(String courseId) {
        if (!courseRepository.existsById(courseId)) {
            throw new RuntimeException(ErrorCode.COURSE_NOT_FOUND.getMessage());
        }
        List<Registration> registrations = registrationRepository.findByCourseId(courseId);
        List<RegistrationDTO> dtoList = new ArrayList<>();
        for (Registration registration : registrations) {
            dtoList.add(toDTO(registration));
        }
        return dtoList;
    }

    public List<RegistrationDTO> getPendingCancellationsByInstructor(String instructorId) {
        if (!instructorRepository.existsById(instructorId)) {
            throw new RuntimeException(ErrorCode.INSTRUCTOR_NOT_FOUND.getMessage());
        }

        List<Course> courses = courseRepository.findByInstructorId(instructorId);
        List<RegistrationDTO> dtoList = new ArrayList<>();

        for (Course course : courses) {
            List<Registration> registrations = registrationRepository.findByCourseIdAndStatus(
                    course.getId(), RegistrationStatus.CANCEL_PENDING);
            for (Registration registration : registrations) {
                dtoList.add(toDTO(registration));
            }
        }

        return dtoList;
    }

    private RegistrationDTO toDTO(Registration registration) {
        RegistrationDTO dto = new RegistrationDTO();
        dto.setId(registration.getId());
        dto.setEmployeeId(registration.getEmployeeId());
        dto.setCourseId(registration.getCourseId());
        dto.setStatus(registration.getStatus());
        dto.setScore(registration.getScore());
        dto.setScoreComment(registration.getScoreComment());
        dto.setRating(registration.getRating());
        dto.setEvaluationComment(registration.getEvaluationComment());
        dto.setCancelReason(registration.getCancelReason());
        dto.setCancellationApprovalComment(registration.getCancellationApprovalComment());
        dto.setRegisteredAt(registration.getRegisteredAt());
        dto.setCancelledAt(registration.getCancelledAt());
        dto.setCompletedAt(registration.getCompletedAt());
        dto.setEvaluatedAt(registration.getEvaluatedAt());

        if (registration.getScore() != null) {
            dto.setPassed(registration.getScore() >= 60);
        }

        Optional<Employee> employeeOpt = employeeRepository.findById(registration.getEmployeeId());
        if (employeeOpt.isPresent()) {
            Employee employee = employeeOpt.get();
            dto.setEmployeeName(employee.getName());
            dto.setDepartmentId(employee.getDepartmentId());

            Optional<Department> deptOpt = departmentRepository.findById(employee.getDepartmentId());
            if (deptOpt.isPresent()) {
                dto.setDepartmentName(deptOpt.get().getName());
            }
        }

        Optional<Course> courseOpt = courseRepository.findById(registration.getCourseId());
        if (courseOpt.isPresent()) {
            Course course = courseOpt.get();
            dto.setCourseName(course.getName());

            Optional<Instructor> instructorOpt = instructorRepository.findById(course.getInstructorId());
            if (instructorOpt.isPresent()) {
                dto.setInstructorName(instructorOpt.get().getName());
            }
        }

        return dto;
    }
}
