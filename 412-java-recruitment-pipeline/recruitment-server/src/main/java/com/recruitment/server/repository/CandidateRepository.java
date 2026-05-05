package com.recruitment.server.repository;

import com.recruitment.common.dto.CandidateDTO;
import org.springframework.stereotype.Repository;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;

@Repository
public class CandidateRepository {
    private final Map<String, CandidateDTO> candidates = new ConcurrentHashMap<>();

    public CandidateDTO save(CandidateDTO candidate) {
        candidates.put(candidate.getId(), candidate);
        return candidate;
    }

    public Optional<CandidateDTO> findById(String id) {
        return Optional.ofNullable(candidates.get(id));
    }

    public List<CandidateDTO> findAll() {
        return new ArrayList<>(candidates.values());
    }

    public Optional<CandidateDTO> findByPhone(String phone) {
        return candidates.values().stream()
                .filter(c -> phone.equals(c.getPhone()))
                .findFirst();
    }

    public boolean existsById(String id) {
        return candidates.containsKey(id);
    }
}
