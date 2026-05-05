package com.training.server.service;

import com.training.common.enums.CourseStatus;
import com.training.common.enums.RegistrationStatus;
import com.training.server.entity.Course;
import com.training.server.entity.Instructor;
import com.training.server.entity.Registration;
import com.training.server.repository.CourseRepository;
import com.training.server.repository.InstructorRepository;
import com.training.server.repository.RegistrationRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.time.temporal.ChronoUnit;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;

@Service
public class ScheduledTaskService {

    private static final Logger logger = LoggerFactory.getLogger(ScheduledTaskService.class);

    @Autowired
    private CourseRepository courseRepository;

    @Autowired
    private RegistrationRepository registrationRepository;

    @Autowired
    private InstructorRepository instructorRepository;

    private final Map<String, Integer> reminderCounts = new ConcurrentHashMap<>();

    @Scheduled(cron = "0 0 9 * * ?")
    public void checkIncompleteScores() {
        logger.info("开始执行定时任务：检查课程结束超过7天未打分的情况");
        
        LocalDateTime now = LocalDateTime.now();
        List<String> coursesToRemind = new ArrayList<>();

        List<Course> completedCourses = courseRepository.findByStatus(CourseStatus.COMPLETED);
        
        for (Course course : completedCourses) {
            if (course.getEndTime() == null) {
                continue;
            }

            long daysSinceEnd = ChronoUnit.DAYS.between(course.getEndTime(), now);
            
            if (daysSinceEnd > 7) {
                List<Registration> registrations = registrationRepository.findByCourseId(course.getId());
                List<Registration> unScoredRegistrations = new ArrayList<>();
                
                for (Registration registration : registrations) {
                    if (registration.getStatus() == RegistrationStatus.REGISTERED) {
                        unScoredRegistrations.add(registration);
                    }
                }

                if (!unScoredRegistrations.isEmpty()) {
                    coursesToRemind.add(course.getId());
                    int count = reminderCounts.getOrDefault(course.getId(), 0) + 1;
                    reminderCounts.put(course.getId(), count);

                    Optional<Instructor> instructorOpt = instructorRepository.findById(course.getInstructorId());
                    String instructorName = instructorOpt.map(Instructor::getName).orElse("未知讲师");

                    logger.warn("【自动提醒】第{}次提醒：课程 '{}' (ID: {}) 已结束{}天，还有{}名学员未打分。讲师：{}",
                            count,
                            course.getName(),
                            course.getId(),
                            daysSinceEnd,
                            unScoredRegistrations.size(),
                            instructorName);
                    
                    for (Registration registration : unScoredRegistrations) {
                        logger.info("  - 待打分学员报名ID: {}, 员工ID: {}", 
                                registration.getId(), 
                                registration.getEmployeeId());
                    }
                }
            }
        }

        if (coursesToRemind.isEmpty()) {
            logger.info("定时任务执行完成：无需提醒，所有已结束课程都已完成打分");
        } else {
            logger.info("定时任务执行完成：共提醒 {} 门课程", coursesToRemind.size());
        }
    }

    public Map<String, Integer> getReminderCounts() {
        return new ConcurrentHashMap<>(reminderCounts);
    }

    public void clearReminderCount(String courseId) {
        reminderCounts.remove(courseId);
    }
}
