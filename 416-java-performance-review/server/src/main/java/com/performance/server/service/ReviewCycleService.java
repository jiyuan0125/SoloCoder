package com.performance.server.service;

import com.performance.common.constant.ErrorCode;
import com.performance.common.dto.ReviewCycleDTO;
import com.performance.common.enums.ReviewStatus;
import com.performance.common.request.CreateCycleRequest;
import com.performance.server.entity.Employee;
import com.performance.server.entity.PerformanceReview;
import com.performance.server.entity.ReviewCycle;
import com.performance.server.repository.EmployeeRepository;
import com.performance.server.repository.PerformanceReviewRepository;
import com.performance.server.repository.ReviewCycleRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

@Service
public class ReviewCycleService {

    @Autowired
    private ReviewCycleRepository reviewCycleRepository;

    @Autowired
    private EmployeeRepository employeeRepository;

    @Autowired
    private PerformanceReviewRepository performanceReviewRepository;

    public Optional<ErrorCode> createCycle(CreateCycleRequest request) {
        if (reviewCycleRepository.existsByYearAndQuarter(request.getYear(), request.getQuarter())) {
            return Optional.of(ErrorCode.CYCLE_ALREADY_EXISTS);
        }

        ReviewCycle cycle = new ReviewCycle();
        cycle.setYear(request.getYear());
        cycle.setQuarter(request.getQuarter());
        cycle.setCycleName(request.getYear() + "年第" + request.getQuarter() + "季度");
        cycle.setStatus(ReviewStatus.SELF_REVIEW);
        cycle.setStartDate(request.getStartDate());
        cycle.setEndDate(request.getEndDate());

        ReviewCycle savedCycle = reviewCycleRepository.save(cycle);

        List<Employee> employees = employeeRepository.findAll();
        for (Employee employee : employees) {
            PerformanceReview review = new PerformanceReview();
            review.setEmployeeId(employee.getId());
            review.setDepartmentId(employee.getDepartmentId());
            review.setCycleId(savedCycle.getId());
            review.setStatus(ReviewStatus.SELF_REVIEW);
            review.setManagerId(employee.getManagerId());
            performanceReviewRepository.save(review);
        }

        return Optional.empty();
    }

    public Optional<ReviewCycleDTO> getCycleById(Long id) {
        Optional<ReviewCycle> cycleOpt = reviewCycleRepository.findById(id);
        if (cycleOpt.isPresent()) {
            return Optional.of(toDTO(cycleOpt.get()));
        }
        return Optional.empty();
    }

    public List<ReviewCycleDTO> getAllCycles() {
        List<ReviewCycle> cycles = reviewCycleRepository.findAll();
        List<ReviewCycleDTO> dtos = new ArrayList<>();
        for (ReviewCycle cycle : cycles) {
            dtos.add(toDTO(cycle));
        }
        return dtos;
    }

    public Optional<ErrorCode> advanceToNextStage(Long cycleId) {
        Optional<ReviewCycle> cycleOpt = reviewCycleRepository.findById(cycleId);
        if (!cycleOpt.isPresent()) {
            return Optional.of(ErrorCode.CYCLE_NOT_FOUND);
        }

        ReviewCycle cycle = cycleOpt.get();
        ReviewStatus currentStatus = cycle.getStatus();

        if (currentStatus == ReviewStatus.SELF_REVIEW) {
            cycle.setStatus(ReviewStatus.MANAGER_REVIEW);
            updateAllReviewsStatus(cycleId, ReviewStatus.MANAGER_REVIEW);
        } else if (currentStatus == ReviewStatus.MANAGER_REVIEW) {
            cycle.setStatus(ReviewStatus.HR_CONFIRM);
            updateAllReviewsStatus(cycleId, ReviewStatus.HR_CONFIRM);
        } else if (currentStatus == ReviewStatus.HR_CONFIRM) {
            cycle.setStatus(ReviewStatus.COMPLETED);
        } else {
            return Optional.of(ErrorCode.CYCLE_ALREADY_COMPLETED);
        }

        reviewCycleRepository.save(cycle);
        return Optional.empty();
    }

    private void updateAllReviewsStatus(Long cycleId, ReviewStatus status) {
        List<PerformanceReview> reviews = performanceReviewRepository.findByCycleId(cycleId);
        for (PerformanceReview review : reviews) {
            review.setStatus(status);
            performanceReviewRepository.save(review);
        }
    }

    private ReviewCycleDTO toDTO(ReviewCycle cycle) {
        ReviewCycleDTO dto = new ReviewCycleDTO();
        dto.setId(cycle.getId());
        dto.setYear(cycle.getYear());
        dto.setQuarter(cycle.getQuarter());
        dto.setCycleName(cycle.getCycleName());
        dto.setStatus(cycle.getStatus());
        dto.setStartDate(cycle.getStartDate());
        dto.setEndDate(cycle.getEndDate());
        return dto;
    }
}