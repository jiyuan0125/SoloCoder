package com.orgchart.server.repository;

import com.orgchart.server.entity.VirtualTeam;
import org.springframework.stereotype.Repository;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;

@Repository
public class VirtualTeamRepository {

    private final Map<String, VirtualTeam> virtualTeams = new ConcurrentHashMap<>();

    public VirtualTeam save(VirtualTeam virtualTeam) {
        virtualTeams.put(virtualTeam.getId(), virtualTeam);
        return virtualTeam;
    }

    public Optional<VirtualTeam> findById(String id) {
        return Optional.ofNullable(virtualTeams.get(id));
    }

    public List<VirtualTeam> findAll() {
        return new ArrayList<>(virtualTeams.values());
    }

    public void deleteById(String id) {
        virtualTeams.remove(id);
    }

    public boolean existsById(String id) {
        return virtualTeams.containsKey(id);
    }

    public List<VirtualTeam> findByMemberId(String memberId) {
        return virtualTeams.values().stream()
                .filter(vt -> vt.getMemberIds().contains(memberId))
                .collect(Collectors.toList());
    }

    public boolean existsByName(String name) {
        if (name == null) {
            return false;
        }
        return virtualTeams.values().stream()
                .anyMatch(vt -> name.equals(vt.getName()));
    }

    public boolean existsByNameExcludingId(String name, String excludeId) {
        if (name == null) {
            return false;
        }
        return virtualTeams.values().stream()
                .anyMatch(vt -> name.equals(vt.getName()) && !excludeId.equals(vt.getId()));
    }
}
