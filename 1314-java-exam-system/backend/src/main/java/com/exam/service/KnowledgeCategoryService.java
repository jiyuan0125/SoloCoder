package com.exam.service;

import com.exam.dto.KnowledgeCategoryDTO;
import com.exam.entity.KnowledgeCategory;
import com.exam.repository.KnowledgeCategoryRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;

@Service
@RequiredArgsConstructor
@Transactional
public class KnowledgeCategoryService {

    private final KnowledgeCategoryRepository categoryRepository;

    public List<KnowledgeCategory> getAllCategories() {
        return categoryRepository.findAllByOrderBySortOrderAsc();
    }

    public KnowledgeCategory getCategoryById(Long id) {
        return categoryRepository.findById(id)
                .orElseThrow(() -> new RuntimeException("分类不存在: " + id));
    }

    public KnowledgeCategory createCategory(KnowledgeCategoryDTO dto) {
        if (categoryRepository.existsByName(dto.getName())) {
            throw new RuntimeException("分类名称已存在: " + dto.getName());
        }

        KnowledgeCategory category = new KnowledgeCategory();
        category.setName(dto.getName());
        category.setDescription(dto.getDescription());
        category.setSortOrder(dto.getSortOrder());

        return categoryRepository.save(category);
    }

    public KnowledgeCategory updateCategory(Long id, KnowledgeCategoryDTO dto) {
        KnowledgeCategory category = getCategoryById(id);

        if (!category.getName().equals(dto.getName()) && categoryRepository.existsByName(dto.getName())) {
            throw new RuntimeException("分类名称已存在: " + dto.getName());
        }

        category.setName(dto.getName());
        category.setDescription(dto.getDescription());
        category.setSortOrder(dto.getSortOrder());

        return categoryRepository.save(category);
    }

    public void deleteCategory(Long id) {
        KnowledgeCategory category = getCategoryById(id);
        
        if (!category.getQuestions().isEmpty()) {
            throw new RuntimeException("该分类下存在题目，无法删除");
        }
        
        categoryRepository.delete(category);
    }
}
