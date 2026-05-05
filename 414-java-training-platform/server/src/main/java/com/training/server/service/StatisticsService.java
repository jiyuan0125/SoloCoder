package com.training.server.service;

import com.training.common.dto.response.CourseEvaluationDTO;
import com.training.common.dto.response.CreditDetailDTO;
import com.training.common.dto.response.CreditSummaryDTO;
import com.training.common.dto.response.DepartmentStatisticsDTO;
import com.training.common.dto.response.InstructorEvaluationSummaryDTO;
import com.training.common.dto.response.MonthlyStatisticsDTO;
import com.training.common.dto.response.StatisticsDTO;
import com.training.common.dto.response.StudentEvaluationDTO;
import com.training.common.enums.CourseStatus;
import com.training.common.enums.CourseType;
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

import java.time.LocalDateTime;
import java.time.Year;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;

@Service
public class StatisticsService {

    @Autowired
    private CourseRepository courseRepository;

    @Autowired
    private RegistrationRepository registrationRepository;

    @Autowired
    private EmployeeRepository employeeRepository;

    @Autowired
    private DepartmentRepository departmentRepository;

    @Autowired
    private InstructorRepository instructorRepository;

    public StatisticsDTO getOverallStatistics() {
        StatisticsDTO dto = new StatisticsDTO();

        List<Course> courses = courseRepository.findAll();
        List<Registration> registrations = registrationRepository.findAll();
        List<Employee> employees = employeeRepository.findAll();

        dto.setTotalCourses(courses.size());
        dto.setTotalEmployees(employees.size());
        dto.setTotalRegistrations(registrations.size());

        int publishedCount = 0;
        int ongoingCount = 0;
        int completedCount = 0;
        int requiredCount = 0;
        int electiveCount = 0;

        for (Course course : courses) {
            if (course.getStatus() == CourseStatus.PUBLISHED) publishedCount++;
            if (course.getStatus() == CourseStatus.ONGOING) ongoingCount++;
            if (course.getStatus() == CourseStatus.COMPLETED) completedCount++;
            if (course.getCourseType() == CourseType.REQUIRED) requiredCount++;
            if (course.getCourseType() == CourseType.ELECTIVE) electiveCount++;
        }

        dto.setPublishedCourses(publishedCount);
        dto.setOngoingCourses(ongoingCount);
        dto.setCompletedCourses(completedCount);
        dto.setRequiredCourses(requiredCount);
        dto.setElectiveCourses(electiveCount);

        int completedRegCount = 0;
        int passedCount = 0;
        int failedCount = 0;

        for (Registration registration : registrations) {
            if (registration.getStatus() == RegistrationStatus.COMPLETED) {
                completedRegCount++;
                passedCount++;
            }
            if (registration.getStatus() == RegistrationStatus.FAILED) {
                completedRegCount++;
                failedCount++;
            }
        }

        dto.setCompletedRegistrations(completedRegCount);
        dto.setPassedRegistrations(passedCount);
        dto.setFailedRegistrations(failedCount);

        if (completedRegCount > 0) {
            dto.setPassRate((double) passedCount / completedRegCount * 100);
        } else {
            dto.setPassRate(0.0);
        }

        Map<String, Integer> courseTypeDistribution = new HashMap<>();
        courseTypeDistribution.put("REQUIRED", requiredCount);
        courseTypeDistribution.put("ELECTIVE", electiveCount);
        dto.setCourseTypeDistribution(courseTypeDistribution);

        dto.setDepartmentStatistics(getDepartmentStatistics());

        return dto;
    }

    public List<DepartmentStatisticsDTO> getDepartmentStatistics() {
        List<DepartmentStatisticsDTO> result = new ArrayList<>();
        List<Department> departments = departmentRepository.findAll();

        for (Department department : departments) {
            DepartmentStatisticsDTO dto = new DepartmentStatisticsDTO();
            dto.setDepartmentId(department.getId());
            dto.setDepartmentName(department.getName());

            List<Employee> employees = employeeRepository.findByDepartmentId(department.getId());
            dto.setEmployeeCount(employees.size());

            int registeredCount = 0;
            int completedCount = 0;
            int passedCount = 0;
            int totalCredits = 0;

            for (Employee employee : employees) {
                List<Registration> registrations = registrationRepository.findByEmployeeId(employee.getId());
                for (Registration registration : registrations) {
                    if (registration.getStatus() == RegistrationStatus.REGISTERED) {
                        registeredCount++;
                    }
                    if (registration.getStatus() == RegistrationStatus.COMPLETED || 
                        registration.getStatus() == RegistrationStatus.FAILED) {
                        completedCount++;
                    }
                    if (registration.getStatus() == RegistrationStatus.COMPLETED) {
                        passedCount++;
                        Optional<Course> courseOpt = courseRepository.findById(registration.getCourseId());
                        if (courseOpt.isPresent()) {
                            totalCredits += courseOpt.get().getCredits();
                        }
                    }
                }
            }

            dto.setRegisteredCount(registeredCount);
            dto.setCompletedCount(completedCount);
            dto.setPassedCount(passedCount);
            dto.setTotalCreditsEarned(totalCredits);

            if (completedCount > 0) {
                dto.setPassRate((double) passedCount / completedCount * 100);
            } else {
                dto.setPassRate(0.0);
            }

            result.add(dto);
        }

        return result;
    }

