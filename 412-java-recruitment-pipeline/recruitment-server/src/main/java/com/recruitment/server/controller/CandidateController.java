package com.recruitment.server.controller;

import com.recruitment.common.dto.CandidateDTO;
import com.recruitment.common.request.CreateCandidateRequest;
import com.recruitment.common.request.QueryCandidatesRequest;
import com.recruitment.common.response.ApiResponse;
import com.recruitment.common.response.CandidateListResponse;
import com.recruitment.server.service.CandidateService;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/candidates")
public class CandidateController {
    private final CandidateService candidateService;

    public CandidateController(CandidateService candidateService) {
        this.candidateService = candidateService;
    }

    @PostMapping
    public ApiResponse<CandidateDTO> createCandidate(@RequestBody CreateCandidateRequest request) {
        return candidateService.createCandidate(request);
    }

    @GetMapping("/{id}")
    public ApiResponse<CandidateListResponse> getCandidateWithApplications(@PathVariable String id) {
        return candidateService.getCandidateWithApplications(id);
    }

    @PostMapping("/query")
    public ApiResponse<List<CandidateListResponse>> queryCandidates(@RequestBody QueryCandidatesRequest request) {
        return candidateService.queryCandidates(request);
    }

    @GetMapping("/simple/{id}")
    public ApiResponse<CandidateDTO> getCandidateById(@PathVariable String id) {
        return candidateService.getCandidateById(id);
    }
}
