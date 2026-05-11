package com.exam.controller;

import com.exam.dto.ApiResponse;
import com.exam.dto.KnowledgeCategoryDTO;
import com.exam.entity.KnowledgeCategory;
import com.exam.service.KnowledgeCategoryService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/categories")
@RequiredArgsConstructor
@Tag(name = "知识点分类管理", description = "知识点分类的增删改查接口")
public class KnowledgeCategoryController {

    private final KnowledgeCategoryService categoryService;

    @GetMapping
    @Operation(summary = "获取所有分类", description = "按排序顺序获取所有知识点分类")
    public ApiResponse<List<KnowledgeCategory>> getAllCategories() {
        return ApiResponse.success(categoryService.getAllCategories());
    }

    @GetMapping("/{id}")
    @Operation(summary = "获取单个分类", description = "根据ID获取知识点分类详情")
    public ApiResponse<KnowledgeCategory> getCategoryById(@PathVariable Long id) {
        return ApiResponse.success(categoryService.getCategoryById(id));
    }

    @PostMapping
    @Operation(summary = "创建分类", description = "创建新的知识点分类")
    public ApiResponse<KnowledgeCategory> createCategory(@Valid @RequestBody KnowledgeCategoryDTO dto) {
        return ApiResponse.success("创建成功", categoryService.createCategory(dto));
    }

    @PutMapping("/{id}")
    @Operation(summary = "更新分类", description = "更新知识点分类信息")
    public ApiResponse<KnowledgeCategory> updateCategory(
            @PathVariable Long id,
            @Valid @RequestBody KnowledgeCategoryDTO dto) {
        return ApiResponse.success("更新成功", categoryService.updateCategory(id, dto));
    }

    @DeleteMapping("/{id}")
    @Operation(summary = "删除分类", description = "删除知识点分类（分类下无题目时才能删除）")
    public ApiResponse<Void> deleteCategory(@PathVariable Long id) {
        categoryService.deleteCategory(id);
        return ApiResponse.success("删除成功", null);
    }
}
