package com.school.service;

import com.school.exception.SchedulingException;
import com.school.model.*;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.time.LocalTime;
import java.util.Arrays;
import java.util.List;

import static org.junit.jupiter.api.Assertions.*;

class TimetableServiceTest {
    private TimetableService service;
    private Teacher teacher1;
    private Teacher teacher2;
    private Classroom classroom1;
    private Classroom classroom2;

    @BeforeEach
    void setUp() {
        service = new TimetableService();
        teacher1 = new Teacher("T1", "张老师", "数学");
        teacher2 = new Teacher("T2", "李老师", "英语");
        classroom1 = new Classroom("C1", "101", 40);
        classroom2 = new Classroom("C2", "102", 50);

        service.addTeacher(teacher1);
        service.addTeacher(teacher2);
        service.addClassroom(classroom1);
        service.addClassroom(classroom2);
    }

    @Test
    void testAddCourse_Success() {
        List<TimeSlot> slots = Arrays.asList(
                new TimeSlot(TimeSlot.DayOfWeek.MONDAY, LocalTime.of(8, 0), LocalTime.of(9, 40))
        );
        Course course = new Course("C001", "高等数学", teacher1, classroom1, slots, 30);
        service.addCourse(course);
        assertEquals(1, service.getAllCourses().size());
    }

    @Test
    void testAddCourse_TeacherConflict() {
        List<TimeSlot> slots1 = Arrays.asList(
                new TimeSlot(TimeSlot.DayOfWeek.MONDAY, LocalTime.of(8, 0), LocalTime.of(9, 40))
        );
        Course course1 = new Course("C001", "高等数学", teacher1, classroom1, slots1, 30);
        service.addCourse(course1);

        List<TimeSlot> slots2 = Arrays.asList(
                new TimeSlot(TimeSlot.DayOfWeek.MONDAY, LocalTime.of(9, 30), LocalTime.of(11, 10))
        );
        Course course2 = new Course("C002", "线性代数", teacher1, classroom2, slots2, 35);

        SchedulingException exception = assertThrows(SchedulingException.class, () -> service.addCourse(course2));
        assertEquals(SchedulingException.ConflictType.TEACHER_CONFLICT, exception.getConflictType());
    }

    @Test
    void testAddCourse_ClassroomConflict() {
        List<TimeSlot> slots1 = Arrays.asList(
                new TimeSlot(TimeSlot.DayOfWeek.MONDAY, LocalTime.of(8, 0), LocalTime.of(9, 40))
        );
        Course course1 = new Course("C001", "高等数学", teacher1, classroom1, slots1, 30);
        service.addCourse(course1);

        List<TimeSlot> slots2 = Arrays.asList(
                new TimeSlot(TimeSlot.DayOfWeek.MONDAY, LocalTime.of(8, 30), LocalTime.of(10, 10))
        );
        Course course2 = new Course("C002", "大学英语", teacher2, classroom1, slots2, 40);

        SchedulingException exception = assertThrows(SchedulingException.class, () -> service.addCourse(course2));
        assertEquals(SchedulingException.ConflictType.CLASSROOM_CONFLICT, exception.getConflictType());
    }

    @Test
    void testAddCourse_CapacityExceeded() {
        List<TimeSlot> slots = Arrays.asList(
                new TimeSlot(TimeSlot.DayOfWeek.MONDAY, LocalTime.of(8, 0), LocalTime.of(9, 40))
        );
        Course course = new Course("C001", "高等数学", teacher1, classroom1, slots, 50);

        SchedulingException exception = assertThrows(SchedulingException.class, () -> service.addCourse(course));
        assertEquals(SchedulingException.ConflictType.CAPACITY_EXCEEDED, exception.getConflictType());
    }

    @Test
    void testAddCourse_MultipleTimeSlots_OneConflict() {
        List<TimeSlot> slots1 = Arrays.asList(
                new TimeSlot(TimeSlot.DayOfWeek.MONDAY, LocalTime.of(8, 0), LocalTime.of(9, 40)),
                new TimeSlot(TimeSlot.DayOfWeek.WEDNESDAY, LocalTime.of(8, 0), LocalTime.of(9, 40))
        );
        Course course1 = new Course("C001", "高等数学", teacher1, classroom1, slots1, 30);
        service.addCourse(course1);

        List<TimeSlot> slots2 = Arrays.asList(
                new TimeSlot(TimeSlot.DayOfWeek.TUESDAY, LocalTime.of(8, 0), LocalTime.of(9, 40)),
                new TimeSlot(TimeSlot.DayOfWeek.WEDNESDAY, LocalTime.of(9, 30), LocalTime.of(11, 10))
        );
        Course course2 = new Course("C002", "线性代数", teacher1, classroom2, slots2, 35);

        SchedulingException exception = assertThrows(SchedulingException.class, () -> service.addCourse(course2));
        assertEquals(SchedulingException.ConflictType.TEACHER_CONFLICT, exception.getConflictType());
    }

