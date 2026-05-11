package com.safety.inspection.controller;

import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.safety.inspection.common.Result;
import com.safety.inspection.dto.InspectionAreaDTO;
import com.safety.inspection.entity.InspectionArea;
import com.safety.inspection.service.InspectionAreaService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.security.access.prepost.PreAuthorize;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/inspection-areas")
@RequiredArgsConstructor
public class InspectionAreaController {

    private final InspectionAreaService areaService;

    @GetMapping("/list")
    public Result<List<InspectionArea>> getAllAreas(
            @RequestParam(required = false) String areaName,
            @RequestParam(required = false) String areaType) {
        return Result.success(areaService.getAllAreas(areaName, areaType));
    }

    @GetMapping("/page")
    @PreAuthorize("hasRole('ADMIN')")
    public Result<Page<InspectionArea>> getAreaPage(
            @RequestParam(defaultValue = "1") int pageNum,
            @RequestParam(defaultValue = "10") int pageSize,
            @RequestParam(required = false) String areaName,
            @RequestParam(required = false) String areaType,
            @RequestParam(required = false) Long departmentId) {
        return Result.success(areaService.getAreaPage(pageNum, pageSize, areaName, areaType, departmentId));
    }

    @GetMapping("/{id}")
    public Result<InspectionArea> getAreaById(@PathVariable Long id) {
        return Result.success(areaService.getAreaById(id));
    }

    @PostMapping
    @PreAuthorize("hasRole('ADMIN')")
    public Result<Void> createArea(@RequestBody @Valid InspectionAreaDTO dto) {
        areaService.createArea(dto);
        return Result.success();
    }

    @PutMapping
    @PreAuthorize("hasRole('ADMIN')")
    public Result<Void> updateArea(@RequestBody @Valid InspectionAreaDTO dto) {
        areaService.updateArea(dto);
        return Result.success();
    }

    @DeleteMapping("/{id}")
    @PreAuthorize("hasRole('ADMIN')")
    public Result<Void> deleteArea(@PathVariable Long id) {
        areaService.deleteArea(id);
        return Result.success();
    }
}
