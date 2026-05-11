package com.safety.inspection.controller;

import com.safety.inspection.common.Result;
import com.safety.inspection.dto.InspectionCheckItemDTO;
import com.safety.inspection.entity.InspectionCheckItem;
import com.safety.inspection.service.InspectionCheckItemService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.security.access.prepost.PreAuthorize;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/check-items")
@RequiredArgsConstructor
public class InspectionCheckItemController {

    private final InspectionCheckItemService checkItemService;

    @GetMapping("/by-area/{areaId}")
    public Result<List<InspectionCheckItem>> getItemsByArea(@PathVariable Long areaId) {
        return Result.success(checkItemService.getItemsByArea(areaId));
    }

    @GetMapping("/{id}")
    public Result<InspectionCheckItem> getItemById(@PathVariable Long id) {
        return Result.success(checkItemService.getItemById(id));
    }

    @PostMapping
    @PreAuthorize("hasRole('ADMIN')")
    public Result<Void> createCheckItem(@RequestBody @Valid InspectionCheckItemDTO dto) {
        checkItemService.createCheckItem(dto);
        return Result.success();
    }

    @PostMapping("/batch")
    @PreAuthorize("hasRole('ADMIN')")
    public Result<Void> batchCreateCheckItems(@RequestBody @Valid List<InspectionCheckItemDTO> items) {
        checkItemService.batchCreateCheckItems(items);
        return Result.success();
    }

    @PutMapping
    @PreAuthorize("hasRole('ADMIN')")
    public Result<Void> updateCheckItem(@RequestBody @Valid InspectionCheckItemDTO dto) {
        checkItemService.updateCheckItem(dto);
        return Result.success();
    }

    @DeleteMapping("/{id}")
    @PreAuthorize("hasRole('ADMIN')")
    public Result<Void> deleteCheckItem(@PathVariable Long id) {
        checkItemService.deleteCheckItem(id);
        return Result.success();
    }
}
