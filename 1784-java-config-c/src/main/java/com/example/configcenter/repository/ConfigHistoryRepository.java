package com.example.configcenter.repository;

import com.example.configcenter.entity.ConfigHistory;
import org.springframework.data.jpa.repository.JpaRepository;

import java.time.LocalDateTime;
import java.util.List;

public interface ConfigHistoryRepository extends JpaRepository<ConfigHistory, Long> {
    List<ConfigHistory> findByConfigKeyOrderByCreatedAtDesc(String configKey);
    List<ConfigHistory> findByConfigKeyAndCreatedAtBetweenOrderByCreatedAtDesc(String configKey, LocalDateTime startTime, LocalDateTime endTime);
    List<ConfigHistory> findTop50ByConfigKeyOrderByCreatedAtDesc(String configKey);
}
