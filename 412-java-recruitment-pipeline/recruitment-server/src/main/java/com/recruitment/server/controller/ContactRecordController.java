package com.recruitment.server.controller;

import com.recruitment.common.dto.ContactRecordDTO;
import com.recruitment.common.request.AddContactRecordRequest;
import com.recruitment.common.response.ApiResponse;
import com.recruitment.server.service.ContactRecordService;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/contacts")
public class ContactRecordController {
    private final ContactRecordService contactRecordService;

    public ContactRecordController(ContactRecordService contactRecordService) {
        this.contactRecordService = contactRecordService;
    }

    @PostMapping
    public ApiResponse<ContactRecordDTO> addContactRecord(@RequestBody AddContactRecordRequest request) {
        return contactRecordService.addContactRecord(request);
    }

    @GetMapping("/candidate/{candidateId}")
    public ApiResponse<List<ContactRecordDTO>> getContactRecords(@PathVariable String candidateId) {
        return contactRecordService.getContactRecords(candidateId);
    }
}
