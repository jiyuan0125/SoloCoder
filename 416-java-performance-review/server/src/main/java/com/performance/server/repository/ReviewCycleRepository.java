package com.performance.server.repository;

import com.performance.common.enums.ReviewStatus;
import com.performance.server.entity.ReviewCycle;
import org.springframework.stereotype.Repository;

import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicLong;

@Repository
public class ReviewCycleRepository {
    private final Map<Long, ReviewCycle> cycles = new ConcurrentHashMap<>();
    private final AtomicLong idGenerator = new AtomicLong(1);

    public ReviewCycle save(ReviewCycle cycle) {
        if (cycle.getId() == null) {
            cycle.setId(idGenerator.getAndIncrement());
            cycle.setCreatedAt(LocalDateTime.now());
        }
        cycle.setUpdatedAt(LocalDateTime.now());
        cycles.put(cycle.getId(), cycle);
        return cycle;
    }

    public Optional<ReviewCycle> findById(Long id) {
        return Optional.ofNullable(cycles.get(id));
    }

    public List<ReviewCycle> findAll() {
        return new ArrayList<>(cycles.values());
    }

    public Optional<ReviewCycle> findByYearAndQuarter(String year, String quarter) {
        for (ReviewCycle cycle : cycles.values()) {
            if (year.equals(cycle.getYear()) && quarter.equals(cycle.getQuarter())) {
                return Optional.of(cycle);
            }
        }
        return Optional.empty();
    }

    public List<ReviewCycle> findByStatus(ReviewStatus status) {
        List<ReviewCycle> result = new ArrayList<>();
        for (ReviewCycle cycle : cycles.values()) {
            if (status == cycle.getStatus()) {
                result.add(cycle);
            }
        }
        return result;
    }

    public boolean existsByYearAndQuarter(String year, String quarter) {
        for (ReviewCycle cycle : cycles.values()) {
            if (year.equals(cycle.getYear()) && quarter.equals(cycle.getQuarter())) {
                return true;
            }
        }
        return false;
    }

    public boolean existsById(Long id) {
        return cycles.containsKey(id);
    }

    public void deleteById(Long id) {
        cycles.remove(id);
    }
}