package com.configcenter.model;

import com.fasterxml.jackson.annotation.JsonIgnore;
import jakarta.persistence.*;
import lombok.Data;
import org.hibernate.annotations.CreationTimestamp;
import org.hibernate.annotations.UpdateTimestamp;

import java.time.LocalDateTime;

@Data
@Entity
@Table(name = "config_items", uniqueConstraints = {
        @UniqueConstraint(columnNames = {"project_id", "environment", "configKey"})
})
public class ConfigItem {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @JsonIgnore
    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "project_id", nullable = false)
    private Project project;

    @Column(nullable = false)
    private String environment;

    @Column(nullable = false)
    private String configKey;

    @Column(columnDefinition = "TEXT")
    private String currentValue;

    @Column(columnDefinition = "TEXT")
    private String pendingValue;

    @Enumerated(EnumType.STRING)
    private ReleaseStatus status = ReleaseStatus.RELEASED;

    private Long activeReleaseId;

    private Long pendingReleaseId;

    @CreationTimestamp
    private LocalDateTime createdAt;

    @UpdateTimestamp
    private LocalDateTime updatedAt;

    public enum ReleaseStatus {
        DRAFT,
        PENDING,
        GRAY_RELEASE,
        RELEASED,
        ROLLED_BACK
    }
}
