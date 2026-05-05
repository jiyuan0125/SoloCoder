package com.performance.server.service;

import com.performance.common.constant.ErrorCode;
import com.performance.common.dto.OriginalDepartmentCommentDTO;
import com.performance.common.dto.PerformanceReviewDTO;
import com.performance.common.enums.PerformanceGrade;
import com.performance.common.enums.ReviewStatus;
import com.performance.common.request.ConfirmReviewRequest;
import com.performance.common.request.QueryHistoryRequest;
import com.performance.common.request.SubmitManagerReviewRequest;
import com.performance.common.request.SubmitSelfReviewRequest;
import com.performance.common.util.ScoreValidator;
import com.performance.server.entity.Department;
import com.performance.server.entity.Employee;
import com.performance.server.entity.OriginalDepartmentComment;
import com.performance.server.entity.PerformanceReview;
import com.performance.server.entity.ReviewCycle;
import com.performance.server.repository.DepartmentRepository;
import com.performance.server.repository.EmployeeRepository;
import com.performance.server.repository.PerformanceReviewRepository;
import com.performance.server.repository.ReviewCycleRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.math.BigDecimal;
import java.math.RoundingMode;
import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;

@Service
public class PerformanceReviewService {

    @Autowired
    private PerformanceReviewRepository performanceReviewRepository;

    @Autowired
    private ReviewCycleRepository reviewCycleRepository;

    @Autowired
    private EmployeeRepository employeeRepository;

    @Autowired
    private DepartmentRepository departmentRepository;

    public Optional<ErrorCode> submitSelfReview(SubmitSelfReviewRequest request) {
        Optional<ReviewCycle> cycleOpt = reviewCycleRepository.findById(request.getCycleId());
        if (!cycleOpt.isPresent()) {
            return Optional.of(ErrorCode.CYCLE_NOT_FOUND);
        }

        ReviewCycle cycle = cycleOpt.get();
        if (cycle.getStatus() != ReviewStatus.SELF_REVIEW) {
            return Optional.of(ErrorCode.CYCLE_NOT_IN_SELF_REVIEW);
        }

        Optional<PerformanceReview> reviewOpt = performanceReviewRepository.findByCycleIdAndEmployeeId(
                request.getCycleId(), request.getEmployeeId());
        if (!reviewOpt.isPresent()) {
            return Optional.of(ErrorCode.NOT_FOUND);
        }

        PerformanceReview review = reviewOpt.get();
        if (review.getSelfScore() != null) {
            return Optional.of(ErrorCode.REVIEW_ALREADY_SUBMITTED);
        }

        if (!ScoreValidator.isValidScore(request.getScore())) {
            return Optional.of(ErrorCode.SCORE_INVALID);
        }

        review.setSelfScore(ScoreValidator.normalizeScore(request.getScore()));
        review.setSelfComment(request.getComment());
        review.setSelfReviewTime(LocalDateTime.now());

        performanceReviewRepository.save(review);
        return Optional.empty();
    }

    public Optional<ErrorCode> submitManagerReview(SubmitManagerReviewRequest request) {
        Optional<ReviewCycle> cycleOpt = reviewCycleRepository.findById(request.getCycleId());
        if (!cycleOpt.isPresent()) {
            return Optional.of(ErrorCode.CYCLE_NOT_FOUND);
        }

        ReviewCycle cycle = cycleOpt.get();
        if (cycle.getStatus() != ReviewStatus.MANAGER_REVIEW) {
            return Optional.of(ErrorCode.CYCLE_NOT_IN_MANAGER_REVIEW);
        }

        Optional<PerformanceReview> reviewOpt = performanceReviewRepository.findByCycleIdAndEmployeeId(
                request.getCycleId(), request.getEmployeeId());
        if (!reviewOpt.isPresent()) {
            return Optional.of(ErrorCode.NOT_FOUND);
        }

        PerformanceReview review = reviewOpt.get();
        
        if (!request.getManagerId().equals(review.getManagerId())) {
            return Optional.of(ErrorCode.FORBIDDEN);
        }

        if (review.getManagerScore() != null) {
            return Optional.of(ErrorCode.REVIEW_ALREADY_SUBMITTED);
        }

        if (!ScoreValidator.isValidScore(request.getScore())) {
            return Optional.of(ErrorCode.SCORE_INVALID);
        }

        if (review.getSelfScore() != null && 
            ScoreValidator.isDifferenceZero(request.getScore(), review.getSelfScore())) {
            return Optional.of(ErrorCode.SCORE_DIFFERENCE_ZERO);
        }

        review.setManagerScore(ScoreValidator.normalizeScore(request.getScore()));
        review.setManagerComment(request.getComment());
        review.setManagerReviewTime(LocalDateTime.now());

        performanceReviewRepository.save(review);
        return Optional.empty();
    }

