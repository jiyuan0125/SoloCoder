package com.recruitment.server.repository;

import com.recruitment.common.dto.ContactRecordDTO;
import org.springframework.stereotype.Repository;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;

@Repository
public class ContactRecordRepository {
    private final Map<String, ContactRecordDTO> contactRecords = new ConcurrentHashMap<>();

    public ContactRecordDTO save(ContactRecordDTO record) {
        contactRecords.put(record.getId(), record);
        return record;
    }

    public List<ContactRecordDTO> findByCandidateId(String candidateId) {
        return contactRecords.values().stream()
                .filter(r -> candidateId.equals(r.getCandidateId()))
                .sorted(Comparator.comparing(ContactRecordDTO::getContactTime).reversed())
                .collect(Collectors.toList());
    }

    public Optional<ContactRecordDTO> findById(String id) {
        return Optional.ofNullable(contactRecords.get(id));
    }
}
