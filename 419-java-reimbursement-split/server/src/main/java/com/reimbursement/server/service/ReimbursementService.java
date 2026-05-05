package com.reimbursement.server.service;

import com.reimbursement.common.constant.ErrorCode;
import com.reimbursement.common.dto.*;
import com.reimbursement.common.enums.ApprovalStatus;
import com.reimbursement.common.enums.ReimbursementStatus;
import com.reimbursement.server.entity.CostCenter;
import com.reimbursement.server.entity.Reimbursement;
import com.reimbursement.server.entity.ReimbursementAllocation;
import com.reimbursement.server.repository.InMemoryDataStore;
import org.springframework.stereotype.Service;

import java.math.BigDecimal;
import java.math.RoundingMode;
import java.time.LocalDateTime;
import java.time.YearMonth;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

@Service
public class ReimbursementService {
    private static final int MAX_COST_CENTERS = 5;
    private static final int MIN_PERCENTAGE = 1;
    private static final int MAX_PERCENTAGE = 100;
    private static final int TOTAL_PERCENTAGE = 100;

    private final InMemoryDataStore dataStore;
    private final MonthlyReportService monthlyReportService;

    public ReimbursementService(InMemoryDataStore dataStore, MonthlyReportService monthlyReportService) {
        this.dataStore = dataStore;
        this.monthlyReportService = monthlyReportService;
    }

    public ApiResponse<ReimbursementDetailDTO> createReimbursement(CreateReimbursementRequest request) {
        int allocationCount = request.getAllocations().size();
        
        if (allocationCount > MAX_COST_CENTERS) {
            if (!request.isSpecialApprovalRequested()) {
                return ApiResponse.error(ErrorCode.SPECIAL_APPROVAL_REQUIRED, 
                    "成本中心数量超过" + MAX_COST_CENTERS + "个，需要申请特殊审批");
            }
        }

        int totalPercentage = request.getAllocations().stream()
                .mapToInt(CreateReimbursementRequest.AllocationItem::getPercentage)
                .sum();
        
        if (totalPercentage != TOTAL_PERCENTAGE) {
            return ApiResponse.error(ErrorCode.INVALID_PERCENTAGE_SUM, 
                "分配百分比总和必须为100%，当前为" + totalPercentage + "%");
        }

        for (CreateReimbursementRequest.AllocationItem item : request.getAllocations()) {
            if (item.getPercentage() < MIN_PERCENTAGE || item.getPercentage() > MAX_PERCENTAGE) {
                return ApiResponse.error(ErrorCode.INVALID_PERCENTAGE_RANGE, 
                    "每个成本中心的分配百分比必须在1%-100%之间");
            }
        }

        for (CreateReimbursementRequest.AllocationItem item : request.getAllocations()) {
            if (dataStore.getCostCenterById(item.getCostCenterId()) == null) {
                return ApiResponse.error(ErrorCode.COST_CENTER_NOT_FOUND, 
                    "成本中心不存在: " + item.getCostCenterId());
            }
        }

        Reimbursement reimbursement = new Reimbursement();
        reimbursement.setId("RB" + dataStore.generateId());
        reimbursement.setEmployeeId(request.getEmployeeId());
        reimbursement.setEmployeeName(request.getEmployeeName());
        reimbursement.setReimbursementType(request.getReimbursementType());
        reimbursement.setDescription(request.getDescription());
        reimbursement.setTotalAmount(request.getTotalAmount());
        reimbursement.setStatus(ReimbursementStatus.DRAFT);
        reimbursement.setCreateTime(LocalDateTime.now());
        reimbursement.setUpdateTime(LocalDateTime.now());
        reimbursement.setSpecialApprovalRequested(request.isSpecialApprovalRequested());
        reimbursement.setSpecialApprovalGranted(request.isSpecialApprovalRequested());

        List<ReimbursementAllocation> allocations = new ArrayList<>();
        for (CreateReimbursementRequest.AllocationItem item : request.getAllocations()) {
            CostCenter costCenter = dataStore.getCostCenterById(item.getCostCenterId());
            ReimbursementAllocation allocation = new ReimbursementAllocation();
            allocation.setCostCenterId(costCenter.getId());
            allocation.setCostCenterName(costCenter.getName());
            allocation.setPercentage(item.getPercentage());
            allocation.setAllocatedAmount(calculateAllocatedAmount(request.getTotalAmount(), item.getPercentage()));
            allocations.add(allocation);
        }
        reimbursement.setAllocations(allocations);

        dataStore.saveReimbursement(reimbursement);

        return ApiResponse.success("创建成功", convertToDetailDTO(reimbursement));
    }