    public Optional<ErrorCode> addOriginalDepartmentComment(Long cycleId, Long employeeId, Long originalDeptManagerId, String comment) {
        Optional<PerformanceReview> reviewOpt = performanceReviewRepository.findByCycleIdAndEmployeeId(
                cycleId, employeeId);
        if (!reviewOpt.isPresent()) {
            return Optional.of(ErrorCode.NOT_FOUND);
        }

        PerformanceReview review = reviewOpt.get();
        
        Optional<Employee> managerOpt = employeeRepository.findById(originalDeptManagerId);
        if (!managerOpt.isPresent()) {
            return Optional.of(ErrorCode.MANAGER_NOT_FOUND);
        }
        Employee manager = managerOpt.get();

        OriginalDepartmentComment originalComment = new OriginalDepartmentComment();
        originalComment.setDepartmentId(manager.getDepartmentId());
        originalComment.setManagerId(manager.getId());
        originalComment.setManagerName(manager.getName());
        originalComment.setComment(comment);
        originalComment.setCommentTime(LocalDateTime.now());

        if (manager.getDepartmentId() != null) {
            Optional<Department> deptOpt = departmentRepository.findById(manager.getDepartmentId());
            if (deptOpt.isPresent()) {
                originalComment.setDepartmentName(deptOpt.get().getName());
            }
        }

        review.getOriginalDepartmentComments().add(originalComment);
        performanceReviewRepository.save(review);
        
        return Optional.empty();
    }

    public Optional<ErrorCode> confirmReview(ConfirmReviewRequest request) {
        Optional<ReviewCycle> cycleOpt = reviewCycleRepository.findById(request.getCycleId());
        if (!cycleOpt.isPresent()) {
            return Optional.of(ErrorCode.CYCLE_NOT_FOUND);
        }

        ReviewCycle cycle = cycleOpt.get();
        if (cycle.getStatus() != ReviewStatus.HR_CONFIRM) {
            return Optional.of(ErrorCode.CYCLE_NOT_IN_HR_CONFIRM);
        }

        Optional<PerformanceReview> reviewOpt = performanceReviewRepository.findByCycleIdAndEmployeeId(
                request.getCycleId(), request.getEmployeeId());
        if (!reviewOpt.isPresent()) {
            return Optional.of(ErrorCode.NOT_FOUND);
        }

        PerformanceReview review = reviewOpt.get();
        
        if (review.getSelfScore() == null) {
            return Optional.of(ErrorCode.REVIEW_NOT_SUBMITTED);
        }
        if (review.getManagerScore() == null) {
            return Optional.of(ErrorCode.REVIEW_NOT_SUBMITTED);
        }

        BigDecimal finalScore = calculateFinalScore(review.getSelfScore(), review.getManagerScore());
        review.setFinalScore(finalScore);
        
        PerformanceGrade grade = PerformanceGrade.fromScore(finalScore);
        review.setFinalGrade(grade);
        review.setBonusCoefficient(grade.getBonusCoefficient());
        
        review.setCompletedTime(LocalDateTime.now());
        review.setStatus(ReviewStatus.COMPLETED);

        int consecutiveCGrads = performanceReviewRepository.countConsecutiveCGrads(review.getEmployeeId(), grade);
        review.setConsecutiveCGrads(consecutiveCGrads);
        
        review.setPipEligible(consecutiveCGrads >= 2);
        review.setTerminationEligible(consecutiveCGrads >= 3 || grade == PerformanceGrade.D);

        performanceReviewRepository.save(review);
        
        return Optional.empty();
    }

