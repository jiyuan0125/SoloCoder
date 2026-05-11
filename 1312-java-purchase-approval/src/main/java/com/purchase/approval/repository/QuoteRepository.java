package com.purchase.approval.repository;

import com.purchase.approval.entity.Quote;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface QuoteRepository extends JpaRepository<Quote, Long> {
    List<Quote> findByPurchaseRequestId(Long purchaseRequestId);
    List<Quote> findByPurchaseRequestIdAndIsValidTrue(Long purchaseRequestId);
    List<Quote> findBySupplierId(Long supplierId);
}
