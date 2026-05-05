package com.recruitment.server.controller;

import com.recruitment.common.dto.ApplicationDTO;
import com.recruitment.common.request.*;
import com.recruitment.common.response.ApiResponse;
import com.recruitment.common.response.ApplicationDetailResponse;
import com.recruitment.server.service.ApplicationService;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/applications")
public class ApplicationController {
    private final ApplicationService applicationService;

    public ApplicationController(ApplicationService applicationService) {
        this.applicationService = applicationService;
    }

    @PostMapping("/add")
    public ApiResponse<ApplicationDTO> addApplication(@RequestParam String candidateId, @RequestParam String position) {
        return applicationService.addApplication(candidateId, position);
    }

    @PostMapping("/stage")
    public ApiResponse<ApplicationDTO> updateStage(@RequestBody UpdateStageRequest request) {
        return applicationService.updateStage(request);
    }

    @PostMapping("/interview")
    public ApiResponse<ApplicationDTO> addInterviewRound(@RequestBody AddInterviewRequest request) {
        return applicationService.addInterviewRound(request);
    }

    @PostMapping("/offer/send")
    public ApiResponse<ApplicationDTO> sendOffer(@RequestBody SendOfferRequest request) {
        return applicationService.sendOffer(request);
    }

    @PostMapping("/offer/respond")
    public ApiResponse<ApplicationDTO> respondOffer(@RequestBody RespondOfferRequest request) {
        return applicationService.respondOffer(request);
    }

    @PostMapping("/onboard")
    public ApiResponse<ApplicationDTO> confirmOnboard(@RequestParam String applicationId) {
        return applicationService.confirmOnboard(applicationId);
    }

    @GetMapping("/{id}")
    public ApiResponse<ApplicationDetailResponse> getApplicationDetail(@PathVariable String id) {
        return applicationService.getApplicationDetail(id);
    }
}
