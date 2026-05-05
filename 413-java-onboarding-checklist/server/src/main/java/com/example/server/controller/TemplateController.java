package com.example.server.controller;

import com.example.common.dto.ChecklistTemplateDTO;
import com.example.common.enums.ErrorCode;
import com.example.common.request.CreateTemplateRequest;
import com.example.common.response.ApiResponse;
import com.example.server.service.TemplateService;
import jakarta.validation.Valid;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import java.util.List;
import java.util.Optional;

@RestController
@RequestMapping("/api/templates")
public class TemplateController {

    @Autowired
    private TemplateService templateService;

    @PostMapping
    public ApiResponse<ChecklistTemplateDTO> createTemplate(@Valid @RequestBody CreateTemplateRequest request) {
        ChecklistTemplateDTO template = templateService.createTemplate(request);
        return ApiResponse.success(template);
    }

    @GetMapping
    public ApiResponse<List<ChecklistTemplateDTO>> getAllTemplates() {
        List<ChecklistTemplateDTO> templates = templateService.getAllTemplates();
        return ApiResponse.success(templates);
    }

    @GetMapping("/{id}")
    public ApiResponse<ChecklistTemplateDTO> getTemplateById(@PathVariable String id) {
        Optional<ChecklistTemplateDTO> template = templateService.getTemplateById(id);
        if (template.isPresent()) {
            return ApiResponse.success(template.get());
        }
        return ApiResponse.error(ErrorCode.TEMPLATE_NOT_FOUND);
    }

    @DeleteMapping("/{id}")
    public ApiResponse<Void> deleteTemplate(@PathVariable String id) {
        boolean deleted = templateService.deleteTemplate(id);
        if (deleted) {
            return ApiResponse.success();
        }
        return ApiResponse.error(ErrorCode.TEMPLATE_NOT_FOUND);
    }

    @PostMapping("/{sourceId}/copy")
    public ApiResponse<ChecklistTemplateDTO> copyTemplate(
            @PathVariable String sourceId,
            @RequestParam String newName) {
        Optional<ChecklistTemplateDTO> copiedTemplate = templateService.copyTemplate(sourceId, newName);
        if (copiedTemplate.isPresent()) {
            return ApiResponse.success(copiedTemplate.get());
        }
        return ApiResponse.error(ErrorCode.TEMPLATE_NOT_FOUND, "源模板不存在或新名称已存在");
    }
}
