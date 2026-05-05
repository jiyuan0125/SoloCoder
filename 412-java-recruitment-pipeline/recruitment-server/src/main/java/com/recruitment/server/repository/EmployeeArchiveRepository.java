package com.recruitment.server.repository;

import com.recruitment.common.dto.EmployeeArchiveDTO;
import org.springframework.stereotype.Repository;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;

@Repository
public class EmployeeArchiveRepository {
    private final Map<String, EmployeeArchiveDTO> archives = new ConcurrentHashMap<>();

    public EmployeeArchiveDTO save(EmployeeArchiveDTO archive) {
        archives.put(archive.getId(), archive);
        return archive;
    }

    public Optional<EmployeeArchiveDTO> findById(String id) {
        return Optional.ofNullable(archives.get(id));
    }

    public List<EmployeeArchiveDTO> findAll() {
        return new ArrayList<>(archives.values());
    }

    public Optional<EmployeeArchiveDTO> findByCandidateId(String candidateId) {
        return archives.values().stream()
                .filter(a -> candidateId.equals(a.getCandidateId()))
                .findFirst();
    }

    public int count() {
        return archives.size();
    }
}
