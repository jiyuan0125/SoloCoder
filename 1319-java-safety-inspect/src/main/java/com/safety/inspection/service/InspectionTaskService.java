package com.safety.inspection.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import com.safety.inspection.common.BusinessException;
import com.safety.inspection.dto.InspectionRecordDTO;
import com.safety.inspection.dto.SubmitInspectionDTO;
import com.safety.inspection.entity.InspectionPlan;
import com.safety.inspection.entity.InspectionRecord;
import com.safety.inspection.entity.InspectionTask;
import com.safety.inspection.enums.TaskStatusEnum;
import com.safety.inspection.mapper.InspectionPlanMapper;
import com.safety.inspection.mapper.InspectionRecordMapper;
import com.safety.inspection.mapper.InspectionTaskMapper;
import com.safety.inspection.security.CurrentUser;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDate;
import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.List;

@Slf4j
@Service
@RequiredArgsConstructor
public class InspectionTaskService extends ServiceImpl<InspectionTaskMapper, InspectionTask> {

    private final InspectionTaskMapper taskMapper;
    private final InspectionPlanMapper planMapper;
    private final InspectionRecordMapper recordMapper;
    private final CurrentUser currentUser;
    private final NotificationService notificationService;
    private final HazardService hazardService;

    @Scheduled(cron = "0 0 1 * * ?")
    public void generateDailyTasks() {
        log.info("开始生成今日巡检任务...");
        LocalDate today = LocalDate.now();

        List<InspectionPlan> activePlans = planMapper.selectList(
            new LambdaQueryWrapper<InspectionPlan>()
                .eq(InspectionPlan::getStatus, 1)
                .le(InspectionPlan::getStartDate, today)
                .and(w -> w.isNull(InspectionPlan::getEndDate).or().ge(InspectionPlan::getEndDate, today))
        );

        for (InspectionPlan plan : activePlans) {
            if (shouldGenerateTask(plan, today)) {
                createTaskIfNotExists(plan, today);
            }
        }

        log.info("今日巡检任务生成完成");
    }

    @Scheduled(cron = "0 0 2 * * ?")
    public void checkMissedTasks() {
        log.info("开始检查漏检任务...");
        LocalDate today = LocalDate.now();

        List<InspectionTask> pendingTasks = taskMapper.selectOverduePendingTasks(today);

        for (InspectionTask task : pendingTasks) {
            task.setTaskStatus(TaskStatusEnum.MISSED.getCode());
            taskMapper.updateById(task);

            notificationService.createMissedTaskNotification(task);
        }

        log.info("漏检任务检查完成，处理 {} 条任务", pendingTasks.size());
    }

    private boolean shouldGenerateTask(InspectionPlan plan, LocalDate date) {
        LocalDate startDate = plan.getStartDate();
        int frequencyDays = plan.getFrequencyDays() != null ? plan.getFrequencyDays() : 1;

        long daysBetween = java.time.temporal.ChronoUnit.DAYS.between(startDate, date);
        return daysBetween >= 0 && daysBetween % frequencyDays == 0;
    }

    private void createTaskIfNotExists(InspectionPlan plan, LocalDate date) {
        Long count = taskMapper.selectCount(
            new LambdaQueryWrapper<InspectionTask>()
                .eq(InspectionTask::getPlanId, plan.getId())
                .eq(InspectionTask::getTaskDate, date)
        );

        if (count == 0) {
            InspectionTask task = new InspectionTask();
            task.setTaskNo(generateTaskNo(date));
            task.setPlanId(plan.getId());
            task.setAreaId(plan.getAreaId());
            task.setInspectorId(plan.getInspectorId());
            task.setTaskDate(date);
            task.setTaskStatus(TaskStatusEnum.PENDING.getCode());
            taskMapper.insert(task);

            notificationService.createTaskAssignedNotification(task);
        }
    }

    private String generateTaskNo(LocalDate date) {
        return "TASK-" + date.format(DateTimeFormatter.ofPattern("yyyyMMdd")) + "-" +
               String.format("%04d", System.currentTimeMillis() % 10000);
    }

    public Page<InspectionTask> getTaskPage(int pageNum, int pageSize, Long inspectorId, String taskStatus, LocalDate startDate, LocalDate endDate) {
        Page<InspectionTask> page = new Page<>(pageNum, pageSize);
        LambdaQueryWrapper<InspectionTask> wrapper = new LambdaQueryWrapper<>();

        if (inspectorId != null) {
            wrapper.eq(InspectionTask::getInspectorId, inspectorId);
        }
        if (taskStatus != null) {
            wrapper.eq(InspectionTask::getTaskStatus, taskStatus);
        }
        if (startDate != null) {
            wrapper.ge(InspectionTask::getTaskDate, startDate);
        }
        if (endDate != null) {
            wrapper.le(InspectionTask::getTaskDate, endDate);
        }
        wrapper.orderByDesc(InspectionTask::getTaskDate);

        return taskMapper.selectPage(page, wrapper);
    }

    public List<InspectionTask> getMyTasks() {
        Long userId = currentUser.getCurrentUserId();
        return taskMapper.selectList(
            new LambdaQueryWrapper<InspectionTask>()
                .eq(InspectionTask::getInspectorId, userId)
                .orderByDesc(InspectionTask::getTaskDate)
        );
    }

    public InspectionTask getTaskById(Long id) {
        return taskMapper.selectById(id);
    }

    public List<InspectionRecord> getTaskRecords(Long taskId) {
        return recordMapper.selectList(
            new LambdaQueryWrapper<InspectionRecord>()
                .eq(InspectionRecord::getTaskId, taskId)
        );
    }

    @Transactional
    public void startTask(Long taskId) {
        InspectionTask task = taskMapper.selectById(taskId);
        if (task == null) {
            throw new BusinessException("任务不存在");
        }
        if (!TaskStatusEnum.PENDING.getCode().equals(task.getTaskStatus())) {
            throw new BusinessException("任务状态不允许开始");
        }

        task.setTaskStatus(TaskStatusEnum.IN_PROGRESS.getCode());
        task.setStartTime(LocalDateTime.now());
        taskMapper.updateById(task);
    }

    @Transactional
    public void submitInspection(Long taskId, SubmitInspectionDTO dto) {
        InspectionTask task = taskMapper.selectById(taskId);
        if (task == null) {
            throw new BusinessException("任务不存在");
        }
        if (!TaskStatusEnum.IN_PROGRESS.getCode().equals(task.getTaskStatus()) &&
            !TaskStatusEnum.PENDING.getCode().equals(task.getTaskStatus())) {
            throw new BusinessException("任务状态不允许提交");
        }

        Long userId = currentUser.getCurrentUserId();

        for (InspectionRecordDTO recordDTO : dto.getRecords()) {
            InspectionRecord record = new InspectionRecord();
            record.setTaskId(taskId);
            record.setCheckItemId(recordDTO.getCheckItemId());
            record.setCheckResult(recordDTO.getCheckResult());
            record.setDescription(recordDTO.getDescription());
            record.setPhotos(recordDTO.getPhotos());
            record.setCreatedBy(userId);
            record.setCreatedAt(LocalDateTime.now());
            recordMapper.insert(record);

            if ("ABNORMAL".equals(recordDTO.getCheckResult())) {
                hazardService.createHazardFromInspection(task, record);
            }
        }

        task.setTaskStatus(TaskStatusEnum.COMPLETED.getCode());
        task.setEndTime(LocalDateTime.now());
        task.setRemark(dto.getRemark());
        taskMapper.updateById(task);
    }
}
