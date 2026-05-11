package com.school;

import com.school.exception.SchedulingException;
import com.school.model.*;
import com.school.service.TimetableService;
import com.school.util.TimeSlotParser;

import java.time.LocalTime;
import java.util.*;

public class Main {
    private static final Scanner scanner = new Scanner(System.in);
    private static final TimetableService timetableService = new TimetableService();

    public static void main(String[] args) {
        initializeSampleData();
        while (true) {
            printMenu();
            try {
                int choice = Integer.parseInt(scanner.nextLine().trim());
                handleChoice(choice);
            } catch (NumberFormatException e) {
                System.out.println("请输入有效的数字选项！");
            } catch (Exception e) {
                System.out.println("错误：" + e.getMessage());
            }
        }
    }

    private static void initializeSampleData() {
        Teacher zhang = new Teacher("T001", "张老师", "数学");
        Teacher li = new Teacher("T002", "李老师", "英语");
        Teacher wang = new Teacher("T003", "王老师", "物理");
        Teacher zhao = new Teacher("T004", "赵老师", "化学");

        timetableService.addTeacher(zhang);
        timetableService.addTeacher(li);
        timetableService.addTeacher(wang);
        timetableService.addTeacher(zhao);

        Classroom c101 = new Classroom("C101", "101教室", 40);
        Classroom c102 = new Classroom("C102", "102教室", 50);
        Classroom c201 = new Classroom("C201", "201教室", 30);
        Classroom c202 = new Classroom("C202", "202教室", 60);

        timetableService.addClassroom(c101);
        timetableService.addClassroom(c102);
        timetableService.addClassroom(c201);
        timetableService.addClassroom(c202);

        try {
            List<TimeSlot> mathSlots = Arrays.asList(
                    new TimeSlot(TimeSlot.DayOfWeek.MONDAY, LocalTime.of(8, 0), LocalTime.of(9, 40)),
                    new TimeSlot(TimeSlot.DayOfWeek.WEDNESDAY, LocalTime.of(8, 0), LocalTime.of(9, 40)),
                    new TimeSlot(TimeSlot.DayOfWeek.FRIDAY, LocalTime.of(8, 0), LocalTime.of(9, 40))
            );
            Course math = new Course("C001", "高等数学", zhang, c101, mathSlots, 35);
            timetableService.addCourse(math);

            List<TimeSlot> englishSlots = Arrays.asList(
                    new TimeSlot(TimeSlot.DayOfWeek.MONDAY, LocalTime.of(10, 0), LocalTime.of(11, 40)),
                    new TimeSlot(TimeSlot.DayOfWeek.TUESDAY, LocalTime.of(8, 0), LocalTime.of(9, 40))
            );
            Course english = new Course("C002", "大学英语", li, c102, englishSlots, 45);
            timetableService.addCourse(english);

            List<TimeSlot> physicsSlots = Arrays.asList(
                    new TimeSlot(TimeSlot.DayOfWeek.THURSDAY, LocalTime.of(8, 0), LocalTime.of(9, 40)),
                    new TimeSlot(TimeSlot.DayOfWeek.FRIDAY, LocalTime.of(14, 0), LocalTime.of(15, 40))
            );
            Course physics = new Course("C003", "大学物理", wang, c201, physicsSlots, 28);
            timetableService.addCourse(physics);
        } catch (SchedulingException e) {
            System.out.println("初始化样例数据时出错：" + e.getMessage());
        }
    }

    private static void printMenu() {
        System.out.println("\n========== 课程排课系统 ==========");
        System.out.println("1. 查看所有老师");
        System.out.println("2. 查看所有教室");
        System.out.println("3. 添加课程");
        System.out.println("4. 修改课程（调课）");
        System.out.println("5. 删除课程");
        System.out.println("6. 按老师查询课表");
        System.out.println("7. 按教室查询课表");
        System.out.println("8. 查看全校周课表");
        System.out.println("0. 退出系统");
        System.out.print("请选择操作：");
    }

    private static void handleChoice(int choice) {
        switch (choice) {
            case 1:
                listAllTeachers();
                break;
            case 2:
                listAllClassrooms();
                break;
            case 3:
                addCourse();
                break;
            case 4:
                updateCourse();
                break;
            case 5:
                deleteCourse();
                break;
            case 6:
                queryByTeacher();
                break;
            case 7:
                queryByClassroom();
                break;
            case 8:
                showWeeklyTimetable();
                break;
            case 0:
                System.out.println("感谢使用课程排课系统，再见！");
                System.exit(0);
            default:
                System.out.println("无效的选项，请重新选择！");
        }
    }

    private static void listAllTeachers() {
        System.out.println("\n--- 所有老师 ---");
        List<Teacher> teachers = timetableService.getAllTeachers();
        if (teachers.isEmpty()) {
            System.out.println("暂无老师数据");
            return;
        }
        System.out.println("ID\t\t姓名\t\t所教科目");
        for (Teacher t : teachers) {
            System.out.println(t.getId() + "\t" + t.getName() + "\t" + t.getSubject());
        }
    }

