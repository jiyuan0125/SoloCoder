package com.exam.repository;

import com.exam.entity.ExamRecord;
import com.exam.enums.ExamStatus;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
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
    
    @Query("SELECT COUNT(e) FROM ExamRecord e WHERE e.user.id = :userId AND e.examPaper.id = :examPaperId AND e.status IN :statuses")
    long countByUserIdAndExamPaperIdAndStatusIn(
            @Param("userId") Long userId,
            @Param("examPaperId") Long examPaperId,
            @Param("statuses") List<ExamStatus> statuses);
    
    List<ExamRecord> findByStatus(ExamStatus status);
}
