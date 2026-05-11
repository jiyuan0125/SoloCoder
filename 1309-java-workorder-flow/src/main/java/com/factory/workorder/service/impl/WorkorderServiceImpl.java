package com.factory.workorder.service.impl;

import com.factory.workorder.dto.*;
import com.factory.workorder.entity.Workorder;
import com.factory.workorder.entity.WorkorderLog;
import com.factory.workorder.enums.OperationType;
import com.factory.workorder.enums.WorkorderPriority;
import com.factory.workorder.enums.WorkorderStatus;
import com.factory.workorder.exception.BusinessException;
import com.factory.workorder.exception.WorkorderNotFoundException;
import com.factory.workorder.repository.WorkorderLogRepository;
import com.factory.workorder.repository.WorkorderRepository;
import com.factory.workorder.service.WorkorderService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.Duration;
import java.time.LocalDateTime;
import java.util.List;
import java.util.concurrent.atomic.AtomicLong;

@Slf4j
@Service
@RequiredArgsConstructor
public class WorkorderServiceImpl implements WorkorderService {

    private final WorkorderRepository workorderRepository;
    private final WorkorderLogRepository workorderLogRepository;
    private final AtomicLong orderSequence = new AtomicLong(0);

    @Override
    @Transactional
    public Workorder createWorkorder(CreateWorkorderRequest request) {
        WorkorderPriority priority = parsePriority(request.getPriority());

        Workorder workorder = Workorder.builder()
                .orderNo(generateOrderNo())
                .equipmentName(request.getEquipmentName())
                .faultDescription(request.getFaultDescription())
                .reporter(request.getReporter())
                .status(WorkorderStatus.PENDING_ASSIGN)
                .priority(priority)
                .build();

        workorder = workorderRepository.save(workorder);

        createLog(workorder.getId(), OperationType.CREATE,
                request.getReporter(),
                String.format("创建工单，优先级: %s", priority.getDescription()));

        log.info("工单创建成功: {}", workorder.getOrderNo());
        return workorder;
    }

    @Override
    @Transactional
    public Workorder assignWorkorder(Long workorderId, AssignWorkorderRequest request) {
        Workorder workorder = getWorkorderById(workorderId);

        if (!WorkorderStatus.PENDING_ASSIGN.equals(workorder.getStatus())) {
            throw new BusinessException(
                    String.format("工单当前状态为【%s】，无法分派，只有【待分派】状态的工单才能分派",
                            workorder.getStatus().getDescription()));
        }

        workorder.setCurrentHandler(request.getCurrentHandler());
        workorder.setStatus(WorkorderStatus.IN_PROGRESS);
        workorder = workorderRepository.save(workorder);

        createLog(workorderId, OperationType.ASSIGN, request.getOperator(),
                String.format("分派给【%s】处理", request.getCurrentHandler()));

        log.info("工单分派成功: {} -> {}", workorder.getOrderNo(), request.getCurrentHandler());
        return workorder;
    }

    @Override
    @Transactional
    public Workorder submitAcceptance(Long workorderId, SubmitAcceptanceRequest request) {
        Workorder workorder = getWorkorderById(workorderId);

        if (!WorkorderStatus.IN_PROGRESS.equals(workorder.getStatus())) {
            throw new BusinessException(
                    String.format("工单当前状态为【%s】，无法提交验收，只有【处理中】状态的工单才能提交验收",
                            workorder.getStatus().getDescription()));
        }

        workorder.setStatus(WorkorderStatus.PENDING_ACCEPTANCE);
        workorder = workorderRepository.save(workorder);

        createLog(workorderId, OperationType.SUBMIT_ACCEPTANCE, request.getOperator(), request.getRemark());

        log.info("工单提交验收: {}", workorder.getOrderNo());
        return workorder;
    }

    @Override
    @Transactional
    public Workorder acceptWorkorder(Long workorderId, AcceptanceRequest request) {
        Workorder workorder = getWorkorderById(workorderId);

        if (!WorkorderStatus.PENDING_ACCEPTANCE.equals(workorder.getStatus())) {
            throw new BusinessException(
                    String.format("工单当前状态为【%s】，无法验收通过，只有【待验收】状态的工单才能验收",
                            workorder.getStatus().getDescription()));
        }

        workorder.setStatus(WorkorderStatus.COMPLETED);
        workorder.setCompleteTime(LocalDateTime.now());
        workorder = workorderRepository.save(workorder);

        createLog(workorderId, OperationType.ACCEPT, request.getOperator(), request.getRemark());

        log.info("工单验收通过: {}", workorder.getOrderNo());
        return workorder;
    }

