package com.training.server.controller;

import com.training.common.dto.request.ApproveCancellationRequest;
import com.training.common.dto.request.CancelRegistrationRequest;
import com.training.common.dto.request.EvaluationRequest;
import com.training.common.dto.request.RegisterRequest;
import com.training.common.dto.request.ScoreRequest;
import com.training.common.response.ApiResponse;
import com.training.common.dto.response.RegistrationDTO;
import com.training.common.enums.ErrorCode;
import com.training.server.service.RegistrationService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@RestController
@RequestMapping("/api/registrations")
public class RegistrationController {

    @Autowired
    private RegistrationService registrationService;

    @PostMapping("/register")
    public ApiResponse<RegistrationDTO> register(@RequestBody RegisterRequest request) {
        try {
            RegistrationDTO registration = registrationService.register(request);
            return ApiResponse.success(registration);
        } catch (RuntimeException e) {
            return ApiResponse.error(ErrorCode.INVALID_PARAM, e.getMessage());
        }
    }

    @PostMapping("/cancel")
    public ApiResponse<RegistrationDTO> requestCancel(@RequestBody CancelRegistrationRequest request) {
        try {
            RegistrationDTO registration = registrationService.requestCancel(request);
            return ApiResponse.success(registration);
        } catch (RuntimeException e) {
            return ApiResponse.error(ErrorCode.INVALID_PARAM, e.getMessage());
        }
    }

    @PostMapping("/approve-cancellation")
    public ApiResponse<RegistrationDTO> approveCancellation(@RequestBody ApproveCancellationRequest request) {
        try {
            RegistrationDTO registration = registrationService.approveCancellation(request);
            return ApiResponse.success(registration);
        } catch (RuntimeException e) {
            return ApiResponse.error(ErrorCode.INVALID_PARAM, e.getMessage());
        }
    }

    @PostMapping("/score")
    public ApiResponse<RegistrationDTO> submitScore(@RequestBody ScoreRequest request) {
        try {
            RegistrationDTO registration = registrationService.submitScore(request);
            return ApiResponse.success(registration);
        } catch (RuntimeException e) {
            return ApiResponse.error(ErrorCode.INVALID_PARAM, e.getMessage());
        }
    }

    @PostMapping("/evaluate")
    public ApiResponse<RegistrationDTO> submitEvaluation(@RequestBody EvaluationRequest request) {
        try {
            RegistrationDTO registration = registrationService.submitEvaluation(request);
            return ApiResponse.success(registration);
        } catch (RuntimeException e) {
            return ApiResponse.error(ErrorCode.INVALID_PARAM, e.getMessage());
        }
    }

    @GetMapping("/{id}")
    public ApiResponse<RegistrationDTO> getRegistrationById(@PathVariable String id) {
        try {
            RegistrationDTO registration = registrationService.getRegistrationById(id);
            return ApiResponse.success(registration);
        } catch (RuntimeException e) {
            return ApiResponse.error(ErrorCode.REGISTRATION_NOT_FOUND);
        }
    }

    @GetMapping(params = "employeeId")
    public ApiResponse<List<RegistrationDTO>> getRegistrationsByEmployee(@RequestParam String employeeId) {
        try {
            List<RegistrationDTO> registrations = registrationService.getRegistrationsByEmployee(employeeId);
            return ApiResponse.success(registrations);
        } catch (RuntimeException e) {
            return ApiResponse.error(ErrorCode.EMPLOYEE_NOT_FOUND);
        }
    }

    @GetMapping(params = "courseId")
    public ApiResponse<List<RegistrationDTO>> getRegistrationsByCourse(@RequestParam String courseId) {
        try {
            List<RegistrationDTO> registrations = registrationService.getRegistrationsByCourse(courseId);
            return ApiResponse.success(registrations);
        } catch (RuntimeException e) {
            return ApiResponse.error(ErrorCode.COURSE_NOT_FOUND);
        }
    }

    @GetMapping(value = "/pending-cancellations", params = "instructorId")
    public ApiResponse<List<RegistrationDTO>> getPendingCancellationsByInstructor(@RequestParam String instructorId) {
        try {
            List<RegistrationDTO> registrations = registrationService.getPendingCancellationsByInstructor(instructorId);
            return ApiResponse.success(registrations);
        } catch (RuntimeException e) {
            return ApiResponse.error(ErrorCode.INSTRUCTOR_NOT_FOUND);
        }
    }
}
