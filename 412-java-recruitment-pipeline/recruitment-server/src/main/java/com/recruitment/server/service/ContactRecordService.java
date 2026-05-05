package com.recruitment.server.service;

import com.recruitment.common.dto.ContactRecordDTO;
import com.recruitment.common.enums.ErrorCode;
import com.recruitment.common.request.AddContactRecordRequest;
import com.recruitment.common.response.ApiResponse;
import com.recruitment.server.repository.CandidateRepository;
import com.recruitment.server.repository.ContactRecordRepository;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Service
public class ContactRecordService {
    private final ContactRecordRepository contactRecordRepository;
    private final CandidateRepository candidateRepository;

    public ContactRecordService(ContactRecordRepository contactRecordRepository, CandidateRepository candidateRepository) {
        this.contactRecordRepository = contactRecordRepository;
        this.candidateRepository = candidateRepository;
    }

    public ApiResponse<ContactRecordDTO> addContactRecord(AddContactRecordRequest request) {
        if (request.getCandidateId() == null || request.getCandidateId().isEmpty()) {
            return ApiResponse.error(ErrorCode.VALIDATION_ERROR, "候选人ID不能为空");
        }
        if (request.getContent() == null || request.getContent().isEmpty()) {
            return ApiResponse.error(ErrorCode.VALIDATION_ERROR, "联系内容不能为空");
        }

        if (!candidateRepository.existsById(request.getCandidateId())) {
            return ApiResponse.error(ErrorCode.CANDIDATE_NOT_FOUND);
        }

        ContactRecordDTO record = new ContactRecordDTO();
        record.setId(UUID.randomUUID().toString());
        record.setCandidateId(request.getCandidateId());
        record.setContent(request.getContent());
        record.setContactType(request.getContactType());
        record.setContactTime(LocalDateTime.now());
        record.setOperator(request.getOperator());

        contactRecordRepository.save(record);
        return ApiResponse.success(record);
    }

    public ApiResponse<List<ContactRecordDTO>> getContactRecords(String candidateId) {
        if (!candidateRepository.existsById(candidateId)) {
            return ApiResponse.error(ErrorCode.CANDIDATE_NOT_FOUND);
        }

        List<ContactRecordDTO> records = contactRecordRepository.findByCandidateId(candidateId);
        return ApiResponse.success(records);
    }
}
