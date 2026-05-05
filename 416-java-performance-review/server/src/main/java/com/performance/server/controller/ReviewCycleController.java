package com.performance.server.controller;

import com.performance.common.constant.ErrorCode;
import com.performance.common.dto.ApiResponse;
import com.performance.common.dto.ReviewCycleDTO;
import com.performance.common.request.CreateCycleRequest;
import com.performance.server.service.ReviewCycleService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;
import java.util.Optional;

@RestController
@RequestMapping("/api/cycles")
public class ReviewCycleController {

    @Autowired
    private ReviewCycleService reviewCycleService;

    @PostMapping
    public ApiResponse<Void> createCycle(@RequestBody CreateCycleRequest request) {
        Optional<ErrorCode> error = reviewCycleService.createCycle(request);
        if (error.isPresent()) {
            return ApiResponse.error(error.get());
        }
        return ApiResponse.success();
    }

    @GetMapping("/{id}")
    public ApiResponse<ReviewCycleDTO> getCycleById(@PathVariable Long id) {
        Optional<ReviewCycleDTO> cycleOpt = reviewCycleService.getCycleById(id);
        if (cycleOpt.isPresent()) {
            return ApiResponse.success(cycleOpt.get());
        }
        return ApiResponse.error(ErrorCode.NOT_FOUND);
    }

    @GetMapping
    public ApiResponse<List<ReviewCycleDTO>> getAllCycles() {
        List<ReviewCycleDTO> cycles = reviewCycleService.getAllCycles();
        return ApiResponse.success(cycles);
    }

    @PutMapping("/{id}/advance")
    public ApiResponse<Void> advanceToNextStage(@PathVariable Long id) {
        Optional<ErrorCode> error = reviewCycleService.advanceToNextStage(id);
        if (error.isPresent()) {
            return ApiResponse.error(error.get());
        }
        return ApiResponse.success();
    }
}