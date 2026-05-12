package com.example.configcenter.entity;

import jakarta.persistence.*;
import lombok.Data;
import org.hibernate.annotations.CreationTimestamp;

import java.time.LocalDateTime;

@Data
@Entity
@Table(name = "config_subscribers", indexes = {
    @Index(name = "idx_config_key", columnList = "configKey")
})
public class ConfigSubscriber {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;
    
    @Column(nullable = false)
    private String configKey;
    
    @Column(nullable = false)
    private String subscriberUrl;
    
    @CreationTimestamp
    private LocalDateTime createdAt;
}
