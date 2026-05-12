package com.configcenter.controller;

import com.configcenter.dto.ApiResponse;
import com.configcenter.dto.ProjectDTO;
import com.configcenter.model.Project;
import com.configcenter.service.ProjectService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/projects")
@RequiredArgsConstructor
public class ProjectController {

    private final ProjectService projectService;

    @PostMapping
    public ApiResponse<Project> createProject(@Valid @RequestBody ProjectDTO dto) {
        Project project = projectService.createProject(dto);
        return ApiResponse.success("Project created successfully", project);
    }

    @GetMapping("/{id}")
    public ApiResponse<Project> getProject(@PathVariable Long id) {
        Project project = projectService.getProject(id);
        return ApiResponse.success(project);
    }

    @GetMapping("/name/{name}")
    public ApiResponse<Project> getProjectByName(@PathVariable String name) {
        Project project = projectService.getProjectByName(name);
        return ApiResponse.success(project);
    }

    @GetMapping
    public ApiResponse<List<Project>> getAllProjects() {
        List<Project> projects = projectService.getAllProjects();
        return ApiResponse.success(projects);
    }

    @PutMapping("/{id}")
    public ApiResponse<Project> updateProject(@PathVariable Long id, @Valid @RequestBody ProjectDTO dto) {
        Project project = projectService.updateProject(id, dto);
        return ApiResponse.success("Project updated successfully", project);
    }

    @DeleteMapping("/{id}")
    public ApiResponse<Void> deleteProject(@PathVariable Long id) {
        projectService.deleteProject(id);
        return ApiResponse.success("Project deleted successfully", null);
    }
}
