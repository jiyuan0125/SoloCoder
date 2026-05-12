package com.exam.repository;

import com.exam.entity.AnswerRecord;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

@Repository
public interface AnswerRecordRepository extends JpaRepository<AnswerRecord, Long> {
    
    @Query("SELECT ar FROM AnswerRecord ar JOIN FETCH ar.examQuestion WHERE ar.examRecord.id = :examRecordId")
    List<AnswerRecord> findByExamRecordId(@Param("examRecordId") Long examRecordId);
    
    Optional<AnswerRecord> findByExamRecordIdAndExamQuestionId(Long examRecordId, Long examQuestionId);
    
    void deleteByExamRecordId(Long examRecordId);
}
