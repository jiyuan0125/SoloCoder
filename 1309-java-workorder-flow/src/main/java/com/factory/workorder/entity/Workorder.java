package com.factory.workorder.entity;

import com.factory.workorder.enums.WorkorderPriority;
import com.factory.workorder.enums.WorkorderStatus;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import javax.persistence.*;
import java.time.LocalDateTime;

@Entity
@Table(name = "workorders")
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class Workorder {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(nullable = false, unique = true)
    private String orderNo;

    @Column(nullable = false)
    private String equipmentName;

    @Column(nullable = false, columnDefinition = "TEXT")
    private String faultDescription;

    @Column(nullable = false)
    private String reporter;

    private String currentHandler;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    private WorkorderStatus status;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    private WorkorderPriority priority;

    @Column(nullable = false, updatable = false)
    private LocalDateTime createTime;

    private LocalDateTime completeTime;

    @Version
    private Long version;

    @PrePersist
    protected void onCreate() {
        if (createTime == null) {
            createTime = LocalDateTime.now();
        }
        if (status == null) {
            status = WorkorderStatus.PENDING_ASSIGN;
        }
    }

    public boolean isCompleted() {
        return WorkorderStatus.COMPLETED.equals(status);
    }
}
