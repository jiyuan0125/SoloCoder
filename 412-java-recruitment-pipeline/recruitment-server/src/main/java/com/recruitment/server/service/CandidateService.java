package com.recruitment.server.service;

import com.recruitment.common.dto.ApplicationDTO;
import com.recruitment.common.dto.CandidateDTO;
import com.recruitment.common.enums.ErrorCode;
import com.recruitment.common.request.CreateCandidateRequest;
import com.recruitment.common.request.QueryCandidatesRequest;
import com.recruitment.common.response.ApiResponse;
import com.recruitment.common.response.CandidateListResponse;
import com.recruitment.server.repository.ApplicationRepository;
import com.recruitment.server.repository.CandidateRepository;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Service
public class CandidateService {
    private final CandidateRepository candidateRepository;
    private final ApplicationRepository applicationRepository;

    public CandidateService(CandidateRepository candidateRepository, ApplicationRepository applicationRepository) {
        this.candidateRepository = candidateRepository;
        this.applicationRepository = applicationRepository;
    }

    public ApiResponse<CandidateDTO> createCandidate(CreateCandidateRequest request) {
        if (request.getName() == null || request.getName().isEmpty()) {
            return ApiResponse.error(ErrorCode.VALIDATION_ERROR, "姓名不能为空");
        }
        if (request.getPhone() == null || request.getPhone().isEmpty()) {
            return ApiResponse.error(ErrorCode.VALIDATION_ERROR, "手机号不能为空");
        }

        Optional<CandidateDTO> existingCandidate = candidateRepository.findByPhone(request.getPhone());
        if (existingCandidate.isPresent()) {
            CandidateDTO candidate = existingCandidate.get();
            if (request.getPosition() != null && !request.getPosition().isEmpty()) {
                ApplicationDTO application = createApplication(candidate.getId(), request.getPosition(), candidate.getSourceChannel());
                applicationRepository.save(application);
            }
            return ApiResponse.success(candidate);
        }

        CandidateDTO candidate = new CandidateDTO();
        candidate.setId(UUID.randomUUID().toString());
        candidate.setName(request.getName());
        candidate.setPhone(request.getPhone());
        candidate.setEmail(request.getEmail());
        candidate.setResume(request.getResume());
        candidate.setSourceChannel(request.getSourceChannel());
        candidate.setCreatedAt(LocalDateTime.now());

        candidateRepository.save(candidate);

        if (request.getPosition() != null && !request.getPosition().isEmpty()) {
            ApplicationDTO application = createApplication(candidate.getId(), request.getPosition(), candidate.getSourceChannel());
            applicationRepository.save(application);
        }

        return ApiResponse.success(candidate);
    }

    private ApplicationDTO createApplication(String candidateId, String position, com.recruitment.common.enums.SourceChannel sourceChannel) {
        ApplicationDTO application = new ApplicationDTO();
        application.setId(UUID.randomUUID().toString());
        application.setCandidateId(candidateId);
        application.setPosition(position);
        application.setCreatedAt(LocalDateTime.now());
        application.setUpdatedAt(LocalDateTime.now());
        return application;
    }

    public ApiResponse<CandidateListResponse> getCandidateWithApplications(String candidateId) {
        Optional<CandidateDTO> candidateOpt = candidateRepository.findById(candidateId);
        if (!candidateOpt.isPresent()) {
            return ApiResponse.error(ErrorCode.CANDIDATE_NOT_FOUND);
        }

        List<ApplicationDTO> applications = applicationRepository.findByCandidateId(candidateId);
        return ApiResponse.success(new CandidateListResponse(candidateOpt.get(), applications));
    }

    public ApiResponse<List<CandidateListResponse>> queryCandidates(QueryCandidatesRequest request) {
        List<ApplicationDTO> applications = applicationRepository.findByPosition(
                request.getPosition(), request.getStage(), request.getSourceChannel());

        List<CandidateListResponse> result = new ArrayList<>();
        List<String> processedCandidateIds = new ArrayList<>();

        for (ApplicationDTO app : applications) {
            if (!processedCandidateIds.contains(app.getCandidateId())) {
                Optional<CandidateDTO> candidateOpt = candidateRepository.findById(app.getCandidateId());
                if (candidateOpt.isPresent()) {
                    CandidateDTO candidate = candidateOpt.get();
                    List<ApplicationDTO> candidateApps = applicationRepository.findByCandidateId(candidate.getId());
                    
                    if (request.getSourceChannel() == null || request.getSourceChannel() == candidate.getSourceChannel()) {
                        result.add(new CandidateListResponse(candidate, candidateApps));
                    }
                    processedCandidateIds.add(candidate.getId());
                }
            }
        }

        return ApiResponse.success(result);
    }

    public ApiResponse<CandidateDTO> getCandidateById(String id) {
        Optional<CandidateDTO> candidateOpt = candidateRepository.findById(id);
        return candidateOpt.map(ApiResponse::success).orElseGet(() -> ApiResponse.error(ErrorCode.CANDIDATE_NOT_FOUND));
    }
}
