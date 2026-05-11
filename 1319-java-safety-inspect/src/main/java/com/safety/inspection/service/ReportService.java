package com.safety.inspection.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import com.safety.inspection.dto.DashboardStatsVO;
import com.safety.inspection.dto.DepartmentStatVO;
import com.safety.inspection.dto.MonthlyReportVO;
import com.safety.inspection.entity.*;
import com.safety.inspection.enums.HazardLevelEnum;
import com.safety.inspection.enums.HazardStatusEnum;
import com.safety.inspection.enums.TaskStatusEnum;
import com.safety.inspection.mapper.*;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.math.RoundingMode;
import java.time.LocalDate;
import java.time.LocalDateTime;
import java.time.YearMonth;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

@Slf4j
@Service
@RequiredArgsConstructor
public class ReportService extends ServiceImpl<MonthlyReportMapper, MonthlyReport> {

    private final MonthlyReportMapper reportMapper;
    private final DepartmentHazardStatMapper deptStatMapper;
    private final HazardMapper hazardMapper;
    private final InspectionTaskMapper taskMapper;
    private final InspectionAreaMapper areaMapper;
    private final DepartmentMapper departmentMapper;

    @Scheduled(cron = "0 0 4 1 * ?")
    @Transactional
    public void generateMonthlyReport() {
        YearMonth lastMonth = YearMonth.now().minusMonths(1);
        log.info("开始生成月度报告：{}-{}", lastMonth.getYear(), lastMonth.getMonthValue());
        
        generateReportForMonth(lastMonth.getYear(), lastMonth.getMonthValue());
    }

    @Transactional
    public MonthlyReportVO generateReportForMonth(int year, int month) {
        YearMonth ym = YearMonth.of(year, month);
        LocalDate startDate = ym.atDay(1);
        LocalDate endDate = ym.atEndOfMonth();
        LocalDateTime startDateTime = startDate.atStartOfDay();
        LocalDateTime endDateTime = endDate.atTime(23, 59, 59);

        MonthlyReport existReport = reportMapper.selectOne(
            new LambdaQueryWrapper<MonthlyReport>()
                .eq(MonthlyReport::getReportYear, year)
                .eq(MonthlyReport::getReportMonth, month)
        );

        if (existReport != null) {
            reportMapper.deleteById(existReport.getId());
            deptStatMapper.delete(
                new LambdaQueryWrapper<DepartmentHazardStat>()
                    .eq(DepartmentHazardStat::getReportId, existReport.getId())
            );
        }

        List<Hazard> monthHazards = hazardMapper.selectList(
            new LambdaQueryWrapper<Hazard>()
                .ge(Hazard::getCreatedAt, startDateTime)
                .le(Hazard::getCreatedAt, endDateTime)
        );

        int totalHazards = monthHazards.size();
        int rectifiedCount = (int) monthHazards.stream()
            .filter(h -> HazardStatusEnum.CLOSED.getCode().equals(h.getStatus()))
            .count();
        int rectifyingCount = (int) monthHazards.stream()
            .filter(h -> HazardStatusEnum.RECTIFYING.getCode().equals(h.getStatus()) ||
                        HazardStatusEnum.PENDING_RECHECK.getCode().equals(h.getStatus()))
            .count();
        int overdueCount = (int) monthHazards.stream()
            .filter(h -> h.getRectificationDeadline() != null && 
                        h.getRectificationDeadline().isBefore(LocalDateTime.now()) &&
                        !HazardStatusEnum.CLOSED.getCode().equals(h.getStatus()))
            .count();

        int generalCount = (int) monthHazards.stream()
            .filter(h -> HazardLevelEnum.GENERAL.getCode().equals(h.getHazardLevel()))
            .count();
        int largerCount = (int) monthHazards.stream()
            .filter(h -> HazardLevelEnum.LARGER.getCode().equals(h.getHazardLevel()))
            .count();
        int majorCount = (int) monthHazards.stream()
            .filter(h -> HazardLevelEnum.MAJOR.getCode().equals(h.getHazardLevel()))
            .count();

        MonthlyReport report = new MonthlyReport();
        report.setReportYear(year);
        report.setReportMonth(month);
        report.setTotalHazards(totalHazards);
        report.setRectifiedCount(rectifiedCount);
        report.setRectifyingCount(rectifyingCount);
        report.setOverdueCount(overdueCount);
        report.setGeneralCount(generalCount);
        report.setLargerCount(largerCount);
        report.setMajorCount(majorCount);
        report.setGeneratedAt(LocalDateTime.now());
        reportMapper.insert(report);

        Map<Long, List<Hazard>> deptHazards = monthHazards.stream()
            .collect(Collectors.groupingBy(h -> {
                InspectionArea area = areaMapper.selectById(h.getAreaId());
                return area != null ? area.getDepartmentId() : null;
            }));

        List<Department> departments = departmentMapper.selectList(
            new LambdaQueryWrapper<Department>().eq(Department::getStatus, 1)
        );

        for (Department dept : departments) {
            List<Hazard> hazards = deptHazards.getOrDefault(dept.getId(), new ArrayList<>());
            int deptTotal = hazards.size();
            int deptRectified = (int) hazards.stream()
                .filter(h -> HazardStatusEnum.CLOSED.getCode().equals(h.getStatus()))
                .count();
            
            BigDecimal rate = deptTotal > 0 
                ? BigDecimal.valueOf(deptRectified * 100.0 / deptTotal).setScale(2, RoundingMode.HALF_UP)
                : BigDecimal.ZERO;

            DepartmentHazardStat stat = new DepartmentHazardStat();
            stat.setReportId(report.getId());
            stat.setDepartmentId(dept.getId());
            stat.setTotalCount(deptTotal);
            stat.setRectifiedCount(deptRectified);
            stat.setRectificationRate(rate);
            stat.setCreatedAt(LocalDateTime.now());
            deptStatMapper.insert(stat);
        }

        log.info("月度报告生成完成：{}-{}", year, month);
        return buildReportVO(report);
    }

