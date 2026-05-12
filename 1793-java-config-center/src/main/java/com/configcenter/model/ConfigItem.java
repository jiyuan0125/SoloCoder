package com.configcenter.model;

import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;

@Entity
@Table(name = "config_items", uniqueConstraints = {
    @UniqueConstraint(columnNames = {"namespace", "groupName", "configKey"})
})
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class ConfigItem {
    
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
    
    @Column(name = "updated_at", nullable = false)
    private LocalDateTime updatedAt;
    
    @PrePersist
    protected void onCreate() {
        createdAt = LocalDateTime.now();
        updatedAt = LocalDateTime.now();
        if (version == null) {
            version = 1L;
        }
    }
    
    @PreUpdate
    protected void onUpdate() {
        updatedAt = LocalDateTime.now();
    }
}
