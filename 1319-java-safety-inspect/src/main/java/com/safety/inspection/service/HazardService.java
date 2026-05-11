package com.safety.inspection.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import com.safety.inspection.common.BusinessException;
import com.safety.inspection.dto.AssignHazardDTO;
import com.safety.inspection.dto.HazardDTO;
import com.safety.inspection.dto.RecheckDTO;
import com.safety.inspection.dto.RectificationResultDTO;
import com.safety.inspection.entity.Hazard;
import com.safety.inspection.entity.HazardUpgradeRecord;
import com.safety.inspection.entity.InspectionRecord;
import com.safety.inspection.entity.InspectionTask;
import com.safety.inspection.entity.User;
import com.safety.inspection.enums.HazardLevelEnum;
import com.safety.inspection.enums.HazardStatusEnum;
import com.safety.inspection.mapper.HazardMapper;
import com.safety.inspection.mapper.HazardUpgradeRecordMapper;
import com.safety.inspection.mapper.UserMapper;
import com.safety.inspection.security.CurrentUser;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.StringUtils;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.List;

@Slf4j
@Service
@RequiredArgsConstructor
public class HazardService extends ServiceImpl<HazardMapper, Hazard> {

    private final HazardMapper hazardMapper;
    private final HazardUpgradeRecordMapper upgradeRecordMapper;
    private final UserMapper userMapper;
    private final CurrentUser currentUser;
    private final NotificationService notificationService;
    private final RepeatHazardService repeatHazardService;

    public Page<Hazard> getHazardPage(int pageNum, int pageSize, String hazardLevel, String status, 
                                       Long areaId, Long responsiblePersonId, Long departmentId) {
        Page<Hazard> page = new Page<>(pageNum, pageSize);
        LambdaQueryWrapper<Hazard> wrapper = new LambdaQueryWrapper<>();

        if (StringUtils.hasText(hazardLevel)) {
            wrapper.eq(Hazard::getHazardLevel, hazardLevel);
        }
        if (StringUtils.hasText(status)) {
            wrapper.eq(Hazard::getStatus, status);
        }
        if (areaId != null) {
            wrapper.eq(Hazard::getAreaId, areaId);
        }
        if (responsiblePersonId != null) {
            wrapper.eq(Hazard::getResponsiblePersonId, responsiblePersonId);
        }
        wrapper.orderByDesc(Hazard::getCreatedAt);

        return hazardMapper.selectPage(page, wrapper);
    }

    public List<Hazard> getMyResponsibleHazards() {
        Long userId = currentUser.getCurrentUserId();
        return hazardMapper.selectList(
            new LambdaQueryWrapper<Hazard>()
                .eq(Hazard::getResponsiblePersonId, userId)
                .ne(Hazard::getStatus, HazardStatusEnum.CLOSED.getCode())
                .orderByDesc(Hazard::getCreatedAt)
        );
    }

    public Hazard getHazardById(Long id) {
        return hazardMapper.selectById(id);
    }

    @Transactional
    public void createHazard(HazardDTO dto) {
        Long userId = currentUser.getCurrentUserId();
        
        Hazard hazard = new Hazard();
        hazard.setHazardNo(generateHazardNo());
        hazard.setAreaId(dto.getAreaId());
        hazard.setHazardLevel(dto.getHazardLevel());
        hazard.setOriginalLevel(dto.getHazardLevel());
        hazard.setHazardType(dto.getHazardType());
        hazard.setLocation(dto.getLocation());
        hazard.setDescription(dto.getDescription());
        hazard.setPhotos(dto.getPhotos());
        hazard.setStatus(HazardStatusEnum.PENDING_RECTIFICATION.getCode());
        hazard.setRectificationDeadline(calculateDeadline(dto.getHazardLevel()));
        hazard.setResponsiblePersonId(dto.getResponsiblePersonId());
        hazard.setRectificationPlan(dto.getRectificationPlan());
        hazard.setCreatedBy(userId);

        hazard = repeatHazardService.checkAndMarkRepeatHazard(hazard);
        hazardMapper.insert(hazard);

        if (dto.getResponsiblePersonId() != null) {
            notificationService.createHazardAssignedNotification(hazard);
        }
    }

