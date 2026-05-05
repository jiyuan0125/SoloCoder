package com.training.client.service;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.training.client.http.HttpClient;
import com.training.common.dto.request.ApproveCancellationRequest;
import com.training.common.dto.request.CancelRegistrationRequest;
import com.training.common.dto.request.CreateCourseRequest;
import com.training.common.dto.request.CreateDepartmentRequest;
import com.training.common.dto.request.CreateEmployeeRequest;
import com.training.common.dto.request.CreateInstructorRequest;
import com.training.common.dto.request.EvaluationRequest;
import com.training.common.dto.request.RegisterRequest;
import com.training.common.dto.request.ScoreRequest;
import com.training.common.dto.request.UpdateCourseRequest;
import com.training.common.response.ApiResponse;
import com.training.common.dto.response.CourseDTO;
import com.training.common.dto.response.CreditSummaryDTO;
import com.training.common.dto.response.DepartmentDTO;
import com.training.common.dto.response.EmployeeDTO;
import com.training.common.dto.response.InstructorDTO;
import com.training.common.dto.response.InstructorEvaluationSummaryDTO;
import com.training.common.dto.response.RegistrationDTO;
import com.training.common.dto.response.StatisticsDTO;

import java.io.IOException;
import java.util.List;

public class TrainingApiService {

    private final HttpClient httpClient;
    private final ObjectMapper objectMapper;

    public TrainingApiService(String baseUrl) {
        this.httpClient = new HttpClient(baseUrl);
        this.objectMapper = httpClient.getObjectMapper();
    }

    public DepartmentDTO createDepartment(CreateDepartmentRequest request) throws IOException, InterruptedException {
        ApiResponse<DepartmentDTO> response = httpClient.post("/api/departments", request, 
                new TypeReference<ApiResponse<DepartmentDTO>>() {});
        return response.getData();
    }

    public DepartmentDTO getDepartment(String id) throws IOException, InterruptedException {
        ApiResponse<DepartmentDTO> response = httpClient.get("/api/departments/" + id, 
                new TypeReference<ApiResponse<DepartmentDTO>>() {});
        return response.getData();
    }

    public List<DepartmentDTO> getAllDepartments() throws IOException, InterruptedException {
        ApiResponse<List<DepartmentDTO>> response = httpClient.get("/api/departments", 
                new TypeReference<ApiResponse<List<DepartmentDTO>>>() {});
        return response.getData();
    }

    public EmployeeDTO createEmployee(CreateEmployeeRequest request) throws IOException, InterruptedException {
        ApiResponse<EmployeeDTO> response = httpClient.post("/api/employees", request, 
                new TypeReference<ApiResponse<EmployeeDTO>>() {});
        return response.getData();
    }

    public EmployeeDTO getEmployee(String id) throws IOException, InterruptedException {
        ApiResponse<EmployeeDTO> response = httpClient.get("/api/employees/" + id, 
                new TypeReference<ApiResponse<EmployeeDTO>>() {});
        return response.getData();
    }

    public List<EmployeeDTO> getAllEmployees() throws IOException, InterruptedException {
        ApiResponse<List<EmployeeDTO>> response = httpClient.get("/api/employees", 
                new TypeReference<ApiResponse<List<EmployeeDTO>>>() {});
        return response.getData();
    }

    public InstructorDTO createInstructor(CreateInstructorRequest request) throws IOException, InterruptedException {
        ApiResponse<InstructorDTO> response = httpClient.post("/api/instructors", request, 
                new TypeReference<ApiResponse<InstructorDTO>>() {});
        return response.getData();
    }

    public InstructorDTO getInstructor(String id) throws IOException, InterruptedException {
        ApiResponse<InstructorDTO> response = httpClient.get("/api/instructors/" + id, 
                new TypeReference<ApiResponse<InstructorDTO>>() {});
        return response.getData();
    }

    public List<InstructorDTO> getAllInstructors() throws IOException, InterruptedException {
        ApiResponse<List<InstructorDTO>> response = httpClient.get("/api/instructors", 
                new TypeReference<ApiResponse<List<InstructorDTO>>>() {});
        return response.getData();
    }

    public CourseDTO createCourse(CreateCourseRequest request) throws IOException, InterruptedException {
        ApiResponse<CourseDTO> response = httpClient.post("/api/courses", request, 
                new TypeReference<ApiResponse<CourseDTO>>() {});
        return response.getData();
    }

    public CourseDTO getCourse(String id) throws IOException, InterruptedException {
        ApiResponse<CourseDTO> response = httpClient.get("/api/courses/" + id, 
                new TypeReference<ApiResponse<CourseDTO>>() {});
        return response.getData();
    }