    @Test
    void testUpdateCourse_Success() {
        List<TimeSlot> slots1 = Arrays.asList(
                new TimeSlot(TimeSlot.DayOfWeek.MONDAY, LocalTime.of(8, 0), LocalTime.of(9, 40))
        );
        Course course = new Course("C001", "高等数学", teacher1, classroom1, slots1, 30);
        service.addCourse(course);

        List<TimeSlot> newSlots = Arrays.asList(
                new TimeSlot(TimeSlot.DayOfWeek.MONDAY, LocalTime.of(10, 0), LocalTime.of(11, 40))
        );
        Course updatedCourse = new Course("C001", "高等数学", teacher1, classroom1, newSlots, 30);
        service.updateCourse("C001", updatedCourse);

        Course result = service.getAllCourses().get(0);
        assertEquals(LocalTime.of(10, 0), result.getTimeSlots().get(0).getStartTime());
    }

    @Test
    void testUpdateCourse_WithConflict() {
        List<TimeSlot> slots1 = Arrays.asList(
                new TimeSlot(TimeSlot.DayOfWeek.MONDAY, LocalTime.of(8, 0), LocalTime.of(9, 40))
        );
        Course course1 = new Course("C001", "高等数学", teacher1, classroom1, slots1, 30);
        service.addCourse(course1);

        List<TimeSlot> slots2 = Arrays.asList(
                new TimeSlot(TimeSlot.DayOfWeek.TUESDAY, LocalTime.of(8, 0), LocalTime.of(9, 40))
        );
        Course course2 = new Course("C002", "大学英语", teacher1, classroom2, slots2, 35);
        service.addCourse(course2);

        List<TimeSlot> newSlots = Arrays.asList(
                new TimeSlot(TimeSlot.DayOfWeek.TUESDAY, LocalTime.of(8, 30), LocalTime.of(10, 10))
        );
        Course updatedCourse = new Course("C001", "高等数学", teacher1, classroom1, newSlots, 30);

        SchedulingException exception = assertThrows(SchedulingException.class, () -> service.updateCourse("C001", updatedCourse));
        assertEquals(SchedulingException.ConflictType.TEACHER_CONFLICT, exception.getConflictType());
    }

    @Test
    void testGetCoursesByTeacher() {
        List<TimeSlot> slots1 = Arrays.asList(
                new TimeSlot(TimeSlot.DayOfWeek.MONDAY, LocalTime.of(8, 0), LocalTime.of(9, 40))
        );
        Course course1 = new Course("C001", "高等数学", teacher1, classroom1, slots1, 30);
        service.addCourse(course1);

        List<TimeSlot> slots2 = Arrays.asList(
                new TimeSlot(TimeSlot.DayOfWeek.MONDAY, LocalTime.of(10, 0), LocalTime.of(11, 40))
        );
        Course course2 = new Course("C002", "大学英语", teacher2, classroom2, slots2, 40);
        service.addCourse(course2);

        List<Course> teacher1Courses = service.getCoursesByTeacher("T1");
        assertEquals(1, teacher1Courses.size());
        assertEquals("高等数学", teacher1Courses.get(0).getName());

        List<Course> teacher2Courses = service.getCoursesByTeacher("T2");
        assertEquals(1, teacher2Courses.size());
        assertEquals("大学英语", teacher2Courses.get(0).getName());
    }

    @Test
    void testGetCoursesByClassroom() {
        List<TimeSlot> slots1 = Arrays.asList(
                new TimeSlot(TimeSlot.DayOfWeek.MONDAY, LocalTime.of(8, 0), LocalTime.of(9, 40))
        );
        Course course1 = new Course("C001", "高等数学", teacher1, classroom1, slots1, 30);
        service.addCourse(course1);

        List<TimeSlot> slots2 = Arrays.asList(
                new TimeSlot(TimeSlot.DayOfWeek.MONDAY, LocalTime.of(10, 0), LocalTime.of(11, 40))
        );
        Course course2 = new Course("C002", "大学英语", teacher2, classroom2, slots2, 40);
        service.addCourse(course2);

        List<Course> c1Courses = service.getCoursesByClassroom("C1");
        assertEquals(1, c1Courses.size());
        assertEquals("高等数学", c1Courses.get(0).getName());
    }
}
