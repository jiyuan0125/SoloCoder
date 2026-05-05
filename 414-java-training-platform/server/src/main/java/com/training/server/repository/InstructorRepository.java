package com.training.server.repository;

import com.training.server.entity.Instructor;
import org.springframework.stereotype.Repository;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;

@Repository
public class InstructorRepository {

    private final Map<String, Instructor> instructors = new HashMap<>();

    public Instructor save(Instructor instructor) {
        instructors.put(instructor.getId(), instructor);
        return instructor;
    }

    public Optional<Instructor> findById(String id) {
        return Optional.ofNullable(instructors.get(id));
    }

    public List<Instructor> findAll() {
        return new ArrayList<>(instructors.values());
    }

    public void deleteById(String id) {
        instructors.remove(id);
    }

    public boolean existsById(String id) {
        return instructors.containsKey(id);
    }

    public Optional<Instructor> findByEmail(String email) {
        for (Instructor instructor : instructors.values()) {
            if (email.equals(instructor.getEmail())) {
                return Optional.of(instructor);
            }
        }
        return Optional.empty();
    }
}
