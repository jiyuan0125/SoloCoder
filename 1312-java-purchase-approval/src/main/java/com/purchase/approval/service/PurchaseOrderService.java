package com.purchase.approval.service;

import com.purchase.approval.config.PurchaseRulesConfig;
import com.purchase.approval.dto.ApprovalDTO;
import com.purchase.approval.dto.InspectionDTO;
import com.purchase.approval.dto.ReceiptDTO;
import com.purchase.approval.dto.ReturnRequestDTO;
import com.purchase.approval.entity.*;
import com.purchase.approval.enums.*;
import com.purchase.approval.exception.BusinessException;
import com.purchase.approval.repository.*;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.List;
import java.util.Optional;

@Slf4j
@Service
@RequiredArgsConstructor
public class PurchaseOrderService {

    private final PurchaseOrderRepository purchaseOrderRepository;
    private final PurchaseRequestRepository purchaseRequestRepository;
    private final QuoteRepository quoteRepository;
    private final ReceiptRecordRepository receiptRecordRepository;
    private final InspectionRecordRepository inspectionRecordRepository;
    private final ReturnRequestRepository returnRequestRepository;
    private final ApprovalRecordRepository approvalRecordRepository;
    private final PurchaseRulesConfig purchaseRulesConfig;

    @Transactional
    public PurchaseOrder createOrderFromApprovedRequest(Long requestId) {
        PurchaseRequest request = purchaseRequestRepository.findById(requestId)
            .orElseThrow(() -> new BusinessException("采购申请不存在"));

        if (request.getStatus() != PurchaseRequestStatus.APPROVED) {
            throw new BusinessException("只有审批通过的申请才能创建订单");
        }

        if (purchaseOrderRepository.findByPurchaseRequestId(requestId).isPresent()) {
            throw new BusinessException("该采购申请已创建订单");
        }

        if (request.getSelectedSupplier() == null || request.getSelectedQuoteId() == null) {
            throw new BusinessException("采购申请未选定供应商");
        }

        Quote selectedQuote = quoteRepository.findById(request.getSelectedQuoteId())
            .orElseThrow(() -> new BusinessException("选中的报价不存在"));

        PurchaseOrder order = new PurchaseOrder();
        order.setOrderNo(generateOrderNo());
        order.setPurchaseRequest(request);
        order.setSupplier(request.getSelectedSupplier());
        order.setItemName(request.getItemName());
        order.setSpecification(request.getSpecification());
        order.setOrderedQuantity(request.getQuantity());
        order.setUnitPrice(selectedQuote.getUnitPrice());
        order.setTotalAmount(selectedQuote.getTotalPrice());
        order.setReceivedQuantity(0);
        order.setInspectedQuantity(0);
        order.setReturnedQuantity(0);
        order.setStatus(OrderStatus.PENDING_CONFIRMATION);

        log.info("采购订单[{}]已创建", order.getOrderNo());
        return purchaseOrderRepository.save(order);
    }

    @Transactional
    public PurchaseOrder confirmOrder(Long orderId) {
        PurchaseOrder order = getOrder(orderId);

        if (order.getStatus() != OrderStatus.PENDING_CONFIRMATION) {
            throw new BusinessException("当前状态不允许确认订单");
        }

        order.setStatus(OrderStatus.CONFIRMED);
        log.info("采购订单[{}]已确认", order.getOrderNo());
        return purchaseOrderRepository.save(order);
    }

    @Transactional
    public PurchaseOrder shipOrder(Long orderId) {
        PurchaseOrder order = getOrder(orderId);

        if (order.getStatus() != OrderStatus.CONFIRMED) {
            throw new BusinessException("当前状态不允许发货");
        }

        order.setStatus(OrderStatus.SHIPPED);
        log.info("采购订单[{}]已发货", order.getOrderNo());
        return purchaseOrderRepository.save(order);
    }

    @Transactional
    public PurchaseOrder receivePartial(Long orderId, ReceiptDTO dto) {
        PurchaseOrder order = getOrder(orderId);

        if (order.getStatus() != OrderStatus.SHIPPED && order.getStatus() != OrderStatus.RECEIVED) {
            throw new BusinessException("当前状态不允许收货");
        }

        Integer pendingQuantity = order.getPendingQuantity();
        if (dto.getReceivedQuantity() > pendingQuantity) {
            throw new BusinessException(String.format("收货数量[%d]超过待收数量[%d]", 
                dto.getReceivedQuantity(), pendingQuantity));
        }

        ReceiptRecord receipt = new ReceiptRecord();
        receipt.setPurchaseOrder(order);
        receipt.setReceivedQuantity(dto.getReceivedQuantity());
        receipt.setReceivedBy(dto.getReceivedBy());
        receipt.setRemarks(dto.getRemarks());
        receiptRecordRepository.save(receipt);

        int newReceivedQuantity = order.getReceivedQuantity() + dto.getReceivedQuantity();
        order.setReceivedQuantity(newReceivedQuantity);

        if (order.isFullyReceived()) {
            order.setStatus(OrderStatus.RECEIVED);
            log.info("采购订单[{}]已全部收货", order.getOrderNo());
        } else {
            log.info("采购订单[{}]部分收货：{}，累计：{}", 
                order.getOrderNo(), dto.getReceivedQuantity(), newReceivedQuantity);
        }

        return purchaseOrderRepository.save(order);
    }

