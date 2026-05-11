package com.exam.repository;

import com.exam.entity.ExamRule;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface ExamRuleRepository extends JpaRepository<ExamRule, Long> {
    List<ExamRule> findByExamPaperId(Long examPaperId);
    
    void deleteByExamPaperId(Long examPaperId);
}
