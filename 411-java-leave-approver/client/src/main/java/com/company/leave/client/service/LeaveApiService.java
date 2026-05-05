package com.company.leave.client.service;

import com.company.leave.client.http.HttpClient;
import com.company.leave.dto.*;

import java.io.IOException;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class LeaveApiService {
    private final HttpClient httpClient;

    public LeaveApiService(HttpClient httpClient) {
        this.httpClient = httpClient;
    }

    public ApiResponse<EmployeeDTO> createEmployee(String name, Long managerId, String joinDate) throws IOException, InterruptedException {
        Map<String, Object> request = new HashMap<>();
        request.put("name", name);
        request.put("managerId", managerId);
        request.put("joinDate", joinDate);
        return httpClient.post("/api/employees", request, ApiResponse.class);
    }

    public ApiResponse<EmployeeDTO> getEmployee(Long id) throws IOException, InterruptedException {
        return httpClient.get("/api/employees/" + id, ApiResponse.class);
    }

    public ApiResponse<List> getAllEmployees() throws IOException, InterruptedException {
        return httpClient.get("/api/employees", ApiResponse.class);
    }

    public ApiResponse<List> getTeamMembers(Long managerId) throws IOException, InterruptedException {
        return httpClient.get("/api/employees/manager/" + managerId + "/team", ApiResponse.class);
    }

    public ApiResponse<LeaveRecordDTO> createLeave(CreateLeaveRequest request) throws IOException, InterruptedException {
        return httpClient.post("/api/leaves", request, ApiResponse.class);
    }

    public ApiResponse<LeaveRecordDTO> approveLeave(Long leaveId, Long managerId, boolean approved, String comment) throws IOException, InterruptedException {
        ApproveLeaveRequest request = new ApproveLeaveRequest();
        request.setLeaveId(leaveId);
        request.setManagerId(managerId);
        request.setApproved(approved);
        request.setComment(comment);
        return httpClient.post("/api/leaves/approve", request, ApiResponse.class);
    }

    public ApiResponse<LeaveRecordDTO> getLeave(Long id) throws IOException, InterruptedException {
        return httpClient.get("/api/leaves/" + id, ApiResponse.class);
    }

    public ApiResponse<List> getEmployeeLeaves(Long employeeId) throws IOException, InterruptedException {
        return httpClient.get("/api/leaves/employee/" + employeeId, ApiResponse.class);
    }

    public ApiResponse<List> getAllLeaves() throws IOException, InterruptedException {
        return httpClient.get("/api/leaves", ApiResponse.class);
    }

    public ApiResponse<CalendarViewDTO> getTeamCalendar(Long managerId, int year, int month) throws IOException, InterruptedException {
        String path = String.format("/api/calendar/team/%d?year=%d&month=%d", managerId, year, month);
        return httpClient.get(path, ApiResponse.class);
    }
}
