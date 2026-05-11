use std::collections::{HashMap, HashSet};
use std::sync::Arc;
use chrono::{DateTime, Utc, Duration};
use dashmap::DashMap;
use crate::error::CourseSystemError;
use crate::models::*;

pub const MAX_CREDITS_PER_SEMESTER: u32 = 25;
pub const WITHDRAWAL_WINDOW_DAYS: i64 = 14;

pub struct CourseService {
    courses: DashMap<CourseId, Course>,
    students: DashMap<StudentId, Student>,
    enrollments: DashMap<(StudentId, CourseId), Enrollment>,
    withdrawals: DashMap<(StudentId, CourseId), Withdrawal>,
    semester_start: DateTime<Utc>,
}

impl CourseService {
    pub fn new(semester_start: DateTime<Utc>) -> Self {
        Self {
            courses: DashMap::new(),
            students: DashMap::new(),
            enrollments: DashMap::new(),
            withdrawals: DashMap::new(),
            semester_start,
        }
    }

    pub fn add_course(&self, course: Course) {
        self.courses.insert(course.id.clone(), course);
    }

    pub fn get_course(&self, id: &CourseId) -> Option<Course> {
        self.courses.get(id).map(|c| c.clone())
    }

    pub fn list_courses(&self) -> Vec<Course> {
        self.courses.iter().map(|c| c.clone()).collect()
    }

    pub fn add_student(&self, student: Student) {
        self.students.insert(student.id.clone(), student);
    }

    pub fn get_student(&self, id: &StudentId) -> Option<Student> {
        self.students.get(id).map(|s| s.clone())
    }

    pub fn list_students(&self) -> Vec<Student> {
        self.students.iter().map(|s| s.clone()).collect()
    }

    fn get_all_prerequisites(&self, course_id: &CourseId) -> Result<Vec<CourseId>, CourseSystemError> {
        let mut visited = HashSet::new();
        let mut prereqs = Vec::new();
        self.collect_prerequisites(course_id, &mut visited, &mut prereqs)?;
        Ok(prereqs)
    }

    fn collect_prerequisites(
        &self,
        course_id: &CourseId,
        visited: &mut HashSet<CourseId>,
        prereqs: &mut Vec<CourseId>,
    ) -> Result<(), CourseSystemError> {
        if !visited.insert(course_id.clone()) {
            return Ok(());
        }

        let course = self.get_course(course_id)
            .ok_or_else(|| CourseSystemError::CourseNotFound(course_id.0.clone()))?;

        for prereq_id in &course.prerequisites {
            self.collect_prerequisites(prereq_id, visited, prereqs)?;
            prereqs.push(prereq_id.clone());
        }

        Ok(())
    }

    fn check_prerequisites(&self, student_id: &StudentId, course_id: &CourseId) -> Result<(), CourseSystemError> {
        let student = self.get_student(student_id)
            .ok_or_else(|| CourseSystemError::StudentNotFound(student_id.0.clone()))?;

        let completed: HashMap<&CourseId, f64> = student.completed_courses
            .iter()
            .map(|c| (&c.course_id, c.score))
            .collect();

        let all_prereqs = self.get_all_prerequisites(course_id)?;

        for prereq_id in &all_prereqs {
            match completed.get(prereq_id) {
                Some(&score) if score >= 60.0 => continue,
                _ => {
                    let course = self.get_course(prereq_id)
                        .ok_or_else(|| CourseSystemError::CourseNotFound(prereq_id.0.clone()))?;
                    return Err(CourseSystemError::PrerequisiteNotSatisfied(
                        format!("{} ({})", course.name, prereq_id.0)
                    ));
                }
            }
        }

        Ok(())
    }

    fn get_enrolled_count(&self, course_id: &CourseId) -> usize {
        self.enrollments.iter()
            .filter(|e| e.course_id == *course_id)
            .count()
    }

    fn get_current_credits(&self, student_id: &StudentId) -> u32 {
        self.enrollments.iter()
            .filter(|e| e.student_id == *student_id)
            .filter_map(|e| self.get_course(&e.course_id))
            .map(|c| c.credits)
            .sum()
    }

