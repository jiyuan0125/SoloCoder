package com.example.server.repository;

import com.example.common.dto.ChecklistTemplateDTO;
import com.example.common.enums.PositionType;
import org.springframework.stereotype.Repository;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;

@Repository
public class TemplateRepository {
    private final Map<String, ChecklistTemplateDTO> templates = new ConcurrentHashMap<>();
    private final Map<String, ChecklistTemplateDTO> templatesByName = new ConcurrentHashMap<>();

    public ChecklistTemplateDTO save(ChecklistTemplateDTO template) {
        templates.put(template.getId(), template);
        templatesByName.put(template.getName(), template);
        return template;
    }

    public Optional<ChecklistTemplateDTO> findById(String id) {
        return Optional.ofNullable(templates.get(id));
    }

    public Optional<ChecklistTemplateDTO> findByName(String name) {
        return Optional.ofNullable(templatesByName.get(name));
    }

    public List<ChecklistTemplateDTO> findAll() {
        return new ArrayList<>(templates.values());
    }

    public void deleteById(String id) {
        ChecklistTemplateDTO template = templates.remove(id);
        if (template != null) {
            templatesByName.remove(template.getName());
        }
    }

    public boolean existsById(String id) {
        return templates.containsKey(id);
    }

    public boolean existsByName(String name) {
        return templatesByName.containsKey(name);
    }

    public List<ChecklistTemplateDTO> findByPositionType(PositionType positionType) {
        return templates.values().stream()
                .filter(t -> t.getPositionType() == positionType)
                .collect(Collectors.toList());
    }

    public Optional<ChecklistTemplateDTO> findDefaultByPositionType(PositionType positionType) {
        return templates.values().stream()
                .filter(t -> t.getPositionType() == positionType && t.isDefault())
                .findFirst();
    }
}
