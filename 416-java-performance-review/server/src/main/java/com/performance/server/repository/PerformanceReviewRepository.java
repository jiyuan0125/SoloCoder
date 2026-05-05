package com.performance.server.repository;

import com.performance.common.enums.PerformanceGrade;
import com.performance.common.enums.ReviewStatus;
import com.performance.server.entity.PerformanceReview;
import org.springframework.stereotype.Repository;

import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicLong;
import java.util.stream.Collectors;

@Repository
public class PerformanceReviewRepository {
    private final Map<Long, PerformanceReview> reviews = new ConcurrentHashMap<>();
    private final AtomicLong idGenerator = new AtomicLong(1);

    public PerformanceReview save(PerformanceReview review) {
        if (review.getId() == null) {
            review.setId(idGenerator.getAndIncrement());
            review.setCreatedAt(LocalDateTime.now());
        }
        review.setUpdatedAt(LocalDateTime.now());
        reviews.put(review.getId(), review);
        return review;
    }

    public Optional<PerformanceReview> findById(Long id) {
        return Optional.ofNullable(reviews.get(id));
    }

    public List<PerformanceReview> findAll() {
        return new ArrayList<>(reviews.values());
    }

    public Optional<PerformanceReview> findByCycleIdAndEmployeeId(Long cycleId, Long employeeId) {
        for (PerformanceReview review : reviews.values()) {
            if (cycleId.equals(review.getCycleId()) && employeeId.equals(review.getEmployeeId())) {
                return Optional.of(review);
            }
        }
        return Optional.empty();
    }

    public List<PerformanceReview> findByCycleId(Long cycleId) {
        List<PerformanceReview> result = new ArrayList<>();
        for (PerformanceReview review : reviews.values()) {
            if (cycleId.equals(review.getCycleId())) {
                result.add(review);
            }
        }
        return result;
    }

    public List<PerformanceReview> findByDepartmentIdAndCycleId(Long departmentId, Long cycleId) {
        List<PerformanceReview> result = new ArrayList<>();
        for (PerformanceReview review : reviews.values()) {
            if (departmentId.equals(review.getDepartmentId()) && cycleId.equals(review.getCycleId())) {
                result.add(review);
            }
        }
        return result;
    }

    public List<PerformanceReview> findByEmployeeId(Long employeeId) {
        List<PerformanceReview> result = new ArrayList<>();
        for (PerformanceReview review : reviews.values()) {
            if (employeeId.equals(review.getEmployeeId())) {
                result.add(review);
            }
        }
        return result.stream()
                .sorted(Comparator.comparing(PerformanceReview::getCycleId).reversed())
                .collect(Collectors.toList());
    }

    public List<PerformanceReview> findByManagerId(Long managerId) {
        List<PerformanceReview> result = new ArrayList<>();
        for (PerformanceReview review : reviews.values()) {
            if (managerId.equals(review.getManagerId())) {
                result.add(review);
            }
        }
        return result;
    }

    public List<PerformanceReview> findByDepartmentIdAndYearAndQuarter(Long departmentId, String year, String quarter) {
        List<PerformanceReview> result = new ArrayList<>();
        for (PerformanceReview review : reviews.values()) {
            if (departmentId.equals(review.getDepartmentId()) && 
                ReviewStatus.COMPLETED == review.getStatus()) {
                result.add(review);
            }
        }
        return result;
    }

    public int countConsecutiveCGrads(Long employeeId, PerformanceGrade currentGrade) {
        List<PerformanceReview> employeeReviews = findByEmployeeId(employeeId);
        int count = 0;
        
        if (currentGrade == PerformanceGrade.C) {
            count = 1;
        } else if (currentGrade == PerformanceGrade.D) {
            return 0;
        }
        
        for (PerformanceReview review : employeeReviews) {
            if (review.getFinalGrade() == PerformanceGrade.C) {
                count++;
            } else if (review.getFinalGrade() != null && review.getFinalGrade() != PerformanceGrade.C) {
                break;
            }
        }
        
        return count;
    }

    public boolean existsById(Long id) {
        return reviews.containsKey(id);
    }

    public void deleteById(Long id) {
        reviews.remove(id);
    }
}