    @Transactional
    public void createHazardFromInspection(InspectionTask task, InspectionRecord record) {
        Hazard hazard = new Hazard();
        hazard.setHazardNo(generateHazardNo());
        hazard.setInspectionRecordId(record.getId());
        hazard.setTaskId(task.getId());
        hazard.setAreaId(task.getAreaId());
        
        String riskLevel = "LOW";
        String hazardLevel = HazardLevelEnum.GENERAL.getCode();
        if ("MEDIUM".equals(riskLevel)) {
            hazardLevel = HazardLevelEnum.LARGER.getCode();
        } else if ("HIGH".equals(riskLevel)) {
            hazardLevel = HazardLevelEnum.MAJOR.getCode();
        }

        hazard.setHazardLevel(hazardLevel);
        hazard.setOriginalLevel(hazardLevel);
        hazard.setDescription(record.getDescription());
        hazard.setPhotos(record.getPhotos());
        hazard.setStatus(HazardStatusEnum.PENDING_RECTIFICATION.getCode());
        hazard.setRectificationDeadline(calculateDeadline(hazardLevel));
        hazard.setCreatedBy(record.getCreatedBy());

        hazard = repeatHazardService.checkAndMarkRepeatHazard(hazard);
        hazardMapper.insert(hazard);

        log.info("从巡检记录创建隐患：hazardId={}, taskId={}", hazard.getId(), task.getId());
    }

    @Transactional
    public void assignHazard(AssignHazardDTO dto) {
        Hazard hazard = hazardMapper.selectById(dto.getHazardId());
        if (hazard == null) {
            throw new BusinessException("隐患不存在");
        }

        hazard.setResponsiblePersonId(dto.getResponsiblePersonId());
        hazard.setRectificationPlan(dto.getRectificationPlan());
        hazard.setStatus(HazardStatusEnum.RECTIFYING.getCode());
        hazardMapper.updateById(hazard);

        notificationService.createHazardAssignedNotification(hazard);
    }

    @Transactional
    public void submitRectification(RectificationResultDTO dto) {
        Long userId = currentUser.getCurrentUserId();
        Hazard hazard = hazardMapper.selectById(dto.getHazardId());
        
        if (hazard == null) {
            throw new BusinessException("隐患不存在");
        }
        if (!userId.equals(hazard.getResponsiblePersonId())) {
            throw new BusinessException("只有整改责任人可以提交整改结果");
        }
        if (!HazardStatusEnum.RECTIFYING.getCode().equals(hazard.getStatus())) {
            throw new BusinessException("隐患状态不允许提交整改结果");
        }

        hazard.setRectificationResult(dto.getRectificationResult());
        hazard.setRectificationPhotos(dto.getRectificationPhotos());
        hazard.setRectificationTime(LocalDateTime.now());
        hazard.setStatus(HazardStatusEnum.PENDING_RECHECK.getCode());
        hazardMapper.updateById(hazard);

        notificationService.createHazardRectifiedNotification(hazard);
    }

    @Transactional
    public void recheckHazard(RecheckDTO dto) {
        Long userId = currentUser.getCurrentUserId();
        Hazard hazard = hazardMapper.selectById(dto.getHazardId());
        
        if (hazard == null) {
            throw new BusinessException("隐患不存在");
        }
        if (userId.equals(hazard.getResponsiblePersonId())) {
            throw new BusinessException("整改责任人不能参与复检");
        }
        if (!HazardStatusEnum.PENDING_RECHECK.getCode().equals(hazard.getStatus())) {
            throw new BusinessException("隐患状态不允许复检");
        }

        hazard.setRecheckerId(userId);
        hazard.setRecheckResult(dto.getRecheckResult());
        hazard.setRecheckTime(LocalDateTime.now());
        hazard.setRecheckRemark(dto.getRecheckRemark());

        if ("PASS".equals(dto.getRecheckResult())) {
            hazard.setStatus(HazardStatusEnum.CLOSED.getCode());
            hazard.setClosedAt(LocalDateTime.now());
            notificationService.createHazardRecheckPassedNotification(hazard);
        } else {
            hazard.setStatus(HazardStatusEnum.RECTIFYING.getCode());
            hazard.setRectificationDeadline(calculateHalfDeadline(hazard.getHazardLevel(), hazard.getRectificationDeadline()));
            notificationService.createHazardRecheckFailedNotification(hazard);
        }

        hazardMapper.updateById(hazard);
    }