    public CreditSummaryDTO getCreditSummary(String employeeId, int year) {
        Optional<Employee> employeeOpt = employeeRepository.findById(employeeId);
        if (employeeOpt.isEmpty()) {
            throw new RuntimeException("员工不存在");
        }

        Employee employee = employeeOpt.get();
        CreditSummaryDTO dto = new CreditSummaryDTO();
        dto.setEmployeeId(employee.getId());
        dto.setEmployeeName(employee.getName());
        dto.setDepartmentId(employee.getDepartmentId());
        dto.setYear(year);

        Optional<Department> deptOpt = departmentRepository.findById(employee.getDepartmentId());
        if (deptOpt.isPresent()) {
            dto.setDepartmentName(deptOpt.get().getName());
        }

        List<Registration> registrations = registrationRepository.findByEmployeeId(employeeId);
        List<CreditDetailDTO> creditDetails = new ArrayList<>();

        int totalRequiredCredits = 0;
        int earnedRequiredCredits = 0;
        int totalElectiveCredits = 0;
        int earnedElectiveCredits = 0;

        for (Registration registration : registrations) {
            if (registration.getCompletedAt() == null) continue;
            
            int regYear = registration.getCompletedAt().getYear();
            if (regYear != year) continue;

            Optional<Course> courseOpt = courseRepository.findById(registration.getCourseId());
            if (courseOpt.isEmpty()) continue;

            Course course = courseOpt.get();
            CreditDetailDTO detail = new CreditDetailDTO();
            detail.setCourseId(course.getId());
            detail.setCourseName(course.getName());
            detail.setCourseType(course.getCourseType());
            detail.setCredits(course.getCredits());
            detail.setScore(registration.getScore());
            detail.setPassed(registration.getStatus() == RegistrationStatus.COMPLETED);
            detail.setCompletedAt(registration.getCompletedAt());

            if (course.getCourseType() == CourseType.REQUIRED) {
                totalRequiredCredits += course.getCredits();
                if (detail.isPassed()) {
                    earnedRequiredCredits += course.getCredits();
                }
            } else {
                totalElectiveCredits += course.getCredits();
                if (detail.isPassed()) {
                    earnedElectiveCredits += course.getCredits();
                }
            }

            creditDetails.add(detail);
        }

        dto.setTotalRequiredCredits(totalRequiredCredits);
        dto.setEarnedRequiredCredits(earnedRequiredCredits);
        dto.setTotalElectiveCredits(totalElectiveCredits);
        dto.setEarnedElectiveCredits(earnedElectiveCredits);
        dto.setTotalCredits(earnedRequiredCredits + earnedElectiveCredits);
        dto.setEligibleForExcellentCertificate(earnedElectiveCredits >= 20);
        dto.setCreditDetails(creditDetails);

        return dto;
    }

    public InstructorEvaluationSummaryDTO getInstructorEvaluations(String instructorId) {
        Optional<Instructor> instructorOpt = instructorRepository.findById(instructorId);
        if (instructorOpt.isEmpty()) {
            throw new RuntimeException("讲师不存在");
        }

        Instructor instructor = instructorOpt.get();
        InstructorEvaluationSummaryDTO dto = new InstructorEvaluationSummaryDTO();
        dto.setInstructorId(instructor.getId());
        dto.setInstructorName(instructor.getName());

        List<Course> courses = courseRepository.findByInstructorId(instructorId);
        dto.setTotalCourses(courses.size());

        List<CourseEvaluationDTO> courseEvaluations = new ArrayList<>();
        int totalEvaluations = 0;
        double totalRating = 0.0;

        for (Course course : courses) {
            CourseEvaluationDTO courseEval = new CourseEvaluationDTO();
            courseEval.setCourseId(course.getId());
            courseEval.setCourseName(course.getName());

            List<Registration> registrations = registrationRepository.findByCourseId(course.getId());
            courseEval.setTotalEnrollments(registrations.size());

            List<StudentEvaluationDTO> studentEvaluations = new ArrayList<>();
            int evalCount = 0;
            double courseTotalRating = 0.0;

            for (Registration registration : registrations) {
                if (registration.getRating() != null) {
                    StudentEvaluationDTO studentEval = new StudentEvaluationDTO();
                    studentEval.setEmployeeId(registration.getEmployeeId());

                    Optional<Employee> employeeOpt = employeeRepository.findById(registration.getEmployeeId());
                    if (employeeOpt.isPresent()) {
                        studentEval.setEmployeeName(employeeOpt.get().getName());
                    }

                    studentEval.setRating(registration.getRating());
                    studentEval.setComment(registration.getEvaluationComment());
                    studentEval.setEvaluatedAt(registration.getEvaluatedAt());

                    studentEvaluations.add(studentEval);
                    evalCount++;
                    courseTotalRating += registration.getRating();
                }
            }

            courseEval.setEvaluationCount(evalCount);
            courseEval.setEvaluations(studentEvaluations);

            if (evalCount > 0) {
                courseEval.setAverageRating(courseTotalRating / evalCount);
            } else {
                courseEval.setAverageRating(0.0);
            }

            courseEvaluations.add(courseEval);
            totalEvaluations += evalCount;
            totalRating += courseTotalRating;
        }

        dto.setCourseEvaluations(courseEvaluations);
        dto.setTotalEvaluations(totalEvaluations);

        if (totalEvaluations > 0) {
            dto.setAverageRating(totalRating / totalEvaluations);
        } else {
            dto.setAverageRating(0.0);
        }

        return dto;
    }
}
