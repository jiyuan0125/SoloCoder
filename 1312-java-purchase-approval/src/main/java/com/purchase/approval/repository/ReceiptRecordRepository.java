package com.purchase.approval.repository;

import com.purchase.approval.entity.ReceiptRecord;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface ReceiptRecordRepository extends JpaRepository<ReceiptRecord, Long> {
    List<ReceiptRecord> findByPurchaseOrderId(Long purchaseOrderId);
}
