package com.purchase.approval.repository;

import com.purchase.approval.entity.ApprovalRecord;
import com.purchase.approval.enums.ApprovalLevel;
import com.purchase.approval.enums.ApprovalStatus;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

@Repository
public interface ApprovalRecordRepository extends JpaRepository<ApprovalRecord, Long> {
    List<ApprovalRecord> findByPurchaseRequestId(Long purchaseRequestId);
    List<ApprovalRecord> findByReturnRequestId(Long returnRequestId);
    Optional<ApprovalRecord> findByPurchaseRequestIdAndApprovalLevel(Long purchaseRequestId, ApprovalLevel approvalLevel);
    Optional<ApprovalRecord> findByReturnRequestIdAndApprovalLevel(Long returnRequestId, ApprovalLevel approvalLevel);
    List<ApprovalRecord> findByPurchaseRequestIdAndStatus(Long purchaseRequestId, ApprovalStatus status);
}
