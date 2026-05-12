package com.example.configcenter.entity;

import jakarta.persistence.*;
import lombok.Data;
import org.hibernate.annotations.CreationTimestamp;

import java.time.LocalDateTime;

@Data
@Entity
@Table(name = "config_history", indexes = {
    @Index(name = "idx_config_key_version", columnList = "configKey, version"),
    @Index(name = "idx_created_at", columnList = "createdAt")
})
public class ConfigHistory {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    
    @Column(nullable = false)
    private String configKey;
    
    @Column(columnDefinition = "TEXT", nullable = false)
    private String configValue;
    
    @Column(nullable = false)
    private Integer version;
    
    @CreationTimestamp
    private LocalDateTime createdAt;
}
