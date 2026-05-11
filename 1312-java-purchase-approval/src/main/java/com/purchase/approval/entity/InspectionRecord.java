package com.purchase.approval.entity;

import lombok.Data;
import javax.persistence.*;
import java.time.LocalDateTime;

@Data
@Entity
@Table(name = "inspection_record")
public class InspectionRecord {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "purchase_order_id", nullable = false)
    private PurchaseOrder purchaseOrder;

    @Column(name = "inspected_quantity", nullable = false)
    private Integer inspectedQuantity;

    @Column(name = "passed_quantity", nullable = false)
    private Integer passedQuantity;

    @Column(name = "rejected_quantity", nullable = false)
    private Integer rejectedQuantity;

    @Column(name = "inspection_result", nullable = false)
    @Enumerated(EnumType.STRING)
    private InspectionResult inspectionResult;

    @Column(name = "inspection_remarks")
    private String inspectionRemarks;

    @Column(name = "inspected_by")
    private String inspectedBy;

    @Column(name = "inspected_at", nullable = false)
    private LocalDateTime inspectedAt;

    @Column(name = "created_at", nullable = false, updatable = false)
    private LocalDateTime createdAt;

    @PrePersist
    protected void onCreate() {
        createdAt = LocalDateTime.now();
        if (inspectedAt == null) {
            inspectedAt = LocalDateTime.now();
        }
    }

    public enum InspectionResult {
        PASSED("合格"),
        PARTIALLY_PASSED("部分合格"),
        REJECTED("不合格");

        private final String description;

        InspectionResult(String description) {
            this.description = description;
        }

        public String getDescription() {
            return description;
        }
    }
}