    pub fn enroll(&self, student_id: &StudentId, course_id: &CourseId) -> Result<(), CourseSystemError> {
        let course = self.get_course(course_id)
            .ok_or_else(|| CourseSystemError::CourseNotFound(course_id.0.clone()))?;

        let _student = self.get_student(student_id)
            .ok_or_else(|| CourseSystemError::StudentNotFound(student_id.0.clone()))?;

        if self.enrollments.contains_key(&(student_id.clone(), course_id.clone())) {
            return Err(CourseSystemError::AlreadyEnrolled);
        }

        self.check_prerequisites(student_id, course_id)?;

        let current_credits = self.get_current_credits(student_id);
        if current_credits + course.credits > MAX_CREDITS_PER_SEMESTER {
            return Err(CourseSystemError::CreditLimitExceeded {
                current: current_credits,
                max: MAX_CREDITS_PER_SEMESTER,
                adding: course.credits,
            });
        }

        if self.get_enrolled_count(course_id) >= course.capacity as usize {
            return Err(CourseSystemError::CourseFull);
        }

        let enrollment = Enrollment {
            student_id: student_id.clone(),
            course_id: course_id.clone(),
            enrolled_at: Utc::now(),
        };

        self.enrollments.insert(
            (student_id.clone(), course_id.clone()),
            enrollment,
        );

        Ok(())
    }

    pub fn withdraw(&self, student_id: &StudentId, course_id: &CourseId) -> Result<bool, CourseSystemError> {
        if !self.enrollments.contains_key(&(student_id.clone(), course_id.clone())) {
            return Err(CourseSystemError::NotEnrolled);
        }

        self.enrollments.remove(&(student_id.clone(), course_id.clone()));

        let now = Utc::now();
        let deadline = self.semester_start + Duration::days(WITHDRAWAL_WINDOW_DAYS);
        let should_record = now > deadline;

        let key = (student_id.clone(), course_id.clone());
        if should_record {
            if !self.withdrawals.contains_key(&key) {
                let withdrawal = Withdrawal {
                    student_id: student_id.clone(),
                    course_id: course_id.clone(),
                    withdrawn_at: now,
                    recorded: true,
                };
                self.withdrawals.insert(key, withdrawal);
            }
        }

        Ok(should_record)
    }

    pub fn get_enrollment_result(&self, student_id: &StudentId) -> Result<EnrollmentResult, CourseSystemError> {
        let _student = self.get_student(student_id)
            .ok_or_else(|| CourseSystemError::StudentNotFound(student_id.0.clone()))?;

        let enrolled_courses: Vec<Course> = self.enrollments.iter()
            .filter(|e| e.student_id == *student_id)
            .filter_map(|e| self.get_course(&e.course_id))
            .collect();

        let total_credits = enrolled_courses.iter().map(|c| c.credits).sum();

        Ok(EnrollmentResult {
            student_id: student_id.clone(),
            enrolled_courses,
            total_credits,
        })
    }

    pub fn get_transcript(&self, student_id: &StudentId) -> Result<Transcript, CourseSystemError> {
        let student = self.get_student(student_id)
            .ok_or_else(|| CourseSystemError::StudentNotFound(student_id.0.clone()))?;

        let mut entries = Vec::new();

        for completed in &student.completed_courses {
            if let Some(course) = self.get_course(&completed.course_id) {
                entries.push(TranscriptEntry::Completed {
                    course_id: completed.course_id.clone(),
                    course_name: course.name.clone(),
                    score: completed.score,
                    credits: course.credits,
                });
            }
        }

        for withdrawal in self.withdrawals.iter() {
            if withdrawal.student_id == *student_id && withdrawal.recorded {
                if let Some(course) = self.get_course(&withdrawal.course_id) {
                    entries.push(TranscriptEntry::Withdrawal {
                        course_id: withdrawal.course_id.clone(),
                        course_name: course.name.clone(),
                    });
                }
            }
        }

        Ok(Transcript {
            student_id: student_id.clone(),
            courses: entries,
        })
    }
}

pub type SharedCourseService = Arc<CourseService>;
