package com.purchase.approval.service;

import com.purchase.approval.config.PurchaseRulesConfig;
import com.purchase.approval.dto.ApprovalDTO;
import com.purchase.approval.dto.ReSubmitDTO;
import com.purchase.approval.entity.ApprovalRecord;
import com.purchase.approval.entity.PurchaseRequest;
import com.purchase.approval.enums.*;
import com.purchase.approval.exception.BusinessException;
import com.purchase.approval.repository.ApprovalRecordRepository;
import com.purchase.approval.repository.PurchaseRequestRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collections;
import java.util.List;
import java.util.Optional;

@Slf4j
@Service
@RequiredArgsConstructor
public class ApprovalService {

    private final PurchaseRequestRepository purchaseRequestRepository;
    private final ApprovalRecordRepository approvalRecordRepository;
    private final SupplierService supplierService;
    private final PurchaseRulesConfig purchaseRulesConfig;

    @Transactional
    public PurchaseRequest approvePurchaseRequest(Long requestId, ApprovalLevel approvalLevel, ApprovalDTO dto) {
        PurchaseRequest request = getRequest(requestId);

        if (request.getStatus() != PurchaseRequestStatus.APPROVING) {
            throw new BusinessException("当前状态不允许审批");
        }

        if (request.getCurrentApprovalLevel() != approvalLevel) {
            throw new BusinessException(String.format("当前需[%s]审批，您的级别[%s]不对",
                request.getCurrentApprovalLevel().getDescription(), approvalLevel.getDescription()));
        }

        List<ApprovalLevel> requiredLevels = getRequiredApprovalLevels(
            request.getEstimatedTotalAmount());

        if (!requiredLevels.contains(approvalLevel)) {
            throw new BusinessException("该级别无需审批");
        }

        ApprovalRecord record = getOrCreateApprovalRecord(request, approvalLevel);
        record.setApproverId(dto.getApproverId());
        record.setApproverName(dto.getApproverName());
        record.setRemarks(dto.getRemarks());

        if (dto.getApproved()) {
            record.setStatus(ApprovalStatus.APPROVED);
            approvalRecordRepository.save(record);

            processApprovalSuccess(request, approvalLevel, requiredLevels);
        } else {
            if (dto.getRejectionReason() == null || dto.getRejectionReason().trim().isEmpty()) {
                throw new BusinessException("驳回时必须填写驳回原因");
            }
            record.setStatus(ApprovalStatus.REJECTED);
            record.setRejectionReason(dto.getRejectionReason());
            approvalRecordRepository.save(record);

            request.setStatus(PurchaseRequestStatus.REJECTED);
            request.setRejectionReason(dto.getRejectionReason());
            log.info("采购申请[{}]被[{}]驳回，原因：{}", 
                request.getRequestNo(), approvalLevel.getDescription(), dto.getRejectionReason());
        }

        return purchaseRequestRepository.save(request);
    }

    private void processApprovalSuccess(PurchaseRequest request, ApprovalLevel currentLevel, 
                                         List<ApprovalLevel> requiredLevels) {
        int currentIndex = requiredLevels.indexOf(currentLevel);
        int totalLevels = requiredLevels.size();

        if (currentIndex < totalLevels - 1) {
            ApprovalLevel nextLevel = requiredLevels.get(currentIndex + 1);
            request.setCurrentApprovalLevel(nextLevel);
            log.info("采购申请[{}]通过[{}]审批，等待[{}]审批", 
                request.getRequestNo(), currentLevel.getDescription(), nextLevel.getDescription());
        } else {
            request.setStatus(PurchaseRequestStatus.APPROVED);
            request.setCurrentApprovalLevel(null);
            log.info("采购申请[{}]已通过所有审批", request.getRequestNo());
        }
    }