    private static void listAllClassrooms() {
        System.out.println("\n--- 所有教室 ---");
        List<Classroom> classrooms = timetableService.getAllClassrooms();
        if (classrooms.isEmpty()) {
            System.out.println("暂无教室数据");
            return;
        }
        System.out.println("ID\t\t编号\t\t容量");
        for (Classroom c : classrooms) {
            System.out.println(c.getId() + "\t" + c.getRoomNumber() + "\t" + c.getCapacity() + "人");
        }
    }

    private static void addCourse() {
        System.out.println("\n--- 添加课程 ---");
        
        System.out.print("课程编号：");
        String courseId = scanner.nextLine().trim();
        
        System.out.print("课程名称：");
        String courseName = scanner.nextLine().trim();
        
        listAllTeachers();
        System.out.print("选择老师ID：");
        String teacherId = scanner.nextLine().trim();
        Teacher teacher = timetableService.getTeacher(teacherId);
        if (teacher == null) {
            System.out.println("老师不存在！");
            return;
        }
        
        listAllClassrooms();
        System.out.print("选择教室ID：");
        String classroomId = scanner.nextLine().trim();
        Classroom classroom = timetableService.getClassroom(classroomId);
        if (classroom == null) {
            System.out.println("教室不存在！");
            return;
        }
        
        System.out.print("选课人数：");
        int enrolled = Integer.parseInt(scanner.nextLine().trim());
        
        System.out.println("输入上课时段（格式：周一 08:00-09:40），输入空行结束");
        List<TimeSlot> timeSlots = new ArrayList<>();
        while (true) {
            System.out.print("时段 " + (timeSlots.size() + 1) + "：");
            String slotStr = scanner.nextLine().trim();
            if (slotStr.isEmpty()) {
                break;
            }
            try {
                timeSlots.add(TimeSlotParser.parse(slotStr));
            } catch (Exception e) {
                System.out.println("格式错误：" + e.getMessage() + "，请重新输入");
            }
        }
        
        if (timeSlots.isEmpty()) {
            System.out.println("至少需要一个时段！");
            return;
        }
        
        try {
            Course course = new Course(courseId, courseName, teacher, classroom, timeSlots, enrolled);
            timetableService.addCourse(course);
            System.out.println("课程添加成功！");
        } catch (SchedulingException e) {
            System.out.println("排课冲突：" + e.getMessage());
        } catch (Exception e) {
            System.out.println("添加失败：" + e.getMessage());
        }
    }

    private static void updateCourse() {
        System.out.println("\n--- 修改课程（调课）---");
        
        List<Course> allCourses = timetableService.getAllCourses();
        if (allCourses.isEmpty()) {
            System.out.println("暂无课程数据");
            return;
        }
        
        System.out.println("当前课程列表：");
        for (Course c : allCourses) {
            System.out.println(c.getId() + "\t" + c.getName());
        }
        
        System.out.print("输入要修改的课程编号：");
        String courseId = scanner.nextLine().trim();
        
        Course existingCourse = null;
        for (Course c : allCourses) {
            if (c.getId().equals(courseId)) {
                existingCourse = c;
                break;
            }
        }
        
        if (existingCourse == null) {
            System.out.println("课程不存在！");
            return;
        }
        
        System.out.println("当前课程信息：");
        System.out.println("课程名称：" + existingCourse.getName());
        System.out.println("老师：" + existingCourse.getTeacher().getName());
        System.out.println("教室：" + existingCourse.getClassroom().getRoomNumber());
        System.out.println("选课人数：" + existingCourse.getEnrolledStudents());
        System.out.println("上课时段：");
        for (TimeSlot slot : existingCourse.getTimeSlots()) {
            System.out.println("  " + slot);
        }
        
        System.out.println("\n输入新的课程信息（直接回车保留原值）");
        
        System.out.print("课程名称 [" + existingCourse.getName() + "]：");
        String courseName = scanner.nextLine().trim();
        if (courseName.isEmpty()) courseName = existingCourse.getName();
        
        listAllTeachers();
        System.out.print("老师ID [" + existingCourse.getTeacher().getId() + "]：");
        String teacherId = scanner.nextLine().trim();
        Teacher teacher = existingCourse.getTeacher();
        if (!teacherId.isEmpty()) {
            teacher = timetableService.getTeacher(teacherId);
            if (teacher == null) {
                System.out.println("老师不存在！");
                return;
            }
        }
        
        listAllClassrooms();
        System.out.print("教室ID [" + existingCourse.getClassroom().getId() + "]：");
        String classroomId = scanner.nextLine().trim();
        Classroom classroom = existingCourse.getClassroom();
        if (!classroomId.isEmpty()) {
            classroom = timetableService.getClassroom(classroomId);
            if (classroom == null) {
                System.out.println("教室不存在！");
                return;
            }
        }
        
        System.out.print("选课人数 [" + existingCourse.getEnrolledStudents() + "]：");
        String enrolledStr = scanner.nextLine().trim();
        int enrolled = existingCourse.getEnrolledStudents();
        if (!enrolledStr.isEmpty()) {
            enrolled = Integer.parseInt(enrolledStr);
        }
        
        System.out.println("是否修改上课时段？(y/n)");
        String modifySlots = scanner.nextLine().trim().toLowerCase();
        List<TimeSlot> timeSlots = existingCourse.getTimeSlots();
        
        if (modifySlots.equals("y")) {
            System.out.println("输入新的上课时段（格式：周一 08:00-09:40），输入空行结束");
            timeSlots = new ArrayList<>();
            while (true) {
                System.out.print("时段 " + (timeSlots.size() + 1) + "：");
                String slotStr = scanner.nextLine().trim();
                if (slotStr.isEmpty()) {
                    break;
                }
                try {
                    timeSlots.add(TimeSlotParser.parse(slotStr));
                } catch (Exception e) {
                    System.out.println("格式错误：" + e.getMessage() + "，请重新输入");
                }
            }
            
            if (timeSlots.isEmpty()) {
                System.out.println("至少需要一个时段！");
                return;
            }
        }
        
        try {
            Course updatedCourse = new Course(courseId, courseName, teacher, classroom, timeSlots, enrolled);
            timetableService.updateCourse(courseId, updatedCourse);
            System.out.println("课程修改成功！");
        } catch (SchedulingException e) {
            System.out.println("排课冲突：" + e.getMessage());
        } catch (Exception e) {
            System.out.println("修改失败：" + e.getMessage());
        }
    }

