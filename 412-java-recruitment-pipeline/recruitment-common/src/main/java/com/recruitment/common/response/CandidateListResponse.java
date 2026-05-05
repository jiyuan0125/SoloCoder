package com.recruitment.common.response;

import com.recruitment.common.dto.ApplicationDTO;
import com.recruitment.common.dto.CandidateDTO;

import java.util.List;

public class CandidateListResponse {
    private CandidateDTO candidate;
    private List<ApplicationDTO> applications;

    public CandidateListResponse() {
    }

    public CandidateListResponse(CandidateDTO candidate, List<ApplicationDTO> applications) {
        this.candidate = candidate;
        this.applications = applications;
    }

    public CandidateDTO getCandidate() {
        return candidate;
    }

    public void setCandidate(CandidateDTO candidate) {
        this.candidate = candidate;
    }

    public List<ApplicationDTO> getApplications() {
        return applications;
    }

    public void setApplications(List<ApplicationDTO> applications) {
        this.applications = applications;
    }
}