    public void calculateRankings(Long cycleId, Long departmentId) {
        List<PerformanceReview> reviews = performanceReviewRepository.findByDepartmentIdAndCycleId(
                departmentId, cycleId).stream()
                .filter(r -> r.getStatus() == ReviewStatus.COMPLETED && r.getFinalScore() != null)
                .sorted(Comparator.comparing(PerformanceReview::getFinalScore).reversed())
                .collect(Collectors.toList());

        int totalEmployees = reviews.size();
        
        for (int i = 0; i < reviews.size(); i++) {
            PerformanceReview review = reviews.get(i);
            review.setRanking(i + 1);
            review.setTotalEmployees(totalEmployees);
            
            BigDecimal percentile = calculatePercentile(i + 1, totalEmployees);
            review.setPercentile(percentile);
            
            performanceReviewRepository.save(review);
        }
    }

    private BigDecimal calculateFinalScore(BigDecimal selfScore, BigDecimal managerScore) {
        if (managerScore == null) {
            return selfScore;
        }
        return managerScore;
    }

    private BigDecimal calculatePercentile(int rank, int total) {
        if (total == 0) {
            return BigDecimal.ZERO;
        }
        return new BigDecimal((total - rank) * 100)
                .divide(new BigDecimal(total), 2, RoundingMode.HALF_UP);
    }

    public Optional<PerformanceReviewDTO> getReviewById(Long id, boolean isEmployeeView) {
        Optional<PerformanceReview> reviewOpt = performanceReviewRepository.findById(id);
        if (reviewOpt.isPresent()) {
            return Optional.of(toDTO(reviewOpt.get(), isEmployeeView));
        }
        return Optional.empty();
    }

    public Optional<PerformanceReviewDTO> getReviewByCycleAndEmployee(Long cycleId, Long employeeId, boolean isEmployeeView) {
        Optional<PerformanceReview> reviewOpt = performanceReviewRepository.findByCycleIdAndEmployeeId(cycleId, employeeId);
        if (reviewOpt.isPresent()) {
            return Optional.of(toDTO(reviewOpt.get(), isEmployeeView));
        }
        return Optional.empty();
    }

    public List<PerformanceReviewDTO> getReviewsByCycleId(Long cycleId) {
        List<PerformanceReview> reviews = performanceReviewRepository.findByCycleId(cycleId);
        List<PerformanceReviewDTO> dtos = new ArrayList<>();
        for (PerformanceReview review : reviews) {
            dtos.add(toDTO(review, false));
        }
        return dtos;
    }

    public List<PerformanceReviewDTO> getReviewsByEmployeeId(Long employeeId) {
        List<PerformanceReview> reviews = performanceReviewRepository.findByEmployeeId(employeeId);
        List<PerformanceReviewDTO> dtos = new ArrayList<>();
        for (PerformanceReview review : reviews) {
            dtos.add(toDTO(review, true));
        }
        return dtos;
    }

    public List<PerformanceReviewDTO> queryHistory(QueryHistoryRequest request) {
        List<PerformanceReview> reviews = performanceReviewRepository.findAll();
        
        if (request.getDepartmentId() != null) {
            reviews = reviews.stream()
                    .filter(r -> request.getDepartmentId().equals(r.getDepartmentId()))
                    .collect(Collectors.toList());
        }
        
        if (request.getYear() != null || request.getQuarter() != null) {
            List<Long> cycleIds = new ArrayList<>();
            List<ReviewCycle> cycles = reviewCycleRepository.findAll();
            for (ReviewCycle cycle : cycles) {
                boolean match = true;
                if (request.getYear() != null && !request.getYear().equals(cycle.getYear())) {
                    match = false;
                }
                if (request.getQuarter() != null && !request.getQuarter().equals(cycle.getQuarter())) {
                    match = false;
                }
                if (match) {
                    cycleIds.add(cycle.getId());
                }
            }
            reviews = reviews.stream()
                    .filter(r -> cycleIds.contains(r.getCycleId()))
                    .collect(Collectors.toList());
        }
        
        if (request.getEmployeeId() != null) {
            reviews = reviews.stream()
                    .filter(r -> request.getEmployeeId().equals(r.getEmployeeId()))
                    .collect(Collectors.toList());
        }

        reviews = reviews.stream()
                .filter(r -> r.getStatus() == ReviewStatus.COMPLETED)
                .collect(Collectors.toList());

        List<PerformanceReviewDTO> dtos = new ArrayList<>();
        for (PerformanceReview review : reviews) {
            dtos.add(toDTO(review, request.getEmployeeId() != null));
        }
        return dtos;
    }