    public ApiResponse<ReimbursementDetailDTO> submitReimbursement(String reimbursementId) {
        Reimbursement reimbursement = dataStore.getReimbursementById(reimbursementId);
        if (reimbursement == null) {
            return ApiResponse.error(ErrorCode.REIMBURSEMENT_NOT_FOUND, "报销单不存在");
        }

        if (reimbursement.getStatus() != ReimbursementStatus.DRAFT && 
            reimbursement.getStatus() != ReimbursementStatus.REJECTED) {
            return ApiResponse.error(ErrorCode.INVALID_STATUS_FOR_OPERATION, 
                "当前状态不允许提交: " + reimbursement.getStatus().getDescription());
        }

        reimbursement.setStatus(ReimbursementStatus.PENDING_APPROVAL);
        reimbursement.setUpdateTime(LocalDateTime.now());

        for (ReimbursementAllocation allocation : reimbursement.getAllocations()) {
            allocation.setApprovalStatus(ApprovalStatus.PENDING);
            allocation.setApprovalTime(null);
            allocation.setApproverId(null);
            allocation.setApproverName(null);
            allocation.setComment(null);
        }

        dataStore.saveReimbursement(reimbursement);

        return ApiResponse.success("提交成功", convertToDetailDTO(reimbursement));
    }

    public ApiResponse<ReimbursementDetailDTO> updateReimbursement(UpdateReimbursementRequest request) {
        Reimbursement reimbursement = dataStore.getReimbursementById(request.getReimbursementId());
        if (reimbursement == null) {
            return ApiResponse.error(ErrorCode.REIMBURSEMENT_NOT_FOUND, "报销单不存在");
        }

        if (reimbursement.getStatus() != ReimbursementStatus.DRAFT && 
            reimbursement.getStatus() != ReimbursementStatus.REJECTED) {
            return ApiResponse.error(ErrorCode.CANNOT_MODIFY_WHILE_APPROVING, 
                "审批过程中不能修改分配比例，需待驳回后才能修改");
        }

        if (request.getAllocations() != null && !request.getAllocations().isEmpty()) {
            int allocationCount = request.getAllocations().size();
            
            if (allocationCount > MAX_COST_CENTERS) {
                if (!reimbursement.isSpecialApprovalRequested()) {
                    return ApiResponse.error(ErrorCode.SPECIAL_APPROVAL_REQUIRED, 
                        "成本中心数量超过" + MAX_COST_CENTERS + "个，需要申请特殊审批");
                }
            }

            int totalPercentage = request.getAllocations().stream()
                    .mapToInt(CreateReimbursementRequest.AllocationItem::getPercentage)
                    .sum();
            
            if (totalPercentage != TOTAL_PERCENTAGE) {
                return ApiResponse.error(ErrorCode.INVALID_PERCENTAGE_SUM, 
                    "分配百分比总和必须为100%，当前为" + totalPercentage + "%");
            }

            for (CreateReimbursementRequest.AllocationItem item : request.getAllocations()) {
                if (item.getPercentage() < MIN_PERCENTAGE || item.getPercentage() > MAX_PERCENTAGE) {
                    return ApiResponse.error(ErrorCode.INVALID_PERCENTAGE_RANGE, 
                        "每个成本中心的分配百分比必须在1%-100%之间");
                }
            }

            for (CreateReimbursementRequest.AllocationItem item : request.getAllocations()) {
                if (dataStore.getCostCenterById(item.getCostCenterId()) == null) {
                    return ApiResponse.error(ErrorCode.COST_CENTER_NOT_FOUND, 
                        "成本中心不存在: " + item.getCostCenterId());
                }
            }

            BigDecimal totalAmount = request.getTotalAmount() != null ? 
                request.getTotalAmount() : reimbursement.getTotalAmount();

            List<ReimbursementAllocation> allocations = new ArrayList<>();
            for (CreateReimbursementRequest.AllocationItem item : request.getAllocations()) {
                CostCenter costCenter = dataStore.getCostCenterById(item.getCostCenterId());
                ReimbursementAllocation allocation = new ReimbursementAllocation();
                allocation.setCostCenterId(costCenter.getId());
                allocation.setCostCenterName(costCenter.getName());
                allocation.setPercentage(item.getPercentage());
                allocation.setAllocatedAmount(calculateAllocatedAmount(totalAmount, item.getPercentage()));
                allocations.add(allocation);
            }
            reimbursement.setAllocations(allocations);
        }

        if (request.getDescription() != null) {
            reimbursement.setDescription(request.getDescription());
        }
        if (request.getTotalAmount() != null) {
            reimbursement.setTotalAmount(request.getTotalAmount());
            if (request.getAllocations() == null || request.getAllocations().isEmpty()) {
                for (ReimbursementAllocation allocation : reimbursement.getAllocations()) {
                    allocation.setAllocatedAmount(calculateAllocatedAmount(
                        request.getTotalAmount(), allocation.getPercentage()));
                }
            }
        }

        reimbursement.setUpdateTime(LocalDateTime.now());
        dataStore.saveReimbursement(reimbursement);

        return ApiResponse.success("更新成功", convertToDetailDTO(reimbursement));
    }