    private ApprovalRecord getOrCreateApprovalRecord(PurchaseRequest request, ApprovalLevel level) {
        Optional<ApprovalRecord> existing = approvalRecordRepository
            .findByPurchaseRequestIdAndApprovalLevel(request.getId(), level);

        if (existing.isPresent()) {
            ApprovalRecord record = existing.get();
            if (record.getStatus() == ApprovalStatus.APPROVED) {
                throw new BusinessException("该级别已审批通过，无需重复审批");
            }
            return record;
        }

        ApprovalRecord record = new ApprovalRecord();
        record.setPurchaseRequest(request);
        record.setApprovalLevel(level);
        record.setStatus(ApprovalStatus.PENDING);
        return approvalRecordRepository.save(record);
    }

    @Transactional
    public PurchaseRequest reSubmitPurchaseRequest(Long rejectedRequestId, ReSubmitDTO dto) {
        PurchaseRequest rejectedRequest = getRequest(rejectedRequestId);

        if (rejectedRequest.getStatus() != PurchaseRequestStatus.REJECTED) {
            throw new BusinessException("只有被驳回的申请才能重新提交");
        }

        if (rejectedRequest.getSelectedSupplier() != null) {
            supplierService.validateSupplierNotBlacklisted(rejectedRequest.getSelectedSupplier().getId());
        }

        List<ApprovalLevel> requiredLevels = getRequiredApprovalLevels(
            rejectedRequest.getEstimatedTotalAmount());

        List<ApprovalRecord> existingApprovedRecords = approvalRecordRepository
            .findByPurchaseRequestIdAndStatus(rejectedRequestId, ApprovalStatus.APPROVED);

        List<ApprovalLevel> alreadyApprovedLevels = new ArrayList<>();
        for (ApprovalRecord record : existingApprovedRecords) {
            alreadyApprovedLevels.add(record.getApprovalLevel());
            log.info("采购申请[{}]继承已有审批记录：{}已通过", 
                rejectedRequest.getRequestNo(), record.getApprovalLevel().getDescription());
        }

        ApprovalLevel nextRequiredLevel = null;
        for (ApprovalLevel level : requiredLevels) {
            if (!alreadyApprovedLevels.contains(level)) {
                nextRequiredLevel = level;
                break;
            }
        }

        if (nextRequiredLevel == null) {
            rejectedRequest.setStatus(PurchaseRequestStatus.APPROVED);
            rejectedRequest.setCurrentApprovalLevel(null);
            log.info("采购申请[{}]重新提交，所有审批级别均已通过", rejectedRequest.getRequestNo());
        } else {
            rejectedRequest.setStatus(PurchaseRequestStatus.APPROVING);
            rejectedRequest.setCurrentApprovalLevel(nextRequiredLevel);
            rejectedRequest.setRejectionReason(null);
            log.info("采购申请[{}]重新提交，继续等待[{}]审批", 
                rejectedRequest.getRequestNo(), nextRequiredLevel.getDescription());
        }

        return purchaseRequestRepository.save(rejectedRequest);
    }

    public List<ApprovalRecord> getApprovalRecords(Long requestId) {
        return approvalRecordRepository.findByPurchaseRequestId(requestId);
    }

    public List<ApprovalLevel> getRequiredApprovalLevels(java.math.BigDecimal amount) {
        if (amount.compareTo(purchaseRulesConfig.getSkipComparisonAmount()) <= 0) {
            return Collections.singletonList(ApprovalLevel.DEPARTMENT_MANAGER);
        } else if (amount.compareTo(purchaseRulesConfig.getTwoSupplierMinAmount()) <= 0) {
            return Arrays.asList(
                ApprovalLevel.DEPARTMENT_MANAGER,
                ApprovalLevel.FINANCE_DIRECTOR
            );
        } else {
            return Arrays.asList(
                ApprovalLevel.DEPARTMENT_MANAGER,
                ApprovalLevel.FINANCE_DIRECTOR,
                ApprovalLevel.GENERAL_MANAGER
            );
        }
    }

    public PurchaseRequest getRequest(Long requestId) {
        return purchaseRequestRepository.findById(requestId)
                .orElseThrow(() -> new BusinessException("采购申请不存在"));
    }
}
