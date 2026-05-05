package com.training.server.repository;

import com.training.common.enums.RegistrationStatus;
import com.training.server.entity.Registration;
import org.springframework.stereotype.Repository;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;

@Repository
public class RegistrationRepository {

    private final Map<String, Registration> registrations = new HashMap<>();

    public Registration save(Registration registration) {
        registrations.put(registration.getId(), registration);
        return registration;
    }

    public Optional<Registration> findById(String id) {
        return Optional.ofNullable(registrations.get(id));
    }

    public List<Registration> findAll() {
        return new ArrayList<>(registrations.values());
    }

    public void deleteById(String id) {
        registrations.remove(id);
    }

    public boolean existsById(String id) {
        return registrations.containsKey(id);
    }

    public List<Registration> findByCourseId(String courseId) {
        List<Registration> result = new ArrayList<>();
        for (Registration registration : registrations.values()) {
            if (courseId.equals(registration.getCourseId())) {
                result.add(registration);
            }
        }
        return result;
    }

    public List<Registration> findByEmployeeId(String employeeId) {
        List<Registration> result = new ArrayList<>();
        for (Registration registration : registrations.values()) {
            if (employeeId.equals(registration.getEmployeeId())) {
                result.add(registration);
            }
        }
        return result;
    }

    public Optional<Registration> findByEmployeeIdAndCourseId(String employeeId, String courseId) {
        for (Registration registration : registrations.values()) {
            if (employeeId.equals(registration.getEmployeeId()) && courseId.equals(registration.getCourseId())) {
                return Optional.of(registration);
            }
        }
        return Optional.empty();
    }

    public List<Registration> findByCourseIdAndStatus(String courseId, RegistrationStatus status) {
        List<Registration> result = new ArrayList<>();
        for (Registration registration : registrations.values()) {
            if (courseId.equals(registration.getCourseId()) && status == registration.getStatus()) {
                result.add(registration);
            }
        }
        return result;
    }

    public List<Registration> findByStatus(RegistrationStatus status) {
        List<Registration> result = new ArrayList<>();
        for (Registration registration : registrations.values()) {
            if (status == registration.getStatus()) {
                result.add(registration);
            }
        }
        return result;
    }

    public int countByCourseId(String courseId) {
        int count = 0;
        for (Registration registration : registrations.values()) {
            if (courseId.equals(registration.getCourseId()) && 
                registration.getStatus() != RegistrationStatus.CANCELLED) {
                count++;
            }
        }
        return count;
    }
}