    public ApiResponse<ReimbursementDetailDTO> approve(ApprovalRequest request) {
        Reimbursement reimbursement = dataStore.getReimbursementById(request.getReimbursementId());
        if (reimbursement == null) {
            return ApiResponse.error(ErrorCode.REIMBURSEMENT_NOT_FOUND, "报销单不存在");
        }

        if (reimbursement.getStatus() != ReimbursementStatus.PENDING_APPROVAL && 
            reimbursement.getStatus() != ReimbursementStatus.APPROVING) {
            return ApiResponse.error(ErrorCode.INVALID_STATUS_FOR_OPERATION, 
                "当前状态不允许审批: " + reimbursement.getStatus().getDescription());
        }

        ReimbursementAllocation targetAllocation = null;
        for (ReimbursementAllocation allocation : reimbursement.getAllocations()) {
            if (allocation.getCostCenterId().equals(request.getCostCenterId())) {
                targetAllocation = allocation;
                break;
            }
        }

        if (targetAllocation == null) {
            return ApiResponse.error(ErrorCode.COST_CENTER_NOT_FOUND, 
                "该报销单不包含成本中心: " + request.getCostCenterId());
        }

        if (targetAllocation.getApprovalStatus() != ApprovalStatus.PENDING) {
            return ApiResponse.error(ErrorCode.ALREADY_APPROVED, 
                "该成本中心已" + targetAllocation.getApprovalStatus().getDescription());
        }

        reimbursement.setStatus(ReimbursementStatus.APPROVING);

        if (request.isApproved()) {
            targetAllocation.setApprovalStatus(ApprovalStatus.APPROVED);
            targetAllocation.setApproverId(request.getApproverId());
            targetAllocation.setApproverName(request.getApproverName());
            targetAllocation.setApprovalTime(LocalDateTime.now());
            targetAllocation.setComment(request.getComment());

            boolean allApproved = reimbursement.getAllocations().stream()
                    .allMatch(a -> a.getApprovalStatus() == ApprovalStatus.APPROVED);

            if (allApproved) {
                reimbursement.setStatus(ReimbursementStatus.ALL_APPROVED);
                monthlyReportService.addAllocationToReport(reimbursement);
            }
        } else {
            for (ReimbursementAllocation allocation : reimbursement.getAllocations()) {
                allocation.setApprovalStatus(ApprovalStatus.PENDING);
                allocation.setApproverId(null);
                allocation.setApproverName(null);
                allocation.setApprovalTime(null);
                allocation.setComment(null);
            }

            targetAllocation.setApprovalStatus(ApprovalStatus.REJECTED);
            targetAllocation.setApproverId(request.getApproverId());
            targetAllocation.setApproverName(request.getApproverName());
            targetAllocation.setApprovalTime(LocalDateTime.now());
            targetAllocation.setComment(request.getComment());

            reimbursement.setStatus(ReimbursementStatus.REJECTED);
            reimbursement.setRejectReason(request.getComment() != null ? 
                request.getComment() : "成本中心" + targetAllocation.getCostCenterName() + "驳回");
        }

        reimbursement.setUpdateTime(LocalDateTime.now());
        dataStore.saveReimbursement(reimbursement);

        return ApiResponse.success(request.isApproved() ? "审批通过" : "已驳回", 
            convertToDetailDTO(reimbursement));
    }

