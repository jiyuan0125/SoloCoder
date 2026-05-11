package com.exam.repository;

import com.exam.entity.Question;
import com.exam.enums.DifficultyLevel;
import com.exam.enums.QuestionType;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface QuestionRepository extends JpaRepository<Question, Long> {
    List<Question> findByCategoryId(Long categoryId);
    
    List<Question> findByType(QuestionType type);
    
    List<Question> findByDifficulty(DifficultyLevel difficulty);
    
    @Query("SELECT q FROM Question q WHERE q.category.id = :categoryId AND q.type = :type")
    List<Question> findByCategoryIdAndType(@Param("categoryId") Long categoryId, @Param("type") QuestionType type);
    
    @Query("SELECT q FROM Question q WHERE q.category.id = :categoryId AND q.type = :type AND q.difficulty = :difficulty")
    List<Question> findByCategoryIdAndTypeAndDifficulty(
            @Param("categoryId") Long categoryId,
            @Param("type") QuestionType type,
            @Param("difficulty") DifficultyLevel difficulty);
    
    @Query("SELECT q FROM Question q WHERE q.category.id = :categoryId AND q.type = :type AND q.difficulty IN :difficulties")
    List<Question> findByCategoryIdAndTypeAndDifficultyIn(
            @Param("categoryId") Long categoryId,
            @Param("type") QuestionType type,
            @Param("difficulties") List<DifficultyLevel> difficulties);
    
    long countByCategoryIdAndType(Long categoryId, QuestionType type);
    
    long countByCategoryIdAndTypeAndDifficulty(Long categoryId, QuestionType type, DifficultyLevel difficulty);
}
