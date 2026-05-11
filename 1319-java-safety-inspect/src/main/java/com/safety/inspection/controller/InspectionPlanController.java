package com.safety.inspection.controller;

import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.safety.inspection.common.Result;
import com.safety.inspection.dto.InspectionPlanDTO;
import com.safety.inspection.entity.InspectionPlan;
import com.safety.inspection.service.InspectionPlanService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.security.access.prepost.PreAuthorize;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/inspection-plans")
@RequiredArgsConstructor
public class InspectionPlanController {

    private final InspectionPlanService planService;

    @GetMapping("/page")
    @PreAuthorize("hasRole('ADMIN') or hasRole('INSPECTOR')")
    public Result<Page<InspectionPlan>> getPlanPage(
            @RequestParam(defaultValue = "1") int pageNum,
            @RequestParam(defaultValue = "10") int pageSize,
            @RequestParam(required = false) String planName,
            @RequestParam(required = false) Long areaId,
            @RequestParam(required = false) Long inspectorId,
            @RequestParam(required = false) Integer status) {
        return Result.success(planService.getPlanPage(pageNum, pageSize, planName, areaId, inspectorId, status));
    }

    @GetMapping("/active")
    public Result<List<InspectionPlan>> getActivePlans() {
        return Result.success(planService.getActivePlans());
    }

    @GetMapping("/{id}")
    public Result<InspectionPlan> getPlanById(@PathVariable Long id) {
        return Result.success(planService.getPlanById(id));
    }

    @PostMapping
    @PreAuthorize("hasRole('ADMIN')")
    public Result<Void> createPlan(@RequestBody @Valid InspectionPlanDTO dto) {
        planService.createPlan(dto);
        return Result.success();
    }

    @PutMapping
    @PreAuthorize("hasRole('ADMIN')")
    public Result<Void> updatePlan(@RequestBody @Valid InspectionPlanDTO dto) {
        planService.updatePlan(dto);
        return Result.success();
    }

    @DeleteMapping("/{id}")
    @PreAuthorize("hasRole('ADMIN')")
    public Result<Void> deletePlan(@PathVariable Long id) {
        planService.deletePlan(id);
        return Result.success();
    }

    @PutMapping("/{id}/toggle-status")
    @PreAuthorize("hasRole('ADMIN')")
    public Result<Void> togglePlanStatus(@PathVariable Long id) {
        planService.togglePlanStatus(id);
        return Result.success();
    }
}
