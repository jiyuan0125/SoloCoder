package com.factory.workorder.entity;

import com.factory.workorder.enums.OperationType;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import javax.persistence.*;
import java.time.LocalDateTime;

@Entity
@Table(name = "workorder_logs")
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class WorkorderLog {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(nullable = false)
    private Long workorderId;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    private OperationType operationType;

    @Column(nullable = false)
    private String operator;

    @Column(columnDefinition = "TEXT")
    private String remark;

    @Column(nullable = false, updatable = false)
    private LocalDateTime operateTime;

    @PrePersist
    protected void onCreate() {
        if (operateTime == null) {
            operateTime = LocalDateTime.now();
        }
    }
}
