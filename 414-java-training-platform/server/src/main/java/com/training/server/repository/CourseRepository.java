package com.training.server.repository;

import com.training.server.entity.Course;
import org.springframework.stereotype.Repository;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;

@Repository
public class CourseRepository {

    private final Map<String, Course> courses = new HashMap<>();

    public Course save(Course course) {
        courses.put(course.getId(), course);
        return course;
    }

    public Optional<Course> findById(String id) {
        return Optional.ofNullable(courses.get(id));
    }

    public List<Course> findAll() {
        return new ArrayList<>(courses.values());
    }

    public void deleteById(String id) {
        courses.remove(id);
    }

    public boolean existsById(String id) {
        return courses.containsKey(id);
    }

    public List<Course> findByInstructorId(String instructorId) {
        List<Course> result = new ArrayList<>();
        for (Course course : courses.values()) {
            if (instructorId.equals(course.getInstructorId())) {
                result.add(course);
            }
        }
        return result;
    }
}
