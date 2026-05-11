package com.exam.repository;

import com.exam.entity.ExamQuestion;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface ExamQuestionRepository extends JpaRepository<ExamQuestion, Long> {
    List<ExamQuestion> findByExamRecordIdOrderByDisplayOrderAsc(Long examRecordId);
    
    void deleteByExamRecordId(Long examRecordId);
}
