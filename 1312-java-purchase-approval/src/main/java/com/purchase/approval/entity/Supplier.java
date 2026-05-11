package com.purchase.approval.entity;

import lombok.Data;
import javax.persistence.*;
import java.time.LocalDate;
import java.time.LocalDateTime;

@Data
@Entity
@Table(name = "supplier")
public class Supplier {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(name = "supplier_code", unique = true, nullable = false)
    private String supplierCode;

    @Column(name = "supplier_name", nullable = false)
    private String supplierName;

    @Column(name = "contact_person")
    private String contactPerson;

    @Column(name = "phone")
    private String phone;

    @Column(name = "email")
    private String email;

    @Column(name = "address")
    private String address;

    @Column(name = "is_blacklisted", nullable = false)
    private Boolean isBlacklisted = false;

    @Column(name = "blacklist_reason")
    @Enumerated(EnumType.STRING)
    private com.purchase.approval.enums.BlacklistReason blacklistReason;

    @Column(name = "blacklist_start_date")
    private LocalDate blacklistStartDate;

    @Column(name = "blacklist_end_date")
    private LocalDate blacklistEndDate;

    @Column(name = "blacklist_remarks")
    private String blacklistRemarks;

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

    public boolean isCurrentlyBlacklisted() {
        if (!isBlacklisted) {
            return false;
        }
        if (blacklistEndDate == null) {
            return true;
        }
        return LocalDate.now().isBefore(blacklistEndDate) || LocalDate.now().isEqual(blacklistEndDate);
    }
}