    public MonthlyReportVO getReportByMonth(int year, int month) {
        MonthlyReport report = reportMapper.selectOne(
            new LambdaQueryWrapper<MonthlyReport>()
                .eq(MonthlyReport::getReportYear, year)
                .eq(MonthlyReport::getReportMonth, month)
        );

        if (report == null) {
            return generateReportForMonth(year, month);
        }

        return buildReportVO(report);
    }

    public Page<MonthlyReport> getReportPage(int pageNum, int pageSize) {
        Page<MonthlyReport> page = new Page<>(pageNum, pageSize);
        return reportMapper.selectPage(page, 
            new LambdaQueryWrapper<MonthlyReport>().orderByDesc(MonthlyReport::getReportYear).orderByDesc(MonthlyReport::getReportMonth));
    }

    public DashboardStatsVO getDashboardStats() {
        DashboardStatsVO stats = new DashboardStatsVO();

        stats.setPendingRectification(hazardMapper.selectCount(
            new LambdaQueryWrapper<Hazard>()
                .eq(Hazard::getStatus, HazardStatusEnum.PENDING_RECTIFICATION.getCode())
        ).intValue());

        stats.setRectifying(hazardMapper.selectCount(
            new LambdaQueryWrapper<Hazard>()
                .eq(Hazard::getStatus, HazardStatusEnum.RECTIFYING.getCode())
        ).intValue());

        stats.setPendingRecheck(hazardMapper.selectCount(
            new LambdaQueryWrapper<Hazard>()
                .eq(Hazard::getStatus, HazardStatusEnum.PENDING_RECHECK.getCode())
        ).intValue());

        stats.setOverdue(hazardMapper.selectCount(
            new LambdaQueryWrapper<Hazard>()
                .lt(Hazard::getRectificationDeadline, LocalDateTime.now())
                .ne(Hazard::getStatus, HazardStatusEnum.CLOSED.getCode())
        ).intValue());

        LocalDate today = LocalDate.now();
        stats.setTodayTasks(taskMapper.selectCount(
            new LambdaQueryWrapper<InspectionTask>()
                .eq(InspectionTask::getTaskDate, today)
        ).intValue());

        stats.setCompletedTasks(taskMapper.selectCount(
            new LambdaQueryWrapper<InspectionTask>()
                .eq(InspectionTask::getTaskDate, today)
                .eq(InspectionTask::getTaskStatus, TaskStatusEnum.COMPLETED.getCode())
        ).intValue());

        stats.setMissedTasks(taskMapper.selectCount(
            new LambdaQueryWrapper<InspectionTask>()
                .eq(InspectionTask::getTaskStatus, TaskStatusEnum.MISSED.getCode())
        ).intValue());

        return stats;
    }

    private MonthlyReportVO buildReportVO(MonthlyReport report) {
        MonthlyReportVO vo = new MonthlyReportVO();
        vo.setReportYear(report.getReportYear());
        vo.setReportMonth(report.getReportMonth());
        vo.setTotalHazards(report.getTotalHazards());
        vo.setRectifiedCount(report.getRectifiedCount());
        vo.setRectifyingCount(report.getRectifyingCount());
        vo.setOverdueCount(report.getOverdueCount());
        vo.setGeneralCount(report.getGeneralCount());
        vo.setLargerCount(report.getLargerCount());
        vo.setMajorCount(report.getMajorCount());

        if (report.getTotalHazards() != null && report.getTotalHazards() > 0) {
            vo.setRectificationRate(BigDecimal.valueOf(report.getRectifiedCount() * 100.0 / report.getTotalHazards())
                .setScale(2, RoundingMode.HALF_UP));
        } else {
            vo.setRectificationRate(BigDecimal.ZERO);
        }

        List<DepartmentHazardStat> deptStats = deptStatMapper.selectList(
            new LambdaQueryWrapper<DepartmentHazardStat>().eq(DepartmentHazardStat::getReportId, report.getId())
        );

        List<DepartmentStatVO> deptStatVOs = new ArrayList<>();
        for (DepartmentHazardStat stat : deptStats) {
            DepartmentStatVO statVO = new DepartmentStatVO();
            statVO.setDepartmentId(stat.getDepartmentId());
            
            Department dept = departmentMapper.selectById(stat.getDepartmentId());
            statVO.setDepartmentName(dept != null ? dept.getDeptName() : "未知部门");
            
            statVO.setTotalCount(stat.getTotalCount());
            statVO.setRectifiedCount(stat.getRectifiedCount());
            statVO.setRectificationRate(stat.getRectificationRate());
            statVO.setBelowTarget(stat.getRectificationRate() != null && stat.getRectificationRate().compareTo(BigDecimal.valueOf(90)) < 0);
            deptStatVOs.add(statVO);
        }
        vo.setDepartmentStats(deptStatVOs);

        return vo;
    }
}
