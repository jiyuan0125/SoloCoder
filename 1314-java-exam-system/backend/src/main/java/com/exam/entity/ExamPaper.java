package com.exam.entity;

import jakarta.persistence.*;
import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.AllArgsConstructor;

import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Entity
@Table(name = "exam_papers")
public class ExamPaper {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(nullable = false)
    private String title;

    private String description;

    @Column(nullable = false)
    private int totalScore;

    @Column(nullable = false)
    private int passScore;

    @Column(nullable = false)
    private int durationMinutes;

    @Column(nullable = false)
    private boolean canRetake = false;

    @Column(name = "max_retake_count")
    private Integer maxRetakeCount;

    @Column(name = "switch_count_limit", nullable = false)
    private int switchCountLimit = 3;

    @Column(name = "switch_duration_limit_seconds", nullable = false)
    private int switchDurationLimitSeconds = 30;

    @OneToMany(mappedBy = "examPaper", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<ExamRule> rules = new ArrayList<>();

    @OneToMany(mappedBy = "examPaper", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<ExamRecord> examRecords = new ArrayList<>();

    private boolean enabled = true;

    private LocalDateTime createdAt;

    private LocalDateTime updatedAt;

    @PrePersist
    protected void onCreate() {
        createdAt = LocalDateTime.now();
        updatedAt = LocalDateTime.now();
    }

    @PreUpdate
    protected void onUpdate() {
        updatedAt = LocalDateTime.now();
    }
}
