package com.configcenter.model;

import jakarta.persistence.*;
import lombok.Data;
import org.hibernate.annotations.CreationTimestamp;
import org.hibernate.annotations.UpdateTimestamp;

import java.time.LocalDateTime;

@Data
@Entity
@Table(name = "client_subscriptions", uniqueConstraints = {
        @UniqueConstraint(columnNames = {"instanceId", "projectId", "environment"})
})
public class ClientSubscription {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(nullable = false)
    private String instanceId;

    @Column(nullable = false)
    private Long projectId;

    @Column(nullable = false)
    private String environment;

    private String lastKnownVersion;

    private LocalDateTime lastHeartbeat;

    @CreationTimestamp
    private LocalDateTime createdAt;

    @UpdateTimestamp
    private LocalDateTime updatedAt;
}
