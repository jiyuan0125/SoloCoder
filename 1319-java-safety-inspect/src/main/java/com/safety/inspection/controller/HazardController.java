package com.safety.inspection.controller;

import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.safety.inspection.common.Result;
import com.safety.inspection.dto.AssignHazardDTO;
import com.safety.inspection.dto.HazardDTO;
import com.safety.inspection.dto.RecheckDTO;
import com.safety.inspection.dto.RectificationResultDTO;
import com.safety.inspection.entity.Hazard;
import com.safety.inspection.entity.HazardUpgradeRecord;
import com.safety.inspection.service.HazardService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.security.access.prepost.PreAuthorize;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/hazards")
@RequiredArgsConstructor
public class HazardController {

    private final HazardService hazardService;

    @GetMapping("/page")
    public Result<Page<Hazard>> getHazardPage(
            @RequestParam(defaultValue = "1") int pageNum,
            @RequestParam(defaultValue = "10") int pageSize,
            @RequestParam(required = false) String hazardLevel,
            @RequestParam(required = false) String status,
            @RequestParam(required = false) Long areaId,
            @RequestParam(required = false) Long responsiblePersonId,
            @RequestParam(required = false) Long departmentId) {
        return Result.success(hazardService.getHazardPage(pageNum, pageSize, hazardLevel, status, areaId, responsiblePersonId, departmentId));
    }

    @GetMapping("/my-responsible")
    @PreAuthorize("hasRole('RESPONSIBLE') or hasRole('DEPT_LEADER')")
    public Result<List<Hazard>> getMyResponsibleHazards() {
        return Result.success(hazardService.getMyResponsibleHazards());
    }

    @GetMapping("/{id}")
    public Result<Hazard> getHazardById(@PathVariable Long id) {
        return Result.success(hazardService.getHazardById(id));
    }

    @GetMapping("/{id}/upgrade-records")
    public Result<List<HazardUpgradeRecord>> getUpgradeRecords(@PathVariable Long id) {
        return Result.success(hazardService.getUpgradeRecords(id));
    }

    @PostMapping
    @PreAuthorize("hasRole('ADMIN') or hasRole('INSPECTOR')")
    public Result<Void> createHazard(@RequestBody @Valid HazardDTO dto) {
        hazardService.createHazard(dto);
        return Result.success();
    }

    @PostMapping("/assign")
    @PreAuthorize("hasRole('ADMIN') or hasRole('DEPT_LEADER')")
    public Result<Void> assignHazard(@RequestBody @Valid AssignHazardDTO dto) {
        hazardService.assignHazard(dto);
        return Result.success();
    }

    @PostMapping("/submit-rectification")
    @PreAuthorize("hasRole('RESPONSIBLE')")
    public Result<Void> submitRectification(@RequestBody @Valid RectificationResultDTO dto) {
        hazardService.submitRectification(dto);
        return Result.success();
    }

    @PostMapping("/recheck")
    @PreAuthorize("hasRole('INSPECTOR') or hasRole('ADMIN')")
    public Result<Void> recheckHazard(@RequestBody @Valid RecheckDTO dto) {
        hazardService.recheckHazard(dto);
        return Result.success();
    }
}
