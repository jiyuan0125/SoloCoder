package com.school.model;

import java.util.ArrayList;
import java.util.List;
import java.util.Objects;

public class Course {
    private String id;
    private String name;
    private Teacher teacher;
    private Classroom classroom;
    private List<TimeSlot> timeSlots;
    private int enrolledStudents;

    public Course() {
        this.timeSlots = new ArrayList<>();
    }

    public Course(String id, String name, Teacher teacher, Classroom classroom, List<TimeSlot> timeSlots, int enrolledStudents) {
        this.id = id;
        this.name = name;
        this.teacher = teacher;
        this.classroom = classroom;
        this.timeSlots = timeSlots != null ? new ArrayList<>(timeSlots) : new ArrayList<>();
        this.enrolledStudents = enrolledStudents;
    }

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public Teacher getTeacher() {
        return teacher;
    }

    public void setTeacher(Teacher teacher) {
        this.teacher = teacher;
    }

    public Classroom getClassroom() {
        return classroom;
    }

    public void setClassroom(Classroom classroom) {
        this.classroom = classroom;
    }

    public List<TimeSlot> getTimeSlots() {
        return timeSlots;
    }

    public void setTimeSlots(List<TimeSlot> timeSlots) {
        this.timeSlots = timeSlots != null ? new ArrayList<>(timeSlots) : new ArrayList<>();
    }

    public int getEnrolledStudents() {
        return enrolledStudents;
    }

    public void setEnrolledStudents(int enrolledStudents) {
        this.enrolledStudents = enrolledStudents;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        Course course = (Course) o;
        return Objects.equals(id, course.id);
    }

    @Override
    public int hashCode() {
        return Objects.hash(id);
    }

    @Override
    public String toString() {
        return "Course{" +
                "id='" + id + '\'' +
                ", name='" + name + '\'' +
                ", teacher=" + teacher +
                ", classroom=" + classroom +
                ", timeSlots=" + timeSlots +
                ", enrolledStudents=" + enrolledStudents +
                '}';
    }
}
