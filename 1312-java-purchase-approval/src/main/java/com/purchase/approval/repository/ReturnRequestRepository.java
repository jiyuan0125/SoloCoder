package com.purchase.approval.repository;

import com.purchase.approval.entity.ReturnRequest;
import com.purchase.approval.enums.ReturnRequestStatus;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

@Repository
public interface ReturnRequestRepository extends JpaRepository<ReturnRequest, Long> {
    Optional<ReturnRequest> findByReturnNo(String returnNo);
    List<ReturnRequest> findByPurchaseOrderId(Long purchaseOrderId);
    List<ReturnRequest> findByStatus(ReturnRequestStatus status);
    List<ReturnRequest> findBySubmitterId(Long submitterId);
}
