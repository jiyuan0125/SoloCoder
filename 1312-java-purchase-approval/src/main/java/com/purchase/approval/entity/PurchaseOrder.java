package com.purchase.approval.entity;

import lombok.Data;
import javax.persistence.*;
import java.math.BigDecimal;
import java.time.LocalDateTime;

@Data
@Entity
@Table(name = "purchase_order")
public class PurchaseOrder {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(name = "order_no", unique = true, nullable = false)
    private String orderNo;

    @OneToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "purchase_request_id", unique = true, nullable = false)
    private PurchaseRequest purchaseRequest;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "supplier_id", nullable = false)
    private Supplier supplier;

    @Column(name = "item_name", nullable = false)
    private String itemName;

    @Column(name = "specification")
    private String specification;

    @Column(name = "ordered_quantity", nullable = false)
    private Integer orderedQuantity;

    @Column(name = "unit_price", nullable = false, precision = 12, scale = 2)
    private BigDecimal unitPrice;

    @Column(name = "total_amount", nullable = false, precision = 12, scale = 2)
    private BigDecimal totalAmount;

    @Column(name = "received_quantity", nullable = false)
    private Integer receivedQuantity = 0;

    @Column(name = "inspected_quantity", nullable = false)
    private Integer inspectedQuantity = 0;

    @Column(name = "returned_quantity")
    private Integer returnedQuantity = 0;

    @Column(name = "status", nullable = false)
    @Enumerated(EnumType.STRING)
    private com.purchase.approval.enums.OrderStatus status;

    @Column(name = "remarks")
    private String remarks;

    @Column(name = "created_at", nullable = false, updatable = false)
    private LocalDateTime createdAt;

    @Column(name = "updated_at")
    private LocalDateTime updatedAt;

    @PrePersist
    protected void onCreate() {
        createdAt = LocalDateTime.now();
        updatedAt = LocalDateTime.now();
    }

    @PreUpdate
    protected void onUpdate() {
        updatedAt = LocalDateTime.now();
    }

    public boolean isFullyReceived() {
        return receivedQuantity != null && orderedQuantity != null && receivedQuantity >= orderedQuantity;
    }

    public boolean isFullyInspected() {
        return inspectedQuantity != null && orderedQuantity != null && inspectedQuantity >= orderedQuantity;
    }

    public Integer getPendingQuantity() {
        if (orderedQuantity == null || receivedQuantity == null) {
            return orderedQuantity;
        }
        return orderedQuantity - receivedQuantity;
    }

    public Integer getPendingInspectionQuantity() {
        if (receivedQuantity == null || inspectedQuantity == null) {
            return receivedQuantity;
        }
        return receivedQuantity - inspectedQuantity;
    }
}