    public List<CourseDTO> getAllCourses(boolean publishedOnly) throws IOException, InterruptedException {
        String path = publishedOnly ? "/api/courses?publishedOnly=true" : "/api/courses";
        ApiResponse<List<CourseDTO>> response = httpClient.get(path, 
                new TypeReference<ApiResponse<List<CourseDTO>>>() {});
        return response.getData();
    }

    public CourseDTO updateCourse(String id, UpdateCourseRequest request) throws IOException, InterruptedException {
        ApiResponse<CourseDTO> response = httpClient.put("/api/courses/" + id, request, 
                new TypeReference<ApiResponse<CourseDTO>>() {});
        return response.getData();
    }

    public RegistrationDTO register(RegisterRequest request) throws IOException, InterruptedException {
        ApiResponse<RegistrationDTO> response = httpClient.post("/api/registrations/register", request, 
                new TypeReference<ApiResponse<RegistrationDTO>>() {});
        return response.getData();
    }

    public RegistrationDTO cancelRegistration(CancelRegistrationRequest request) throws IOException, InterruptedException {
        ApiResponse<RegistrationDTO> response = httpClient.post("/api/registrations/cancel", request, 
                new TypeReference<ApiResponse<RegistrationDTO>>() {});
        return response.getData();
    }

    public RegistrationDTO approveCancellation(ApproveCancellationRequest request) throws IOException, InterruptedException {
        ApiResponse<RegistrationDTO> response = httpClient.post("/api/registrations/approve-cancellation", request, 
                new TypeReference<ApiResponse<RegistrationDTO>>() {});
        return response.getData();
    }

    public RegistrationDTO submitScore(ScoreRequest request) throws IOException, InterruptedException {
        ApiResponse<RegistrationDTO> response = httpClient.post("/api/registrations/score", request, 
                new TypeReference<ApiResponse<RegistrationDTO>>() {});
        return response.getData();
    }

    public RegistrationDTO submitEvaluation(EvaluationRequest request) throws IOException, InterruptedException {
        ApiResponse<RegistrationDTO> response = httpClient.post("/api/registrations/evaluate", request, 
                new TypeReference<ApiResponse<RegistrationDTO>>() {});
        return response.getData();
    }

    public RegistrationDTO getRegistration(String id) throws IOException, InterruptedException {
        ApiResponse<RegistrationDTO> response = httpClient.get("/api/registrations/" + id, 
                new TypeReference<ApiResponse<RegistrationDTO>>() {});
        return response.getData();
    }

    public List<RegistrationDTO> getRegistrationsByEmployee(String employeeId) throws IOException, InterruptedException {
        ApiResponse<List<RegistrationDTO>> response = httpClient.get("/api/registrations?employeeId=" + employeeId, 
                new TypeReference<ApiResponse<List<RegistrationDTO>>>() {});
        return response.getData();
    }

    public List<RegistrationDTO> getRegistrationsByCourse(String courseId) throws IOException, InterruptedException {
        ApiResponse<List<RegistrationDTO>> response = httpClient.get("/api/registrations?courseId=" + courseId, 
                new TypeReference<ApiResponse<List<RegistrationDTO>>>() {});
        return response.getData();
    }

    public StatisticsDTO getStatistics() throws IOException, InterruptedException {
        ApiResponse<StatisticsDTO> response = httpClient.get("/api/statistics", 
                new TypeReference<ApiResponse<StatisticsDTO>>() {});
        return response.getData();
    }

    public CreditSummaryDTO getCreditSummary(String employeeId, Integer year) throws IOException, InterruptedException {
        String path = "/api/statistics/credits/" + employeeId;
        if (year != null) {
            path += "?year=" + year;
        }
        ApiResponse<CreditSummaryDTO> response = httpClient.get(path, 
                new TypeReference<ApiResponse<CreditSummaryDTO>>() {});
        return response.getData();
    }

    public InstructorEvaluationSummaryDTO getInstructorEvaluations(String instructorId) throws IOException, InterruptedException {
        ApiResponse<InstructorEvaluationSummaryDTO> response = httpClient.get("/api/statistics/evaluations/" + instructorId, 
                new TypeReference<ApiResponse<InstructorEvaluationSummaryDTO>>() {});
        return response.getData();
    }

    public List<RegistrationDTO> getPendingCancellations(String instructorId) throws IOException, InterruptedException {
        ApiResponse<List<RegistrationDTO>> response = httpClient.get(
                "/api/registrations/pending-cancellations?instructorId=" + instructorId, 
                new TypeReference<ApiResponse<List<RegistrationDTO>>>() {});
        return response.getData();
    }
}
