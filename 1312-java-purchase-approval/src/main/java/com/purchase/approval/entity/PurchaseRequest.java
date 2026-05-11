package com.purchase.approval.entity;

import lombok.Data;
import javax.persistence.*;
import java.math.BigDecimal;
import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;

@Data
@Entity
@Table(name = "purchase_request")
public class PurchaseRequest {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(name = "request_no", unique = true, nullable = false)
    private String requestNo;

    @Column(name = "item_name", nullable = false)
    private String itemName;

    @Column(name = "specification")
    private String specification;

    @Column(name = "quantity", nullable = false)
    private Integer quantity;

    @Column(name = "estimated_unit_price", nullable = false, precision = 12, scale = 2)
    private BigDecimal estimatedUnitPrice;

    @Column(name = "estimated_total_amount", nullable = false, precision = 12, scale = 2)
    private BigDecimal estimatedTotalAmount;

    @Column(name = "expected_delivery_date")
    private LocalDate expectedDeliveryDate;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "recommended_supplier_id")
    private Supplier recommendedSupplier;

    @Column(name = "inquiry_type")
    @Enumerated(EnumType.STRING)
    private com.purchase.approval.enums.InquiryType inquiryType;

    @Column(name = "status", nullable = false)
    @Enumerated(EnumType.STRING)
    private com.purchase.approval.enums.PurchaseRequestStatus status;

    @Column(name = "comparison_status")
    @Enumerated(EnumType.STRING)
    private com.purchase.approval.enums.ComparisonStatus comparisonStatus;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "selected_supplier_id")
    private Supplier selectedSupplier;

    @Column(name = "selected_quote_id")
    private Long selectedQuoteId;

    @Column(name = "rejection_reason")
    private String rejectionReason;

    @Column(name = "current_approval_level")
    @Enumerated(EnumType.STRING)
    private com.purchase.approval.enums.ApprovalLevel currentApprovalLevel;

    @Column(name = "version")
    private Integer version = 1;

    @Column(name = "parent_request_id")
    private Long parentRequestId;

    @Column(name = "submitter_id", nullable = false)
    private Long submitterId;

    @Column(name = "submitter_name", nullable = false)
    private String submitterName;

    @Column(name = "created_at", nullable = false, updatable = false)
    private LocalDateTime createdAt;

    @Column(name = "updated_at")
    private LocalDateTime updatedAt;

    @OneToMany(mappedBy = "purchaseRequest", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<Quote> quotes = new ArrayList<>();

    @OneToMany(mappedBy = "purchaseRequest", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<ApprovalRecord> approvalRecords = new ArrayList<>();

    @PrePersist
    protected void onCreate() {
        createdAt = LocalDateTime.now();
        updatedAt = LocalDateTime.now();
    }

    @PreUpdate
    protected void onUpdate() {
        updatedAt = LocalDateTime.now();
    }

    public void calculateTotalAmount() {
        if (estimatedUnitPrice != null && quantity != null) {
            this.estimatedTotalAmount = estimatedUnitPrice.multiply(BigDecimal.valueOf(quantity));
        }
    }
}
