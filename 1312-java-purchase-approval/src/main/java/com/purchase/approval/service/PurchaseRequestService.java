package com.purchase.approval.service;

import com.purchase.approval.config.PurchaseRulesConfig;
import com.purchase.approval.dto.PurchaseRequestDTO;
import com.purchase.approval.dto.QuoteDTO;
import com.purchase.approval.entity.PurchaseRequest;
import com.purchase.approval.entity.Quote;
import com.purchase.approval.entity.Supplier;
import com.purchase.approval.enums.*;
import com.purchase.approval.exception.BusinessException;
import com.purchase.approval.repository.PurchaseRequestRepository;
import com.purchase.approval.repository.QuoteRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.List;
import java.util.stream.Collectors;

@Slf4j
@Service
@RequiredArgsConstructor
public class PurchaseRequestService {

    private final PurchaseRequestRepository purchaseRequestRepository;
    private final QuoteRepository quoteRepository;
    private final SupplierService supplierService;
    private final PurchaseRulesConfig purchaseRulesConfig;

    @Transactional
    public PurchaseRequest createDraft(PurchaseRequestDTO dto) {
        if (dto.getRecommendedSupplierId() != null) {
            supplierService.validateSupplierNotBlacklisted(dto.getRecommendedSupplierId());
        }

        PurchaseRequest request = new PurchaseRequest();
        request.setRequestNo(generateRequestNo());
        request.setItemName(dto.getItemName());
        request.setSpecification(dto.getSpecification());
        request.setQuantity(dto.getQuantity());
        request.setEstimatedUnitPrice(dto.getEstimatedUnitPrice());
        request.setExpectedDeliveryDate(dto.getExpectedDeliveryDate());
        request.setSubmitterId(dto.getSubmitterId());
        request.setSubmitterName(dto.getSubmitterName());
        request.setStatus(PurchaseRequestStatus.DRAFT);
        request.setComparisonStatus(ComparisonStatus.NOT_STARTED);
        request.setVersion(1);

        request.calculateTotalAmount();

        if (dto.getRecommendedSupplierId() != null) {
            Supplier recommendedSupplier = supplierService.getSupplier(dto.getRecommendedSupplierId());
            request.setRecommendedSupplier(recommendedSupplier);
        }

        return purchaseRequestRepository.save(request);
    }

    @Transactional
    public PurchaseRequest submitRequest(Long requestId) {
        PurchaseRequest request = getRequest(requestId);

        if (request.getStatus() != PurchaseRequestStatus.DRAFT && 
            request.getStatus() != PurchaseRequestStatus.REJECTED) {
            throw new BusinessException("当前状态不允许提交");
        }

        if (request.getRecommendedSupplier() != null) {
            supplierService.validateSupplierNotBlacklisted(request.getRecommendedSupplier().getId());
        }

        request.setStatus(PurchaseRequestStatus.SUBMITTED);

        if (purchaseRulesConfig.canSkipComparison(request.getEstimatedTotalAmount())) {
            request.setInquiryType(InquiryType.DIRECT);
            request.setComparisonStatus(ComparisonStatus.COMPLETED);
            request.setStatus(PurchaseRequestStatus.SELECTING_SUPPLIER);
            log.info("采购申请[{}]金额{}≤5000，跳过比价", request.getRequestNo(), request.getEstimatedTotalAmount());
        } else if (purchaseRulesConfig.needsTwoSuppliers(request.getEstimatedTotalAmount())) {
            request.setInquiryType(InquiryType.LIMITED);
            request.setComparisonStatus(ComparisonStatus.IN_PROGRESS);
            request.setStatus(PurchaseRequestStatus.COMPARING);
            log.info("采购申请[{}]金额5000<{}≤50000，需要至少2家供应商报价", 
                request.getRequestNo(), request.getEstimatedTotalAmount());
        } else {
            request.setInquiryType(InquiryType.OPEN);
            request.setComparisonStatus(ComparisonStatus.IN_PROGRESS);
            request.setStatus(PurchaseRequestStatus.COMPARING);
            log.info("采购申请[{}]金额{}>50000，需要至少3家供应商报价并公开询价", 
                request.getRequestNo(), request.getEstimatedTotalAmount());
        }

        return purchaseRequestRepository.save(request);
    }

    @Transactional
    public Quote addQuote(Long requestId, QuoteDTO dto) {
        PurchaseRequest request = getRequest(requestId);

        if (request.getStatus() != PurchaseRequestStatus.COMPARING) {
            throw new BusinessException("当前状态不允许添加报价");
        }

        if (request.getComparisonStatus() == ComparisonStatus.COMPLETED) {
            throw new BusinessException("比价已完成，不允许添加新报价");
        }

        supplierService.validateSupplierNotBlacklisted(dto.getSupplierId());

        boolean supplierExists = request.getQuotes().stream()
                .anyMatch(q -> q.getSupplier().getId().equals(dto.getSupplierId()) && q.getIsValid());
        if (supplierExists) {
            throw new BusinessException("该供应商已有报价，请修改或作废原有报价");
        }

        Supplier supplier = supplierService.getSupplier(dto.getSupplierId());

        Quote quote = new Quote();
        quote.setPurchaseRequest(request);
        quote.setSupplier(supplier);
        quote.setUnitPrice(dto.getUnitPrice());
        quote.setDeliveryDate(dto.getDeliveryDate());
        quote.setPaymentTerms(dto.getPaymentTerms());
        quote.setRemarks(dto.getRemarks());
        quote.setSubmittedBy(dto.getSubmittedBy());
        quote.setIsValid(true);
        quote.calculateTotalPrice(request.getQuantity());

        Quote savedQuote = quoteRepository.save(quote);

        checkComparisonCompletion(request);

        return savedQuote;
    }

