package com.recruitment.common.response;

import com.recruitment.common.dto.ApplicationDTO;
import com.recruitment.common.dto.CandidateDTO;
import com.recruitment.common.dto.ContactRecordDTO;

import java.util.List;

public class ApplicationDetailResponse {
    private CandidateDTO candidate;
    private ApplicationDTO application;
    private List<ContactRecordDTO> contactRecords;

    public ApplicationDetailResponse() {
    }

    public ApplicationDetailResponse(CandidateDTO candidate, ApplicationDTO application, List<ContactRecordDTO> contactRecords) {
        this.candidate = candidate;
        this.application = application;
        this.contactRecords = contactRecords;
    }

    public CandidateDTO getCandidate() {
        return candidate;
    }

    public void setCandidate(CandidateDTO candidate) {
        this.candidate = candidate;
    }

    public ApplicationDTO getApplication() {
        return application;
    }

    public void setApplication(ApplicationDTO application) {
        this.application = application;
    }

    public List<ContactRecordDTO> getContactRecords() {
        return contactRecords;
    }

    public void setContactRecords(List<ContactRecordDTO> contactRecords) {
        this.contactRecords = contactRecords;
    }
}
