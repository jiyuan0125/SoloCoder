package com.configcenter.controller;

import com.configcenter.dto.ApiResponse;
import com.configcenter.dto.ReleaseRequestDTO;
import com.configcenter.model.Release;
import com.configcenter.service.ReleaseService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/projects/{projectId}/releases")
@RequiredArgsConstructor
public class ReleaseController {

    private final ReleaseService releaseService;

    @PostMapping
    public ApiResponse<Release> createRelease(
            @PathVariable Long projectId,
            @Valid @RequestBody ReleaseRequestDTO request) {
        Release release = releaseService.createRelease(projectId, request);
        return ApiResponse.success("Release created successfully", release);
    }

    @PostMapping("/{releaseId}/execute")
    public ApiResponse<Release> executeRelease(
            @PathVariable Long projectId,
            @PathVariable Long releaseId) {
        Release release = releaseService.executeRelease(releaseId);
        return ApiResponse.success("Release executed successfully", release);
    }

    @PostMapping("/{releaseId}/full-release")
    public ApiResponse<Release> fullReleaseGrayRelease(
            @PathVariable Long projectId,
            @PathVariable Long releaseId) {
        Release release = releaseService.fullReleaseGrayRelease(releaseId);
        return ApiResponse.success("Gray release promoted to full release", release);
    }

    @PostMapping("/{releaseId}/rollback")
    public ApiResponse<Release> rollbackRelease(
            @PathVariable Long projectId,
            @PathVariable Long releaseId) {
        Release release = releaseService.rollbackRelease(releaseId);
        return ApiResponse.success("Release rolled back successfully", release);
    }

    @GetMapping("/{releaseId}")
    public ApiResponse<Release> getRelease(
            @PathVariable Long projectId,
            @PathVariable Long releaseId) {
        Release release = releaseService.getRelease(releaseId);
        return ApiResponse.success(release);
    }

    @GetMapping
    public ApiResponse<List<Release>> getReleases(
            @PathVariable Long projectId,
            @RequestParam String environment) {
        List<Release> releases = releaseService.getReleases(projectId, environment);
        return ApiResponse.success(releases);
    }

    @GetMapping("/gray/active")
    public ApiResponse<List<Release>> getActiveGrayReleases(
            @PathVariable Long projectId,
            @RequestParam String environment) {
        List<Release> releases = releaseService.getActiveGrayReleases(projectId, environment);
        return ApiResponse.success(releases);
    }
}