    @Transactional
    public PurchaseOrder inspectPartial(Long orderId, InspectionDTO dto) {
        PurchaseOrder order = getOrder(orderId);

        if (order.getStatus() != OrderStatus.RECEIVED) {
            throw new BusinessException("当前状态不允许验收");
        }

        Integer pendingInspection = order.getPendingInspectionQuantity();
        if (dto.getInspectedQuantity() > pendingInspection) {
            throw new BusinessException(String.format("验收数量[%d]超过待验收数量[%d]", 
                dto.getInspectedQuantity(), pendingInspection));
        }

        if (dto.getPassedQuantity() > dto.getInspectedQuantity()) {
            throw new BusinessException("合格数量不能超过验收数量");
        }

        int rejectedQuantity = dto.getInspectedQuantity() - dto.getPassedQuantity();
        InspectionRecord.InspectionResult result;
        if (dto.getPassedQuantity().equals(dto.getInspectedQuantity())) {
            result = InspectionRecord.InspectionResult.PASSED;
        } else if (dto.getPassedQuantity() == 0) {
            result = InspectionRecord.InspectionResult.REJECTED;
        } else {
            result = InspectionRecord.InspectionResult.PARTIALLY_PASSED;
        }

        InspectionRecord inspection = new InspectionRecord();
        inspection.setPurchaseOrder(order);
        inspection.setInspectedQuantity(dto.getInspectedQuantity());
        inspection.setPassedQuantity(dto.getPassedQuantity());
        inspection.setRejectedQuantity(rejectedQuantity);
        inspection.setInspectionResult(result);
        inspection.setInspectionRemarks(dto.getInspectionRemarks());
        inspection.setInspectedBy(dto.getInspectedBy());
        inspectionRecordRepository.save(inspection);

        int newInspectedQuantity = order.getInspectedQuantity() + dto.getInspectedQuantity();
        order.setInspectedQuantity(newInspectedQuantity);

        if (order.isFullyInspected()) {
            order.setStatus(OrderStatus.INSPECTED);
            log.info("采购订单[{}]已全部验收", order.getOrderNo());
        } else {
            log.info("采购订单[{}]部分验收：{}，累计：{}", 
                order.getOrderNo(), dto.getInspectedQuantity(), newInspectedQuantity);
        }

        return purchaseOrderRepository.save(order);
    }

    @Transactional
    public PurchaseOrder completeOrder(Long orderId) {
        PurchaseOrder order = getOrder(orderId);

        if (order.getStatus() != OrderStatus.INSPECTED) {
            throw new BusinessException("当前状态不允许完成订单");
        }

        order.setStatus(OrderStatus.COMPLETED);
        log.info("采购订单[{}]已完成", order.getOrderNo());
        return purchaseOrderRepository.save(order);
    }

    @Transactional
    public ReturnRequest createReturnRequest(ReturnRequestDTO dto) {
        PurchaseOrder order = getOrder(dto.getPurchaseOrderId());

        if (order.getStatus() != OrderStatus.INSPECTED && order.getStatus() != OrderStatus.RECEIVED) {
            throw new BusinessException("当前状态不允许发起退货");
        }

        int availableForReturn = order.getInspectedQuantity() - order.getReturnedQuantity();
        if (dto.getReturnQuantity() > availableForReturn) {
            throw new BusinessException(String.format("退货数量[%d]超过可退数量[%d]", 
                dto.getReturnQuantity(), availableForReturn));
        }

        BigDecimal returnAmount = order.getUnitPrice().multiply(BigDecimal.valueOf(dto.getReturnQuantity()));

        ReturnRequest returnRequest = new ReturnRequest();
        returnRequest.setReturnNo(generateReturnNo());
        returnRequest.setPurchaseOrder(order);
        returnRequest.setReturnQuantity(dto.getReturnQuantity());
        returnRequest.setReturnAmount(returnAmount);
        returnRequest.setReturnReason(dto.getReturnReason());
        returnRequest.setStatus(ReturnRequestStatus.DRAFT);
        returnRequest.setSubmitterId(dto.getSubmitterId());
        returnRequest.setSubmitterName(dto.getSubmitterName());

        return returnRequestRepository.save(returnRequest);
    }

    @Transactional
    public ReturnRequest submitReturnRequest(Long returnRequestId) {
        ReturnRequest returnRequest = getReturnRequest(returnRequestId);

        if (returnRequest.getStatus() != ReturnRequestStatus.DRAFT && 
            returnRequest.getStatus() != ReturnRequestStatus.REJECTED) {
            throw new BusinessException("当前状态不允许提交退货申请");
        }

        returnRequest.setStatus(ReturnRequestStatus.APPROVING);

        List<ApprovalLevel> requiredLevels = getReturnApprovalLevels(returnRequest.getReturnAmount());
        returnRequest.setCurrentApprovalLevel(requiredLevels.get(0));

        log.info("退货申请[{}]已提交，进入审批流程", returnRequest.getReturnNo());
        return returnRequestRepository.save(returnRequest);
    }