    private PerformanceReviewDTO toDTO(PerformanceReview review, boolean isEmployeeView) {
        PerformanceReviewDTO dto = new PerformanceReviewDTO();
        dto.setId(review.getId());
        dto.setEmployeeId(review.getEmployeeId());
        dto.setDepartmentId(review.getDepartmentId());
        dto.setCycleId(review.getCycleId());
        dto.setStatus(review.getStatus());
        
        Optional<Employee> employeeOpt = employeeRepository.findById(review.getEmployeeId());
        if (employeeOpt.isPresent()) {
            dto.setEmployeeName(employeeOpt.get().getName());
        }
        
        if (review.getDepartmentId() != null) {
            Optional<Department> deptOpt = departmentRepository.findById(review.getDepartmentId());
            if (deptOpt.isPresent()) {
                dto.setDepartmentName(deptOpt.get().getName());
            }
        }
        
        Optional<ReviewCycle> cycleOpt = reviewCycleRepository.findById(review.getCycleId());
        if (cycleOpt.isPresent()) {
            dto.setCycleName(cycleOpt.get().getCycleName());
        }
        
        if (isEmployeeView && review.getStatus() == ReviewStatus.COMPLETED) {
            dto.setFinalScore(review.getFinalScore());
            dto.setFinalGrade(review.getFinalGrade());
            dto.setBonusCoefficient(review.getBonusCoefficient());
            dto.setPercentile(review.getPercentile());
            dto.setRanking(review.getRanking());
            dto.setTotalEmployees(review.getTotalEmployees());
            dto.setCompletedTime(review.getCompletedTime());
            dto.setPipEligible(review.getPipEligible());
            dto.setTerminationEligible(review.getTerminationEligible());
            dto.setConsecutiveCGrads(review.getConsecutiveCGrads());
        } else if (!isEmployeeView) {
            dto.setSelfScore(review.getSelfScore());
            dto.setSelfComment(review.getSelfComment());
            dto.setSelfReviewTime(review.getSelfReviewTime());
            dto.setManagerId(review.getManagerId());
            
            if (review.getManagerId() != null) {
                Optional<Employee> managerOpt = employeeRepository.findById(review.getManagerId());
                if (managerOpt.isPresent()) {
                    dto.setManagerName(managerOpt.get().getName());
                }
            }
            
            dto.setManagerScore(review.getManagerScore());
            dto.setManagerComment(review.getManagerComment());
            dto.setManagerReviewTime(review.getManagerReviewTime());
            
            List<OriginalDepartmentCommentDTO> originalComments = new ArrayList<>();
            for (OriginalDepartmentComment oc : review.getOriginalDepartmentComments()) {
                OriginalDepartmentCommentDTO ocDto = new OriginalDepartmentCommentDTO();
                ocDto.setDepartmentId(oc.getDepartmentId());
                ocDto.setDepartmentName(oc.getDepartmentName());
                ocDto.setManagerId(oc.getManagerId());
                ocDto.setManagerName(oc.getManagerName());
                ocDto.setComment(oc.getComment());
                ocDto.setCommentTime(oc.getCommentTime());
                originalComments.add(ocDto);
            }
            dto.setOriginalDepartmentComments(originalComments);
            
            dto.setFinalScore(review.getFinalScore());
            dto.setFinalGrade(review.getFinalGrade());
            dto.setBonusCoefficient(review.getBonusCoefficient());
            dto.setPercentile(review.getPercentile());
            dto.setRanking(review.getRanking());
            dto.setTotalEmployees(review.getTotalEmployees());
            dto.setCompletedTime(review.getCompletedTime());
            dto.setPipEligible(review.getPipEligible());
            dto.setTerminationEligible(review.getTerminationEligible());
            dto.setConsecutiveCGrads(review.getConsecutiveCGrads());
        }
        
        return dto;
    }
}