    @Transactional
    public Quote updateQuote(Long requestId, Long quoteId, QuoteDTO dto) {
        PurchaseRequest request = getRequest(requestId);
        Quote quote = quoteRepository.findById(quoteId)
                .orElseThrow(() -> new BusinessException("报价不存在"));

        if (!quote.getPurchaseRequest().getId().equals(requestId)) {
            throw new BusinessException("报价不属于当前采购申请");
        }

        if (request.getStatus() != PurchaseRequestStatus.COMPARING) {
            throw new BusinessException("当前状态不允许修改报价");
        }

        if (!quote.getIsValid()) {
            throw new BusinessException("该报价已作废");
        }

        quote.setUnitPrice(dto.getUnitPrice());
        quote.setDeliveryDate(dto.getDeliveryDate());
        quote.setPaymentTerms(dto.getPaymentTerms());
        quote.setRemarks(dto.getRemarks());
        quote.calculateTotalPrice(request.getQuantity());

        return quoteRepository.save(quote);
    }

    @Transactional
    public void invalidateQuote(Long requestId, Long quoteId) {
        PurchaseRequest request = getRequest(requestId);
        Quote quote = quoteRepository.findById(quoteId)
                .orElseThrow(() -> new BusinessException("报价不存在"));

        if (!quote.getPurchaseRequest().getId().equals(requestId)) {
            throw new BusinessException("报价不属于当前采购申请");
        }

        if (request.getStatus() != PurchaseRequestStatus.COMPARING) {
            throw new BusinessException("当前状态不允许作废报价");
        }

        quote.setIsValid(false);
        quoteRepository.save(quote);

        checkComparisonCompletion(request);
    }

    private void checkComparisonCompletion(PurchaseRequest request) {
        List<Quote> validQuotes = request.getQuotes().stream()
                .filter(Quote::getIsValid)
                .collect(Collectors.toList());

        int validQuoteCount = validQuotes.size();
        int requiredCount = purchaseRulesConfig.getRequiredSupplierCount(request.getEstimatedTotalAmount());

        if (validQuoteCount >= requiredCount) {
            request.setComparisonStatus(ComparisonStatus.COMPLETED);
            request.setStatus(PurchaseRequestStatus.SELECTING_SUPPLIER);
            log.info("采购申请[{}]比价完成，有效报价数：{}", request.getRequestNo(), validQuoteCount);
        } else {
            request.setComparisonStatus(ComparisonStatus.IN_PROGRESS);
        }

        purchaseRequestRepository.save(request);
    }

    @Transactional
    public PurchaseRequest selectSupplier(Long requestId, Long quoteId) {
        PurchaseRequest request = getRequest(requestId);

        if (request.getStatus() != PurchaseRequestStatus.SELECTING_SUPPLIER) {
            throw new BusinessException("当前状态不允许选择供应商");
        }

        Quote selectedQuote = quoteRepository.findById(quoteId)
                .orElseThrow(() -> new BusinessException("报价不存在"));

        if (!selectedQuote.getPurchaseRequest().getId().equals(requestId)) {
            throw new BusinessException("报价不属于当前采购申请");
        }

        if (!selectedQuote.getIsValid()) {
            throw new BusinessException("该报价已作废，无法选择");
        }

        supplierService.validateSupplierNotBlacklisted(selectedQuote.getSupplier().getId());

        request.setSelectedSupplier(selectedQuote.getSupplier());
        request.setSelectedQuoteId(quoteId);
        request.setStatus(PurchaseRequestStatus.APPROVING);

        ApprovalLevel firstLevel = getRequiredApprovalLevels(request.getEstimatedTotalAmount()).get(0);
        request.setCurrentApprovalLevel(firstLevel);

        log.info("采购申请[{}]已选择供应商[{}]，进入审批流程", 
            request.getRequestNo(), selectedQuote.getSupplier().getSupplierName());

        return purchaseRequestRepository.save(request);
    }

    public List<ApprovalLevel> getRequiredApprovalLevels(java.math.BigDecimal amount) {
        if (amount.compareTo(purchaseRulesConfig.getSkipComparisonAmount()) <= 0) {
            return java.util.Collections.singletonList(ApprovalLevel.DEPARTMENT_MANAGER);
        } else if (amount.compareTo(purchaseRulesConfig.getTwoSupplierMinAmount()) <= 0) {
            return java.util.Arrays.asList(
                ApprovalLevel.DEPARTMENT_MANAGER,
                ApprovalLevel.FINANCE_DIRECTOR
            );
        } else {
            return java.util.Arrays.asList(
                ApprovalLevel.DEPARTMENT_MANAGER,
                ApprovalLevel.FINANCE_DIRECTOR,
                ApprovalLevel.GENERAL_MANAGER
            );
        }
    }

    public List<Quote> getValidQuotes(Long requestId) {
        return quoteRepository.findByPurchaseRequestIdAndIsValidTrue(requestId);
    }

    public PurchaseRequest getRequest(Long requestId) {
        return purchaseRequestRepository.findById(requestId)
                .orElseThrow(() -> new BusinessException("采购申请不存在"));
    }

    public List<PurchaseRequest> getRequestsBySubmitter(Long submitterId) {
        return purchaseRequestRepository.findBySubmitterId(submitterId);
    }

    private String generateRequestNo() {
        String prefix = "PR";
        String date = LocalDateTime.now().format(DateTimeFormatter.ofPattern("yyyyMMdd"));
        String suffix = String.format("%04d", (int) (Math.random() * 10000));
        return prefix + date + suffix;
    }
}