    public ApiResponse<ReimbursementDetailDTO> finalReview(FinalReviewRequest request) {
        Reimbursement reimbursement = dataStore.getReimbursementById(request.getReimbursementId());
        if (reimbursement == null) {
            return ApiResponse.error(ErrorCode.REIMBURSEMENT_NOT_FOUND, "报销单不存在");
        }

        if (reimbursement.getStatus() != ReimbursementStatus.ALL_APPROVED) {
            return ApiResponse.error(ErrorCode.INVALID_STATUS_FOR_OPERATION, 
                "当前状态不允许财务复核: " + reimbursement.getStatus().getDescription());
        }

        if (request.isApproved()) {
            reimbursement.setStatus(ReimbursementStatus.FINAL_REVIEWED);
        } else {
            reimbursement.setStatus(ReimbursementStatus.REJECTED);
            reimbursement.setRejectReason(request.getComment() != null ? 
                request.getComment() : "财务复核驳回");
        }

        reimbursement.setUpdateTime(LocalDateTime.now());
        dataStore.saveReimbursement(reimbursement);

        return ApiResponse.success(request.isApproved() ? "财务复核通过" : "财务复核驳回", 
            convertToDetailDTO(reimbursement));
    }

    public ApiResponse<ReimbursementDetailDTO> getReimbursement(String id) {
        Reimbursement reimbursement = dataStore.getReimbursementById(id);
        if (reimbursement == null) {
            return ApiResponse.error(ErrorCode.REIMBURSEMENT_NOT_FOUND, "报销单不存在");
        }
        return ApiResponse.success(convertToDetailDTO(reimbursement));
    }

    public ApiResponse<List<ReimbursementDetailDTO>> getAllReimbursements() {
        List<ReimbursementDetailDTO> list = dataStore.getReimbursementMap().values().stream()
                .map(this::convertToDetailDTO)
                .sorted((a, b) -> b.getCreateTime().compareTo(a.getCreateTime()))
                .collect(Collectors.toList());
        return ApiResponse.success(list);
    }

    public ApiResponse<List<CostCenterDTO>> getAllCostCenters() {
        List<CostCenterDTO> dtoList = new ArrayList<>();
        for (CostCenter costCenter : dataStore.getCostCenterMap().values()) {
            CostCenterDTO dto = new CostCenterDTO();
            dto.setId(costCenter.getId());
            dto.setName(costCenter.getName());
            dto.setManagerId(costCenter.getManagerId());
            dto.setManagerName(costCenter.getManagerName());
            dto.setMonthlyBudget(costCenter.getMonthlyBudget());
            dtoList.add(dto);
        }
        return ApiResponse.success(dtoList);
    }

    private BigDecimal calculateAllocatedAmount(BigDecimal totalAmount, int percentage) {
        return totalAmount.multiply(BigDecimal.valueOf(percentage))
                .divide(BigDecimal.valueOf(100), 2, RoundingMode.HALF_UP);
    }

    private ReimbursementDetailDTO convertToDetailDTO(Reimbursement reimbursement) {
        ReimbursementDetailDTO dto = new ReimbursementDetailDTO();
        dto.setReimbursementId(reimbursement.getId());
        dto.setEmployeeId(reimbursement.getEmployeeId());
        dto.setEmployeeName(reimbursement.getEmployeeName());
        dto.setReimbursementType(reimbursement.getReimbursementType());
        dto.setDescription(reimbursement.getDescription());
        dto.setTotalAmount(reimbursement.getTotalAmount());
        dto.setStatus(reimbursement.getStatus());
        dto.setCreateTime(reimbursement.getCreateTime());
        dto.setUpdateTime(reimbursement.getUpdateTime());
        dto.setRejectReason(reimbursement.getRejectReason());

        List<CostCenterApprovalDTO> approvalDTOs = new ArrayList<>();
        for (ReimbursementAllocation allocation : reimbursement.getAllocations()) {
            CostCenterApprovalDTO approvalDTO = new CostCenterApprovalDTO();
            approvalDTO.setCostCenterId(allocation.getCostCenterId());
            approvalDTO.setCostCenterName(allocation.getCostCenterName());
            approvalDTO.setApproverId(allocation.getApproverId());
            approvalDTO.setApproverName(allocation.getApproverName());
            approvalDTO.setPercentage(allocation.getPercentage());
            approvalDTO.setAllocatedAmount(allocation.getAllocatedAmount());
            approvalDTO.setApprovalStatus(allocation.getApprovalStatus());
            approvalDTO.setApprovalTime(allocation.getApprovalTime());
            approvalDTO.setComment(allocation.getComment());
            approvalDTOs.add(approvalDTO);
        }
        dto.setAllocations(approvalDTOs);

        return dto;
    }
}