    private static void deleteCourse() {
        System.out.println("\n--- 删除课程 ---");
        
        List<Course> allCourses = timetableService.getAllCourses();
        if (allCourses.isEmpty()) {
            System.out.println("暂无课程数据");
            return;
        }
        
        System.out.println("当前课程列表：");
        for (Course c : allCourses) {
            System.out.println(c.getId() + "\t" + c.getName());
        }
        
        System.out.print("输入要删除的课程编号：");
        String courseId = scanner.nextLine().trim();
        
        timetableService.deleteCourse(courseId);
        System.out.println("课程删除成功！");
    }

    private static void queryByTeacher() {
        System.out.println("\n--- 按老师查询课表 ---");
        listAllTeachers();
        System.out.print("选择老师ID：");
        String teacherId = scanner.nextLine().trim();
        
        Teacher teacher = timetableService.getTeacher(teacherId);
        if (teacher == null) {
            System.out.println("老师不存在！");
            return;
        }
        
        List<Course> courses = timetableService.getCoursesByTeacher(teacherId);
        if (courses.isEmpty()) {
            System.out.println("该老师暂无课程安排");
            return;
        }
        
        System.out.println("\n" + teacher.getName() + " 的课表：");
        for (Course course : courses) {
            System.out.println("\n课程：" + course.getName());
            System.out.println("  教室：" + course.getClassroom().getRoomNumber());
            System.out.println("  选课人数：" + course.getEnrolledStudents());
            System.out.println("  上课时段：");
            for (TimeSlot slot : course.getTimeSlots()) {
                System.out.println("    " + slot);
            }
        }
    }

    private static void queryByClassroom() {
        System.out.println("\n--- 按教室查询课表 ---");
        listAllClassrooms();
        System.out.print("选择教室ID：");
        String classroomId = scanner.nextLine().trim();
        
        Classroom classroom = timetableService.getClassroom(classroomId);
        if (classroom == null) {
            System.out.println("教室不存在！");
            return;
        }
        
        List<Course> courses = timetableService.getCoursesByClassroom(classroomId);
        if (courses.isEmpty()) {
            System.out.println("该教室暂无课程安排");
            return;
        }
        
        System.out.println("\n" + classroom.getRoomNumber() + " 的课表：");
        for (Course course : courses) {
            System.out.println("\n课程：" + course.getName());
            System.out.println("  老师：" + course.getTeacher().getName());
            System.out.println("  选课人数：" + course.getEnrolledStudents());
            System.out.println("  上课时段：");
            for (TimeSlot slot : course.getTimeSlots()) {
                System.out.println("    " + slot);
            }
        }
    }

    private static void showWeeklyTimetable() {
        System.out.println("\n--- 全校周课表 ---");
        Map<TimeSlot.DayOfWeek, List<TimetableService.CourseWithTimeSlot>> timetable = 
                timetableService.getWeeklyTimetable();
        
        for (Map.Entry<TimeSlot.DayOfWeek, List<TimetableService.CourseWithTimeSlot>> entry : timetable.entrySet()) {
            System.out.println("\n=== " + entry.getKey().getDisplayName() + " ===");
            List<TimetableService.CourseWithTimeSlot> courses = entry.getValue();
            
            if (courses.isEmpty()) {
                System.out.println("  (无课程)");
                continue;
            }
            
            for (TimetableService.CourseWithTimeSlot cwt : courses) {
                System.out.printf("  %s | %s | 老师：%s | 教室：%s%n",
                        cwt.getTimeSlot(),
                        cwt.getCourse().getName(),
                        cwt.getCourse().getTeacher().getName(),
                        cwt.getCourse().getClassroom().getRoomNumber());
            }
        }
    }
}
