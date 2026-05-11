package com.school.service;

import com.school.exception.SchedulingException;
import com.school.model.*;

import java.util.*;
import java.util.stream.Collectors;

public class TimetableService {
    private final Map<String, Teacher> teachers = new HashMap<>();
    private final Map<String, Classroom> classrooms = new HashMap<>();
    private final Map<String, Course> courses = new HashMap<>();

    public void addTeacher(Teacher teacher) {
        teachers.put(teacher.getId(), teacher);
    }

    public void addClassroom(Classroom classroom) {
        classrooms.put(classroom.getId(), classroom);
    }

    public Teacher getTeacher(String id) {
        return teachers.get(id);
    }

    public Classroom getClassroom(String id) {
        return classrooms.get(id);
    }

    public List<Teacher> getAllTeachers() {
        return new ArrayList<>(teachers.values());
    }

    public List<Classroom> getAllClassrooms() {
        return new ArrayList<>(classrooms.values());
    }

    public List<Course> getAllCourses() {
        return new ArrayList<>(courses.values());
    }

    public void addCourse(Course course) {
        validateCourse(course, null);
        courses.put(course.getId(), course);
    }

    public void updateCourse(String courseId, Course updatedCourse) {
        Course existingCourse = courses.get(courseId);
        if (existingCourse == null) {
            throw new IllegalArgumentException("Course not found: " + courseId);
        }
        validateCourse(updatedCourse, courseId);
        courses.put(courseId, updatedCourse);
    }

    public void deleteCourse(String courseId) {
        courses.remove(courseId);
    }

    private void validateCourse(Course course, String excludeCourseId) {
        if (course.getTeacher() == null) {
            throw new IllegalArgumentException("Course must have a teacher");
        }
        if (course.getClassroom() == null) {
            throw new IllegalArgumentException("Course must have a classroom");
        }
        if (course.getTimeSlots() == null || course.getTimeSlots().isEmpty()) {
            throw new IllegalArgumentException("Course must have at least one time slot");
        }

        Classroom classroom = classrooms.get(course.getClassroom().getId());
        if (classroom == null) {
            throw new IllegalArgumentException("Classroom not found: " + course.getClassroom().getId());
        }

        if (course.getEnrolledStudents() > classroom.getCapacity()) {
            throw new SchedulingException(
                    String.format("教室容量不足：选课人数 %d 人，教室容量 %d 人",
                            course.getEnrolledStudents(), classroom.getCapacity()),
                    SchedulingException.ConflictType.CAPACITY_EXCEEDED
            );
        }

        for (TimeSlot timeSlot : course.getTimeSlots()) {
            validateTimeSlotConflict(course, timeSlot, excludeCourseId);
        }
    }

    private void validateTimeSlotConflict(Course course, TimeSlot timeSlot, String excludeCourseId) {
        for (Course existingCourse : courses.values()) {
            if (excludeCourseId != null && existingCourse.getId().equals(excludeCourseId)) {
                continue;
            }

            boolean teacherConflict = existingCourse.getTeacher().getId().equals(course.getTeacher().getId());
            boolean classroomConflict = existingCourse.getClassroom().getId().equals(course.getClassroom().getId());

            if (!teacherConflict && !classroomConflict) {
                continue;
            }

            for (TimeSlot existingTimeSlot : existingCourse.getTimeSlots()) {
                if (timeSlot.overlapsWith(existingTimeSlot)) {
                    if (teacherConflict) {
                        throw new SchedulingException(
                                String.format("老师时间冲突：老师 %s 在时段 %s 已有课程 %s",
                                        course.getTeacher().getName(),
                                        timeSlot,
                                        existingCourse.getName()),
                                SchedulingException.ConflictType.TEACHER_CONFLICT
                        );
                    }
                    if (classroomConflict) {
                        throw new SchedulingException(
                                String.format("教室时间冲突：教室 %s 在时段 %s 已有课程 %s",
                                        course.getClassroom().getRoomNumber(),
                                        timeSlot,
                                        existingCourse.getName()),
                                SchedulingException.ConflictType.CLASSROOM_CONFLICT
                        );
                    }
                }
            }
        }
    }

    public List<Course> getCoursesByTeacher(String teacherId) {
        return courses.values().stream()
                .filter(course -> course.getTeacher().getId().equals(teacherId))
                .collect(Collectors.toList());
    }

    public List<Course> getCoursesByClassroom(String classroomId) {
        return courses.values().stream()
                .filter(course -> course.getClassroom().getId().equals(classroomId))
                .collect(Collectors.toList());
    }

    public Map<TimeSlot.DayOfWeek, List<CourseWithTimeSlot>> getWeeklyTimetable() {
        Map<TimeSlot.DayOfWeek, List<CourseWithTimeSlot>> timetable = new TreeMap<>(
                Comparator.comparingInt(TimeSlot.DayOfWeek::getOrder)
        );

        for (TimeSlot.DayOfWeek day : TimeSlot.DayOfWeek.values()) {
            timetable.put(day, new ArrayList<>());
        }

        for (Course course : courses.values()) {
            for (TimeSlot timeSlot : course.getTimeSlots()) {
                timetable.get(timeSlot.getDayOfWeek()).add(new CourseWithTimeSlot(course, timeSlot));
            }
        }

        for (List<CourseWithTimeSlot> dayCourses : timetable.values()) {
            dayCourses.sort(Comparator.comparing(cwt -> cwt.getTimeSlot().getStartTime()));
        }

        return timetable;
    }

    public static class CourseWithTimeSlot {
        private final Course course;
        private final TimeSlot timeSlot;

        public CourseWithTimeSlot(Course course, TimeSlot timeSlot) {
            this.course = course;
            this.timeSlot = timeSlot;
        }

        public Course getCourse() {
            return course;
        }

        public TimeSlot getTimeSlot() {
            return timeSlot;
        }
    }
}
