package com.safety.inspection.controller;

import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.safety.inspection.common.Result;
import com.safety.inspection.dto.SubmitInspectionDTO;
import com.safety.inspection.entity.InspectionRecord;
import com.safety.inspection.entity.InspectionTask;
import com.safety.inspection.service.InspectionTaskService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.format.annotation.DateTimeFormat;
import org.springframework.security.access.prepost.PreAuthorize;
import org.springframework.web.bind.annotation.*;

import java.time.LocalDate;
import java.util.List;

@RestController
@RequestMapping("/api/inspection-tasks")
@RequiredArgsConstructor
public class InspectionTaskController {

    private final InspectionTaskService taskService;

    @GetMapping("/page")
    @PreAuthorize("hasRole('ADMIN')")
    public Result<Page<InspectionTask>> getTaskPage(
            @RequestParam(defaultValue = "1") int pageNum,
            @RequestParam(defaultValue = "10") int pageSize,
            @RequestParam(required = false) Long inspectorId,
            @RequestParam(required = false) String taskStatus,
            @RequestParam(required = false) @DateTimeFormat(pattern = "yyyy-MM-dd") LocalDate startDate,
            @RequestParam(required = false) @DateTimeFormat(pattern = "yyyy-MM-dd") LocalDate endDate) {
        return Result.success(taskService.getTaskPage(pageNum, pageSize, inspectorId, taskStatus, startDate, endDate));
    }

    @GetMapping("/my-tasks")
    @PreAuthorize("hasRole('INSPECTOR')")
    public Result<List<InspectionTask>> getMyTasks() {
        return Result.success(taskService.getMyTasks());
    }

    @GetMapping("/{id}")
    public Result<InspectionTask> getTaskById(@PathVariable Long id) {
        return Result.success(taskService.getTaskById(id));
    }

    @GetMapping("/{id}/records")
    public Result<List<InspectionRecord>> getTaskRecords(@PathVariable Long id) {
        return Result.success(taskService.getTaskRecords(id));
    }

    @PostMapping("/{id}/start")
    @PreAuthorize("hasRole('INSPECTOR')")
    public Result<Void> startTask(@PathVariable Long id) {
        taskService.startTask(id);
        return Result.success();
    }

    @PostMapping("/{id}/submit")
    @PreAuthorize("hasRole('INSPECTOR')")
    public Result<Void> submitInspection(@PathVariable Long id, @RequestBody @Valid SubmitInspectionDTO dto) {
        taskService.submitInspection(id, dto);
        return Result.success();
    }
}
