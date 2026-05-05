package com.training.server.controller;

import com.training.common.dto.request.CreateInstructorRequest;
import com.training.common.response.ApiResponse;
import com.training.common.dto.response.InstructorDTO;
import com.training.common.enums.ErrorCode;
import com.training.server.service.InstructorService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@RestController
@RequestMapping("/api/instructors")
public class InstructorController {

    @Autowired
    private InstructorService instructorService;

    @PostMapping
    public ApiResponse<InstructorDTO> createInstructor(@RequestBody CreateInstructorRequest request) {
        InstructorDTO instructor = instructorService.createInstructor(request);
        return ApiResponse.success(instructor);
    }

    @GetMapping("/{id}")
    public ApiResponse<InstructorDTO> getInstructorById(@PathVariable String id) {
        try {
            InstructorDTO instructor = instructorService.getInstructorById(id);
            return ApiResponse.success(instructor);
        } catch (RuntimeException e) {
            return ApiResponse.error(ErrorCode.INSTRUCTOR_NOT_FOUND);
        }
    }

    @GetMapping
    public ApiResponse<List<InstructorDTO>> getAllInstructors() {
        List<InstructorDTO> instructors = instructorService.getAllInstructors();
        return ApiResponse.success(instructors);
    }

    @PutMapping("/{id}")
    public ApiResponse<InstructorDTO> updateInstructor(
            @PathVariable String id, 
            @RequestBody CreateInstructorRequest request) {
        try {
            InstructorDTO instructor = instructorService.updateInstructor(id, request);
            return ApiResponse.success(instructor);
        } catch (RuntimeException e) {
            return ApiResponse.error(ErrorCode.INVALID_PARAM, e.getMessage());
        }
    }

    @DeleteMapping("/{id}")
    public ApiResponse<Void> deleteInstructor(@PathVariable String id) {
        try {
            instructorService.deleteInstructor(id);
            return ApiResponse.success();
        } catch (RuntimeException e) {
            return ApiResponse.error(ErrorCode.INSTRUCTOR_NOT_FOUND);
        }
    }
}
