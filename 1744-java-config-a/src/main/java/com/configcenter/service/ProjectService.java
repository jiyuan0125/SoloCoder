package com.configcenter.service;

import com.configcenter.dto.ProjectDTO;
import com.configcenter.model.Project;
import com.configcenter.repository.ProjectRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;

@Service
@RequiredArgsConstructor
public class ProjectService {

    private final ProjectRepository projectRepository;

    @Transactional
    public Project createProject(ProjectDTO dto) {
        if (projectRepository.existsByName(dto.getName())) {
            throw new RuntimeException("Project with name '" + dto.getName() + "' already exists");
        }
        Project project = new Project();
        project.setName(dto.getName());
        project.setDescription(dto.getDescription());
        project.setValidationCallbackUrl(dto.getValidationCallbackUrl());
        return projectRepository.save(project);
    }

    public Project getProject(Long id) {
        return projectRepository.findById(id)
                .orElseThrow(() -> new RuntimeException("Project not found with id: " + id));
    }

    public Project getProjectByName(String name) {
        return projectRepository.findByName(name)
                .orElseThrow(() -> new RuntimeException("Project not found with name: " + name));
    }

    public List<Project> getAllProjects() {
        return projectRepository.findAll();
    }

    @Transactional
    public Project updateProject(Long id, ProjectDTO dto) {
        Project project = getProject(id);
        if (!project.getName().equals(dto.getName()) && projectRepository.existsByName(dto.getName())) {
            throw new RuntimeException("Project with name '" + dto.getName() + "' already exists");
        }
        project.setName(dto.getName());
        project.setDescription(dto.getDescription());
        project.setValidationCallbackUrl(dto.getValidationCallbackUrl());
        return projectRepository.save(project);
    }

    @Transactional
    public void deleteProject(Long id) {
        projectRepository.deleteById(id);
    }
}
