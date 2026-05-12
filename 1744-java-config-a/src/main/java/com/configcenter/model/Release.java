package com.configcenter.model;

import com.fasterxml.jackson.annotation.JsonIgnore;
import jakarta.persistence.*;
import lombok.Data;
import org.hibernate.annotations.CreationTimestamp;
import org.hibernate.annotations.UpdateTimestamp;

import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;

@Data
@Entity
@Table(name = "releases")
public class Release {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @JsonIgnore
    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "project_id", nullable = false)
    private Project project;

    @Column(nullable = false)
    private String environment;

    @Enumerated(EnumType.STRING)
    private ReleaseType type = ReleaseType.FULL;

    @Enumerated(EnumType.STRING)
    private ReleaseStatus status = ReleaseStatus.PENDING;

    @Column(columnDefinition = "TEXT")
    private String grayInstances;

    private LocalDateTime startTime;

    private LocalDateTime endTime;

    @JsonIgnore
    @OneToMany(mappedBy = "release", cascade = CascadeType.ALL)
    private List<ReleaseConfigChange> configChanges = new ArrayList<>();

    private String validationCallbackResult;

    private String createdBy;

    @CreationTimestamp
    private LocalDateTime createdAt;

    @UpdateTimestamp
    private LocalDateTime updatedAt;

    public enum ReleaseType {
        FULL,
        GRAY
    }

    public enum ReleaseStatus {
        PENDING,
        VALIDATING,
        GRAY_IN_PROGRESS,
        GRAY_COMPLETED,
        IN_PROGRESS,
        COMPLETED,
        FAILED,
        ROLLED_BACK
    }
}
