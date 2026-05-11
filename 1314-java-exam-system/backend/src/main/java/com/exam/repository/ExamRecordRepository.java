package com.exam.repository;

import com.exam.entity.ExamRecord;
import com.exam.enums.ExamStatus;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

@Repository
public interface ExamRecordRepository extends JpaRepository<ExamRecord, Long> {
    List<ExamRecord> findByUserIdOrderByCreatedAtDesc(Long userId);
    
    List<ExamRecord> findByExamPaperIdOrderByCreatedAtDesc(Long examPaperId);
    
    Optional<ExamRecord> findByUserIdAndExamPaperIdAndStatusIn(
            Long userId, 
            Long examPaperId, 
            List<ExamStatus> statuses);
    
    long countByUserIdAndExamPaperId(Long userId, Long examPaperId);
    
    List<ExamRecord> findByStatus(ExamStatus status);
}