    @Scheduled(cron = "0 0 3 * * ?")
    @Transactional
    public void checkAndUpgradeOverdueHazards() {
        log.info("开始检查超期未整改的隐患...");
        LocalDateTime now = LocalDateTime.now();

        List<Hazard> overdueHazards = hazardMapper.selectOverdueHazards(now);

        for (Hazard hazard : overdueHazards) {
            HazardLevelEnum currentLevel = HazardLevelEnum.getByCode(hazard.getHazardLevel());
            HazardLevelEnum nextLevel = currentLevel.getNextLevel();

            if (nextLevel != null) {
                upgradeHazard(hazard, currentLevel, nextLevel, "超期未整改自动升级");
            } else {
                notificationService.createHazardOverdueNotification(hazard);
            }
        }

        log.info("超期隐患检查完成，处理 {} 条隐患", overdueHazards.size());
    }

    @Transactional
    public void upgradeHazard(Hazard hazard, HazardLevelEnum oldLevel, HazardLevelEnum newLevel, String reason) {
        log.info("升级隐患：hazardId={}, 从{}到{}", hazard.getId(), oldLevel.getDesc(), newLevel.getDesc());

        String oldLevelCode = hazard.getHazardLevel();
        
        hazard.setHazardLevel(newLevel.getCode());
        hazard.setUpgradeCount(hazard.getUpgradeCount() != null ? hazard.getUpgradeCount() + 1 : 1);
        hazard.setRectificationDeadline(calculateDeadline(newLevel.getCode()));
        hazardMapper.updateById(hazard);

        HazardUpgradeRecord record = new HazardUpgradeRecord();
        record.setHazardId(hazard.getId());
        record.setOldLevel(oldLevelCode);
        record.setNewLevel(newLevel.getCode());
        record.setReason(reason);
        record.setNewDeadline(hazard.getRectificationDeadline());
        record.setCreatedAt(LocalDateTime.now());
        upgradeRecordMapper.insert(record);

        notificationService.createHazardUpgradedNotification(hazard, oldLevelCode);
    }

    public List<HazardUpgradeRecord> getUpgradeRecords(Long hazardId) {
        return upgradeRecordMapper.selectList(
            new LambdaQueryWrapper<HazardUpgradeRecord>()
                .eq(HazardUpgradeRecord::getHazardId, hazardId)
                .orderByDesc(HazardUpgradeRecord::getCreatedAt)
        );
    }

    private String generateHazardNo() {
        return "HZ-" + LocalDateTime.now().format(DateTimeFormatter.ofPattern("yyyyMMddHHmmss")) + 
               "-" + String.format("%04d", (int)(Math.random() * 10000));
    }

    private LocalDateTime calculateDeadline(String level) {
        HazardLevelEnum levelEnum = HazardLevelEnum.getByCode(level);
        if (levelEnum == null || levelEnum.getDeadlineDays() == 0) {
            return LocalDateTime.now().plusHours(24);
        }
        return LocalDateTime.now().plusDays(levelEnum.getDeadlineDays());
    }

    private LocalDateTime calculateHalfDeadline(String level, LocalDateTime originalDeadline) {
        HazardLevelEnum levelEnum = HazardLevelEnum.getByCode(level);
        if (levelEnum == null || levelEnum.getDeadlineDays() == 0) {
            return LocalDateTime.now().plusHours(12);
        }
        int halfDays = Math.max(1, levelEnum.getDeadlineDays() / 2);
        return LocalDateTime.now().plusDays(halfDays);
    }
}