    @Transactional
    public ReturnRequest approveReturnRequest(Long returnRequestId, ApprovalLevel approvalLevel, ApprovalDTO dto) {
        ReturnRequest returnRequest = getReturnRequest(returnRequestId);

        if (returnRequest.getStatus() != ReturnRequestStatus.APPROVING) {
            throw new BusinessException("当前状态不允许审批");
        }

        if (returnRequest.getCurrentApprovalLevel() != approvalLevel) {
            throw new BusinessException(String.format("当前需[%s]审批，您的级别[%s]不对",
                returnRequest.getCurrentApprovalLevel().getDescription(), approvalLevel.getDescription()));
        }

        List<ApprovalLevel> requiredLevels = getReturnApprovalLevels(returnRequest.getReturnAmount());
        if (!requiredLevels.contains(approvalLevel)) {
            throw new BusinessException("该级别无需审批");
        }

        ApprovalRecord record = getOrCreateReturnApprovalRecord(returnRequest, approvalLevel);
        record.setApproverId(dto.getApproverId());
        record.setApproverName(dto.getApproverName());
        record.setRemarks(dto.getRemarks());

        if (dto.getApproved()) {
            record.setStatus(ApprovalStatus.APPROVED);
            approvalRecordRepository.save(record);

            int currentIndex = requiredLevels.indexOf(approvalLevel);
            if (currentIndex < requiredLevels.size() - 1) {
                returnRequest.setCurrentApprovalLevel(requiredLevels.get(currentIndex + 1));
            } else {
                returnRequest.setStatus(ReturnRequestStatus.APPROVED);
                returnRequest.setCurrentApprovalLevel(null);
                processReturnApproval(returnRequest);
            }
        } else {
            if (dto.getRejectionReason() == null || dto.getRejectionReason().trim().isEmpty()) {
                throw new BusinessException("驳回时必须填写驳回原因");
            }
            record.setStatus(ApprovalStatus.REJECTED);
            record.setRejectionReason(dto.getRejectionReason());
            approvalRecordRepository.save(record);

            returnRequest.setStatus(ReturnRequestStatus.REJECTED);
            returnRequest.setRejectionReason(dto.getRejectionReason());
        }

        return returnRequestRepository.save(returnRequest);
    }

    private void processReturnApproval(ReturnRequest returnRequest) {
        PurchaseOrder order = returnRequest.getPurchaseOrder();
        int newReturnedQuantity = order.getReturnedQuantity() + returnRequest.getReturnQuantity();
        order.setReturnedQuantity(newReturnedQuantity);
        order.setStatus(OrderStatus.RETURNED);
        purchaseOrderRepository.save(order);

        returnRequest.setStatus(ReturnRequestStatus.COMPLETED);
        log.info("退货申请[{}]审批通过，订单[{}]退货{}件", 
            returnRequest.getReturnNo(), order.getOrderNo(), returnRequest.getReturnQuantity());
    }

    private List<ApprovalLevel> getReturnApprovalLevels(BigDecimal amount) {
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

    private ApprovalRecord getOrCreateReturnApprovalRecord(ReturnRequest returnRequest, ApprovalLevel level) {
        Optional<ApprovalRecord> existing = approvalRecordRepository
            .findByReturnRequestIdAndApprovalLevel(returnRequest.getId(), level);

        if (existing.isPresent()) {
            ApprovalRecord record = existing.get();
            if (record.getStatus() == ApprovalStatus.APPROVED) {
                throw new BusinessException("该级别已审批通过，无需重复审批");
            }
            return record;
        }

        ApprovalRecord record = new ApprovalRecord();
        record.setReturnRequest(returnRequest);
        record.setApprovalLevel(level);
        record.setStatus(ApprovalStatus.PENDING);
        return approvalRecordRepository.save(record);
    }

    public PurchaseOrder getOrder(Long orderId) {
        return purchaseOrderRepository.findById(orderId)
            .orElseThrow(() -> new BusinessException("采购订单不存在"));
    }

    public ReturnRequest getReturnRequest(Long returnRequestId) {
        return returnRequestRepository.findById(returnRequestId)
            .orElseThrow(() -> new BusinessException("退货申请不存在"));
    }

    public List<PurchaseOrder> getOrdersByStatus(OrderStatus status) {
        return purchaseOrderRepository.findByStatus(status);
    }

    private String generateOrderNo() {
        String prefix = "PO";
        String date = LocalDateTime.now().format(DateTimeFormatter.ofPattern("yyyyMMdd"));
        String suffix = String.format("%04d", (int) (Math.random() * 10000));
        return prefix + date + suffix;
    }

    private String generateReturnNo() {
        String prefix = "RT";
        String date = LocalDateTime.now().format(DateTimeFormatter.ofPattern("yyyyMMdd"));
        String suffix = String.format("%04d", (int) (Math.random() * 10000));
        return prefix + date + suffix;
    }
}
