package com.logaggregator.repository;

import com.logaggregator.model.LogEntry;
import com.logaggregator.model.LogLevel;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.time.LocalDateTime;
import java.util.List;

@Repository
public interface LogEntryRepository extends JpaRepository<LogEntry, Long> {

    Page<LogEntry> findByServiceAndTimestampBetween(
            String service, LocalDateTime startTime, LocalDateTime endTime, Pageable pageable);

    Page<LogEntry> findByLevelInAndTimestampBetween(
            List<LogLevel> levels, LocalDateTime startTime, LocalDateTime endTime, Pageable pageable);

    Page<LogEntry> findByServiceAndLevelInAndTimestampBetween(
            String service, List<LogLevel> levels, LocalDateTime startTime, LocalDateTime endTime, Pageable pageable);

    List<LogEntry> findByTraceIdOrderByTimestampAsc(String traceId);

    long deleteByTimestampBefore(LocalDateTime timestamp);

    @Query("SELECT l.service, COUNT(l) FROM LogEntry l GROUP BY l.service")
    List<Object[]> countByService();

    @Query("SELECT l.level, COUNT(l) FROM LogEntry l GROUP BY l.level")
    List<Object[]> countByLevel();

    @Query("SELECT FUNCTION('DATE_FORMAT', l.timestamp, '%Y-%m-%d %H:00'), COUNT(l) " +
           "FROM LogEntry l WHERE l.level = 'ERROR' AND l.timestamp >= :since GROUP BY FUNCTION('DATE_FORMAT', l.timestamp, '%Y-%m-%d %H:00')")
    List<Object[]> countErrorsByHour(@Param("since") LocalDateTime since);

    @Query("SELECT COUNT(l) FROM LogEntry l WHERE l.timestamp BETWEEN :start AND :end")
    long countByTimestampBetween(@Param("start") LocalDateTime start, @Param("end") LocalDateTime end);
}