    @Override
    @Transactional
    public Workorder rejectWorkorder(Long workorderId, AcceptanceRequest request) {
        Workorder workorder = getWorkorderById(workorderId);

        if (!WorkorderStatus.PENDING_ACCEPTANCE.equals(workorder.getStatus())) {
            throw new BusinessException(
                    String.format("工单当前状态为【%s】，无法驳回，只有【待验收】状态的工单才能驳回",
                            workorder.getStatus().getDescription()));
        }

        workorder.setStatus(WorkorderStatus.IN_PROGRESS);
        workorder = workorderRepository.save(workorder);

        createLog(workorderId, OperationType.REJECT, request.getOperator(),
                request.getRemark() != null ? request.getRemark() : "验收驳回，退回继续处理");

        log.info("工单验收驳回: {}", workorder.getOrderNo());
        return workorder;
    }

    @Override
    @Transactional(readOnly = true)
    public List<Workorder> getWorkordersByStatus(WorkorderStatus status) {
        List<Workorder> workorders = workorderRepository.findByStatus(status);
        return checkAndUpgradePriorities(workorders);
    }

    @Override
    @Transactional(readOnly = true)
    public List<Workorder> getWorkordersByPriority(WorkorderPriority priority) {
        List<Workorder> workorders = workorderRepository.findByPriority(priority);
        return checkAndUpgradePriorities(workorders);
    }

    @Override
    @Transactional(readOnly = true)
    public List<Workorder> getWorkordersByHandler(String handler) {
        List<Workorder> workorders = workorderRepository.findByCurrentHandler(handler);
        return checkAndUpgradePriorities(workorders);
    }

    @Override
    @Transactional(readOnly = true)
    public WorkorderDetailResponse getWorkorderDetail(Long workorderId) {
        Workorder workorder = getWorkorderById(workorderId);
        checkAndUpgradePriority(workorder);
        List<WorkorderLog> logs = workorderLogRepository.findByWorkorderIdOrderByOperateTimeAsc(workorderId);
        return WorkorderDetailResponse.builder()
                .workorder(workorder)
                .logs(logs)
                .build();
    }

    @Override
    @Transactional(readOnly = true)
    public List<Workorder> getAllWorkorders() {
        List<Workorder> workorders = workorderRepository.findAll();
        return checkAndUpgradePriorities(workorders);
    }

    private List<Workorder> checkAndUpgradePriorities(List<Workorder> workorders) {
        for (Workorder workorder : workorders) {
            checkAndUpgradePriority(workorder);
        }
        return workorders;
    }

    private void checkAndUpgradePriority(Workorder workorder) {
        if (workorder.isCompleted()) {
            return;
        }

        LocalDateTime createTime = workorder.getCreateTime();
        LocalDateTime now = LocalDateTime.now();
        Duration elapsed = Duration.between(createTime, now);
        Duration timeLimit = workorder.getPriority().getTimeLimit();

        if (elapsed.compareTo(timeLimit) > 0) {
            WorkorderPriority currentPriority = workorder.getPriority();
            if (!WorkorderPriority.CRITICAL.equals(currentPriority)) {
                WorkorderPriority newPriority = currentPriority.upgrade();
                workorder.setPriority(newPriority);
                workorderRepository.save(workorder);

                createLog(workorder.getId(), OperationType.UPGRADE_PRIORITY, "SYSTEM",
                        String.format("超时升级: %s -> %s (已耗时%d小时，时限%d小时)",
                                currentPriority.getDescription(),
                                newPriority.getDescription(),
                                elapsed.toHours(),
                                timeLimit.toHours()));

                log.warn("工单超时自动升级: {} 从 {} 升级为 {}",
                        workorder.getOrderNo(), currentPriority, newPriority);
            }
        }
    }

    private Workorder getWorkorderById(Long workorderId) {
        return workorderRepository.findById(workorderId)
                .orElseThrow(() -> new WorkorderNotFoundException(workorderId));
    }

    private WorkorderPriority parsePriority(String priorityStr) {
        try {
            if ("NORMAL".equalsIgnoreCase(priorityStr) || "普通".equals(priorityStr)) {
                return WorkorderPriority.NORMAL;
            }
            if ("URGENT".equalsIgnoreCase(priorityStr) || "紧急".equals(priorityStr)) {
                return WorkorderPriority.URGENT;
            }
            if ("CRITICAL".equalsIgnoreCase(priorityStr) || "特急".equals(priorityStr)) {
                return WorkorderPriority.CRITICAL;
            }
            return WorkorderPriority.valueOf(priorityStr.toUpperCase());
        } catch (IllegalArgumentException e) {
            throw new BusinessException("无效的优先级值: " + priorityStr + "，有效值: NORMAL/普通, URGENT/紧急, CRITICAL/特急");
        }
    }

    private String generateOrderNo() {
        String date = java.time.format.DateTimeFormatter.ofPattern("yyyyMMdd").format(LocalDateTime.now());
        long seq = orderSequence.incrementAndGet();
        return String.format("WO%s%04d", date, seq);
    }

    private void createLog(Long workorderId, OperationType operationType, String operator, String remark) {
        WorkorderLog log = WorkorderLog.builder()
                .workorderId(workorderId)
                .operationType(operationType)
                .operator(operator)
                .remark(remark)
                .build();
        workorderLogRepository.save(log);
    }
}
