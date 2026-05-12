package com.configcenter.model;

import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;

@Entity
@Table(name = "audit_logs")
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class AuditLog {
    
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    
    @Column(nullable = false)
    private String namespace;
    
    @Column(name = "group_name", nullable = false)
    private String group;
    
    @Column(name = "config_key", nullable = false)
    private String key;
    
    @Column(nullable = false)
    private String action;
    
    @Lob
    @Column(name = "old_config_value")
    private String oldValue;
    
    @Lob
    @Column(name = "new_config_value")
    private String newValue;
    
    @Column(name = "target_version")
    private Long targetVersion;
    
    @Column(name = "old_version")
    private Long oldVersion;
    
    @Column(name = "new_version")
    private Long newVersion;
    
    @Column(nullable = false)
    private String operator;
    
    @Column(name = "client_ip", nullable = false)
    private String clientIp;
    
    @Column(name = "created_at", nullable = false)
    private LocalDateTime createdAt;
    
    @PrePersist
    protected void onCreate() {
        createdAt = LocalDateTime.now();
    }
}
