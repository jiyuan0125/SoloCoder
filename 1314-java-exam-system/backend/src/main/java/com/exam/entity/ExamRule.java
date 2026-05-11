package com.exam.entity;

import com.exam.enums.DifficultyLevel;
import com.exam.enums.QuestionType;
import jakarta.persistence.*;
import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.AllArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "exam_rules")
public class ExamRule {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "exam_paper_id", nullable = false)
    private ExamPaper examPaper;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "category_id", nullable = false)
    private KnowledgeCategory category;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    private QuestionType questionType;

    @Column(nullable = false)
    private int questionCount;

    @Column(nullable = false)
    private int scorePerQuestion;

    @Enumerated(EnumType.STRING)
    private DifficultyLevel preferredDifficulty;

    @Column(name = "min_difficulty")
    private Integer minDifficultyLevel;

    @Column(name = "max_difficulty")
    private Integer maxDifficultyLevel;

    @Column(name = "min_specific_difficulty_count")
    private Integer minSpecificDifficultyCount;

    @Enumerated(EnumType.STRING)
    private DifficultyLevel specificDifficulty;
}
