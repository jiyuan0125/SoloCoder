package com.purchase.approval.repository;

import com.purchase.approval.entity.PurchaseOrder;
import com.purchase.approval.enums.OrderStatus;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

@Repository
public interface PurchaseOrderRepository extends JpaRepository<PurchaseOrder, Long> {
    Optional<PurchaseOrder> findByOrderNo(String orderNo);
    Optional<PurchaseOrder> findByPurchaseRequestId(Long purchaseRequestId);
    List<PurchaseOrder> findByStatus(OrderStatus status);
    List<PurchaseOrder> findBySupplierId(Long supplierId);
}
