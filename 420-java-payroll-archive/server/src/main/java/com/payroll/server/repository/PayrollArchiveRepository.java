package com.payroll.server.repository;

import com.payroll.server.entity.PayrollArchive;
import org.springframework.stereotype.Repository;

import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.Set;
import java.util.TreeSet;
import java.util.concurrent.ConcurrentHashMap;

@Repository
public class PayrollArchiveRepository {

    private final Map<Integer, Map<String, PayrollArchive>> yearArchives = new ConcurrentHashMap<>();
    private final Map<String, Integer> archiveIdToYear = new ConcurrentHashMap<>();

    public PayrollArchive save(PayrollArchive archive) {
        int year = archive.getYear();
        String archiveId = archive.getArchiveId();

        yearArchives.computeIfAbsent(year, k -> new ConcurrentHashMap<>())
                .put(archiveId, archive);
        archiveIdToYear.put(archiveId, year);

        return archive;
    }

    public Optional<PayrollArchive> findByArchiveId(String archiveId) {
        Integer year = archiveIdToYear.get(archiveId);
        if (year == null) {
            return Optional.empty();
        }
        Map<String, PayrollArchive> archives = yearArchives.get(year);
        if (archives == null) {
            return Optional.empty();
        }
        return Optional.ofNullable(archives.get(archiveId));
    }

    public Optional<PayrollArchive> findByEmployeeIdAndYearMonth(String employeeId, int year, int month) {
        Map<String, PayrollArchive> archives = yearArchives.get(year);
        if (archives == null) {
            return Optional.empty();
        }
        return archives.values().stream()
                .filter(a -> a.getEmployeeId().equals(employeeId)
                        && a.getYear() == year
                        && a.getMonth() == month)
                .findFirst();
    }

    public List<PayrollArchive> findByYear(int year) {
        Map<String, PayrollArchive> archives = yearArchives.get(year);
        if (archives == null) {
            return Collections.emptyList();
        }
        return new ArrayList<>(archives.values());
    }

    public List<PayrollArchive> findByYearAndMonth(int year, int month) {
        Map<String, PayrollArchive> archives = yearArchives.get(year);
        if (archives == null) {
            return Collections.emptyList();
        }
        List<PayrollArchive> result = new ArrayList<>();
        for (PayrollArchive archive : archives.values()) {
            if (archive.getMonth() == month) {
                result.add(archive);
            }
        }
        return result;
    }

    public List<PayrollArchive> findByEmployeeId(String employeeId) {
        List<PayrollArchive> result = new ArrayList<>();
        for (Map<String, PayrollArchive> archives : yearArchives.values()) {
            for (PayrollArchive archive : archives.values()) {
                if (archive.getEmployeeId().equals(employeeId)) {
                    result.add(archive);
                }
            }
        }
        return result;
    }

    public List<PayrollArchive> findByEmployeeIdAndYear(String employeeId, int year) {
        Map<String, PayrollArchive> archives = yearArchives.get(year);
        if (archives == null) {
            return Collections.emptyList();
        }
        List<PayrollArchive> result = new ArrayList<>();
        for (PayrollArchive archive : archives.values()) {
            if (archive.getEmployeeId().equals(employeeId)) {
                result.add(archive);
            }
        }
        return result;
    }

    public List<PayrollArchive> findByYears(List<Integer> years) {
        List<PayrollArchive> result = new ArrayList<>();
        for (int year : years) {
            Map<String, PayrollArchive> archives = yearArchives.get(year);
            if (archives != null) {
                result.addAll(archives.values());
            }
        }
        return result;
    }

    public Set<Integer> getAvailableYears() {
        return new TreeSet<>(yearArchives.keySet());
    }

    public boolean existsByEmployeeIdAndYearMonth(String employeeId, int year, int month) {
        return findByEmployeeIdAndYearMonth(employeeId, year, month).isPresent();
    }
}
