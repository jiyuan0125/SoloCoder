package com.exam.repository;

import com.exam.entity.KnowledgeCategory;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

@Repository
public interface KnowledgeCategoryRepository extends JpaRepository<KnowledgeCategory, Long> {
    Optional<KnowledgeCategory> findByName(String name);
    
    List<KnowledgeCategory> findAllByOrderBySortOrderAsc();
    
    boolean existsByName(String name);
}
