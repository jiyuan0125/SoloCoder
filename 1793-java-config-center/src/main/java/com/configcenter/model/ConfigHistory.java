package com.configcenter.model;

import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;

@Entity
@Table(name = "config_history")
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class ConfigHistory {
    
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    
    @Column(nullable = false)
    private String namespace;
    
    @Column(name = "group_name", nullable = false)
    private String group;
    
    @Column(name = "config_key", nullable = false)
    private String key;
    
    @Lob
    @Column(name = "config_value", nullable = false)
    private String value;
    
    @Column(nullable = false)
    private Long version;
    
    @Column(name = "created_at", nullable = false)
    private LocalDateTime createdAt;
    
    @Column(name = "created_by", nullable = false)
    private String createdBy;
    
    @Column(name = "client_ip", nullable = false)
    private String clientIp;
    
    @PrePersist
    protected void onCreate() {
        createdAt = LocalDateTime.now();
    }
}
