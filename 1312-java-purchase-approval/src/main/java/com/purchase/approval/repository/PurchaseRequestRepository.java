package com.purchase.approval.repository;

import com.purchase.approval.entity.PurchaseRequest;
import com.purchase.approval.enums.PurchaseRequestStatus;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

@Repository
public interface PurchaseRequestRepository extends JpaRepository<PurchaseRequest, Long> {
    Optional<PurchaseRequest> findByRequestNo(String requestNo);
    List<PurchaseRequest> findBySubmitterId(Long submitterId);
    List<PurchaseRequest> findByStatus(PurchaseRequestStatus status);
    List<PurchaseRequest> findByStatusIn(List<PurchaseRequestStatus> statuses);
}
