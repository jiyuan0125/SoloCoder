package com.exam.repository;

import com.exam.entity.AnswerRecord;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

@Repository
public interface AnswerRecordRepository extends JpaRepository<AnswerRecord, Long> {
    List<AnswerRecord> findByExamRecordId(Long examRecordId);
    
    Optional<AnswerRecord> findByExamRecordIdAndExamQuestionId(Long examRecordId, Long examQuestionId);
    
    void deleteByExamRecordId(Long examRecordId);
}
