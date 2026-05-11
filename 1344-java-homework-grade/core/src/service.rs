use crate::errors::AppError;
use crate::models::*;
use std::collections::HashMap;
use std::sync::{Arc, Mutex};

#[derive(Clone, Default)]
pub struct AppState {
    inner: Arc<Mutex<InnerState>>,
}

#[derive(Default)]
struct InnerState {
    users: HashMap<UserId, User>,
    courses: HashMap<CourseId, Course>,
    assignments: HashMap<AssignmentId, Assignment>,
    submissions: HashMap<SubmissionId, Submission>,
    gradings: HashMap<GradingId, Grading>,
}

impl AppState {
    pub fn new() -> Self {
        Self::default()
    }

    fn lock(&self) -> std::sync::MutexGuard<'_, InnerState> {
        self.inner.lock().expect("Lock poisoned")
    }
}

fn now() -> u64 {
    use std::time::{SystemTime, UNIX_EPOCH};
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap()
        .as_secs()
}

fn generate_id(prefix: &str) -> String {
    use rand::{thread_rng, Rng};
    let random: u64 = thread_rng().gen();
    format!("{}-{}", prefix, random)
}

impl AppState {
    pub fn create_user(&self, name: &str, role: Role) -> User {
        let id = generate_id("user");
        let user = User {
            id: id.clone(),
            name: name.to_string(),
            role,
        };
        let mut state = self.lock();
        state.users.insert(id, user.clone());
        user
    }

    pub fn get_user(&self, id: &UserId) -> Result<User, AppError> {
        let state = self.lock();
        state
            .users
            .get(id)
            .cloned()
            .ok_or_else(|| AppError::UserNotFound(id.clone()))
    }

    pub fn list_users(&self) -> Vec<User> {
        let state = self.lock();
        state.users.values().cloned().collect()
    }

    pub fn create_course(&self, name: &str, teacher_ids: Vec<UserId>) -> Result<Course, AppError> {
        let state = self.lock();
        for tid in &teacher_ids {
            let teacher = state
                .users
                .get(tid)
                .ok_or_else(|| AppError::UserNotFound(tid.clone()))?;
            if teacher.role != Role::Teacher {
                return Err(AppError::NotTeacher);
            }
        }
        drop(state);

        let id = generate_id("course");
        let course = Course {
            id: id.clone(),
            name: name.to_string(),
            teacher_ids,
            student_ids: Vec::new(),
            assignment_ids: Vec::new(),
        };
        let mut state = self.lock();
        state.courses.insert(id, course.clone());
        Ok(course)
    }

    pub fn get_course(&self, id: &CourseId) -> Result<Course, AppError> {
        let state = self.lock();
        state
            .courses
            .get(id)
            .cloned()
            .ok_or_else(|| AppError::CourseNotFound(id.clone()))
    }

    pub fn list_courses(&self) -> Vec<Course> {
        let state = self.lock();
        state.courses.values().cloned().collect()
    }

    pub fn add_student_to_course(
        &self,
        course_id: &CourseId,
        student_id: &UserId,
    ) -> Result<(), AppError> {
        let mut state = self.lock();

        let student_role = {
            let student = state
                .users
                .get(student_id)
                .ok_or_else(|| AppError::UserNotFound(student_id.clone()))?;
            student.role.clone()
        };

        if student_role != Role::Student {
            return Err(AppError::NotStudent);
        }

        let course = state
            .courses
            .get_mut(course_id)
            .ok_or_else(|| AppError::CourseNotFound(course_id.clone()))?;

        if !course.student_ids.contains(student_id) {
            course.student_ids.push(student_id.clone());
        }
        Ok(())
    }

    pub fn create_assignment(
        &self,
        course_id: &CourseId,
        title: &str,
        questions: Vec<Question>,
        teacher_id: &UserId,
    ) -> Result<Assignment, AppError> {
        let mut state = self.lock();

        let teacher = state
            .users
            .get(teacher_id)
            .ok_or_else(|| AppError::UserNotFound(teacher_id.clone()))?;

        if teacher.role != Role::Teacher {
            return Err(AppError::NotTeacher);
        }

        let course = state
            .courses
            .get(course_id)
            .ok_or_else(|| AppError::CourseNotFound(course_id.clone()))?;

        if !course.teacher_ids.contains(teacher_id) {
            return Err(AppError::NotCourseTeacher);
        }

        for q in &questions {
            if q.question_type == QuestionType::Objective && q.answer.is_none() {
                return Err(AppError::Internal(
                    "Objective questions must have answers".to_string(),
                ));
            }
        }

        drop(state);

        let id = generate_id("assign");
        let assignment = Assignment {
            id: id.clone(),
            course_id: course_id.clone(),
            title: title.to_string(),
            questions,
            created_at: now(),
        };

        let mut state = self.lock();
        state
            .courses
            .get_mut(course_id)
            .unwrap()
            .assignment_ids
            .push(id.clone());
        state.assignments.insert(id, assignment.clone());
        Ok(assignment)
    }

    pub fn get_assignment(&self, id: &AssignmentId) -> Result<Assignment, AppError> {
        let state = self.lock();
        state
            .assignments
            .get(id)
            .cloned()
            .ok_or_else(|| AppError::AssignmentNotFound(id.clone()))
    }

    pub fn list_assignments(&self, course_id: &CourseId) -> Result<Vec<Assignment>, AppError> {
        let state = self.lock();
        let course = state
            .courses
            .get(course_id)
            .ok_or_else(|| AppError::CourseNotFound(course_id.clone()))?;

        Ok(course
            .assignment_ids
            .iter()
            .filter_map(|aid| state.assignments.get(aid).cloned())
            .collect())
    }

    pub fn submit_assignment(
        &self,
        assignment_id: &AssignmentId,
        student_id: &UserId,
        answers: Vec<Answer>,
    ) -> Result<Submission, AppError> {
        let mut state = self.lock();

        let student = state
            .users
            .get(student_id)
            .ok_or_else(|| AppError::UserNotFound(student_id.clone()))?;

        if student.role != Role::Student {
            return Err(AppError::NotStudent);
        }

        let assignment = state
            .assignments
            .get(assignment_id)
            .ok_or_else(|| AppError::AssignmentNotFound(assignment_id.clone()))?;

        let course = state
            .courses
            .get(&assignment.course_id)
            .ok_or_else(|| AppError::CourseNotFound(assignment.course_id.clone()))?;

        if !course.student_ids.contains(student_id) {
            return Err(AppError::NotEnrolled);
        }

        let existing = state.submissions.values().any(|s| {
            s.assignment_id == *assignment_id && s.student_id == *student_id
        });

        if existing {
            return Err(AppError::AlreadySubmitted);
        }

        let mut objective_scores = HashMap::new();
        for q in &assignment.questions {
            if q.question_type == QuestionType::Objective {
                let student_answer = answers
                    .iter()
                    .find(|a| a.question_id == q.id)
                    .map(|a| a.answer.clone())
                    .unwrap_or_default();

                let score = if Some(&student_answer) == q.answer.as_ref() {
                    q.max_score
                } else {
                    0
                };
                objective_scores.insert(q.id.clone(), score);
            }
        }

        let has_subjective = assignment
            .questions
            .iter()
            .any(|q| q.question_type == QuestionType::Subjective);

        let id = generate_id("sub");
        let submission = Submission {
            id: id.clone(),
            assignment_id: assignment_id.clone(),
            student_id: student_id.clone(),
            answers,
            submitted_at: now(),
            objective_scores,
            is_completed: !has_subjective,
            grading_ids: Vec::new(),
            has_appealed: false,
        };

        state.submissions.insert(id, submission.clone());
        Ok(submission)
    }

    pub fn get_submission(&self, id: &SubmissionId) -> Result<Submission, AppError> {
        let state = self.lock();
        state
            .submissions
            .get(id)
            .cloned()
            .ok_or_else(|| AppError::SubmissionNotFound(id.clone()))
    }

    pub fn list_submissions(
        &self,
        assignment_id: &AssignmentId,
    ) -> Result<Vec<Submission>, AppError> {
        let state = self.lock();
        Ok(state
            .submissions
            .values()
            .filter(|s| s.assignment_id == *assignment_id)
            .cloned()
            .collect())
    }

    pub fn grade_subjective(
        &self,
        submission_id: &SubmissionId,
        teacher_id: &UserId,
        scores: HashMap<QuestionId, u32>,
    ) -> Result<Grading, AppError> {
        let mut state = self.lock();

        let teacher = state
            .users
            .get(teacher_id)
            .ok_or_else(|| AppError::UserNotFound(teacher_id.clone()))?;

        if teacher.role != Role::Teacher {
            return Err(AppError::NotTeacher);
        }

        let submission = state
            .submissions
            .get(submission_id)
            .ok_or_else(|| AppError::SubmissionNotFound(submission_id.clone()))?;

        let assignment = state
            .assignments
            .get(&submission.assignment_id)
            .ok_or_else(|| AppError::AssignmentNotFound(submission.assignment_id.clone()))?;

        let course = state
            .courses
            .get(&assignment.course_id)
            .ok_or_else(|| AppError::CourseNotFound(assignment.course_id.clone()))?;

        if !course.teacher_ids.contains(teacher_id) {
            return Err(AppError::NotCourseTeacher);
        }

        if !submission.grading_ids.is_empty() {
            let existing_grading = state
                .gradings
                .get(&submission.grading_ids[0])
                .ok_or_else(|| AppError::GradingNotFound(submission.grading_ids[0].clone()))?;

            if existing_grading.teacher_id == *teacher_id {
                return Err(AppError::AlreadyGraded);
            }
        }

        for (qid, score) in &scores {
            let question = assignment
                .questions
                .iter()
                .find(|q| q.id == *qid)
                .ok_or_else(|| AppError::QuestionNotFound(qid.clone()))?;

            if question.question_type != QuestionType::Subjective {
                return Err(AppError::Internal("Can only grade subjective questions".to_string()));
            }

            if *score > question.max_score {
                return Err(AppError::ScoreExceedsMax { max: question.max_score });
            }
        }

        for q in &assignment.questions {
            if q.question_type == QuestionType::Subjective && !scores.contains_key(&q.id) {
                return Err(AppError::Internal(format!(
                    "Missing score for question {}",
                    q.id
                )));
            }
        }

        let id = generate_id("grade");
        let grading = Grading {
            id: id.clone(),
            submission_id: submission_id.clone(),
            teacher_id: teacher_id.clone(),
            status: GradingStatus::Initial,
            subjective_scores: scores,
            graded_at: Some(now()),
            is_final: true,
        };

        let submission = state.submissions.get_mut(submission_id).unwrap();
        submission.grading_ids.push(id.clone());
        submission.is_completed = true;

        state.gradings.insert(id, grading.clone());
        Ok(grading)
    }

    pub fn get_grading(&self, id: &GradingId) -> Result<Grading, AppError> {
        let state = self.lock();
        state
            .gradings
            .get(id)
            .cloned()
            .ok_or_else(|| AppError::GradingNotFound(id.clone()))
    }

    pub fn list_gradings(&self, submission_id: &SubmissionId) -> Result<Vec<Grading>, AppError> {
        let state = self.lock();
        Ok(state
            .gradings
            .values()
            .filter(|g| g.submission_id == *submission_id)
            .cloned()
            .collect())
    }

    pub fn appeal_submission(
        &self,
        submission_id: &SubmissionId,
        student_id: &UserId,
    ) -> Result<(), AppError> {
        let mut state = self.lock();

        let grading_ids_to_update: Vec<GradingId>;
        {
            let submission = state
                .submissions
                .get(submission_id)
                .ok_or_else(|| AppError::SubmissionNotFound(submission_id.clone()))?;

            if submission.student_id != *student_id {
                return Err(AppError::NotEnrolled);
            }

            if submission.has_appealed {
                return Err(AppError::AlreadyAppealed);
            }

            if submission.grading_ids.is_empty() {
                return Err(AppError::Internal("Submission not graded yet".to_string()));
            }

            grading_ids_to_update = submission.grading_ids.clone();
        }

        for gid in &grading_ids_to_update {
            if let Some(g) = state.gradings.get_mut(gid) {
                g.is_final = false;
            }
        }

        let submission = state.submissions.get_mut(submission_id).unwrap();
        submission.has_appealed = true;
        Ok(())
    }

    pub fn grade_appeal(
        &self,
        submission_id: &SubmissionId,
        teacher_id: &UserId,
        scores: HashMap<QuestionId, u32>,
    ) -> Result<Grading, AppError> {
        let mut state = self.lock();

        let teacher = state
            .users
            .get(teacher_id)
            .ok_or_else(|| AppError::UserNotFound(teacher_id.clone()))?;

        if teacher.role != Role::Teacher {
            return Err(AppError::NotTeacher);
        }

        let submission = state
            .submissions
            .get(submission_id)
            .ok_or_else(|| AppError::SubmissionNotFound(submission_id.clone()))?;

        if !submission.has_appealed {
            return Err(AppError::Internal("No appeal for this submission".to_string()));
        }

        let has_final = submission
            .grading_ids
            .iter()
            .filter_map(|gid| state.gradings.get(gid))
            .any(|g| g.is_final);

        if has_final {
            return Err(AppError::AlreadyGraded);
        }

        let original_teacher_ids: Vec<UserId> = submission
            .grading_ids
            .iter()
            .filter_map(|gid| state.gradings.get(gid).map(|g| g.teacher_id.clone()))
            .collect();

        if original_teacher_ids.contains(teacher_id) {
            return Err(AppError::MustBeDifferentTeacher);
        }

        let assignment = state
            .assignments
            .get(&submission.assignment_id)
            .ok_or_else(|| AppError::AssignmentNotFound(submission.assignment_id.clone()))?;

        let course = state
            .courses
            .get(&assignment.course_id)
            .ok_or_else(|| AppError::CourseNotFound(assignment.course_id.clone()))?;

        if !course.teacher_ids.contains(teacher_id) {
            return Err(AppError::NotCourseTeacher);
        }

        for (qid, score) in &scores {
            let question = assignment
                .questions
                .iter()
                .find(|q| q.id == *qid)
                .ok_or_else(|| AppError::QuestionNotFound(qid.clone()))?;

            if question.question_type != QuestionType::Subjective {
                return Err(AppError::Internal("Can only grade subjective questions".to_string()));
            }

            if *score > question.max_score {
                return Err(AppError::ScoreExceedsMax { max: question.max_score });
            }
        }

        for q in &assignment.questions {
            if q.question_type == QuestionType::Subjective && !scores.contains_key(&q.id) {
                return Err(AppError::Internal(format!(
                    "Missing score for question {}",
                    q.id
                )));
            }
        }

        let id = generate_id("grade");
        let grading = Grading {
            id: id.clone(),
            submission_id: submission_id.clone(),
            teacher_id: teacher_id.clone(),
            status: GradingStatus::Appeal,
            subjective_scores: scores,
            graded_at: Some(now()),
            is_final: true,
        };

        let submission = state.submissions.get_mut(submission_id).unwrap();
        submission.grading_ids.push(id.clone());
        submission.is_completed = true;

        state.gradings.insert(id, grading.clone());
        Ok(grading)
    }

    pub fn get_course_statistics(&self, course_id: &CourseId) -> Result<CourseStatistics, AppError> {
        let state = self.lock();

        let course = state
            .courses
            .get(course_id)
            .ok_or_else(|| AppError::CourseNotFound(course_id.clone()))?;

        let mut all_scores = Vec::new();
        let mut pass_count = 0;

        for aid in &course.assignment_ids {
            let assignment = match state.assignments.get(aid) {
                Some(a) => a,
                None => continue,
            };
            let max_score = assignment
                .questions
                .iter()
                .map(|q| q.max_score)
                .sum::<u32>();

            for submission in state
                .submissions
                .values()
                .filter(|s| s.assignment_id == *aid)
            {
                if let Some(score) = submission.get_final_score(
                    &submission
                        .grading_ids
                        .iter()
                        .filter_map(|gid| state.gradings.get(gid).cloned())
                        .collect::<Vec<_>>(),
                ) {
                    let percentage = if max_score > 0 {
                        (score as f64 / max_score as f64) * 100.0
                    } else {
                        0.0
                    };
                    all_scores.push(score);
                    if percentage >= 60.0 {
                        pass_count += 1;
                    }
                }
            }
        }

        let total = all_scores.len();
        let (avg_score, max_score, min_score) = if total > 0 {
            let avg = all_scores.iter().sum::<u32>() as f64 / total as f64;
            let max = *all_scores.iter().max().unwrap_or(&0);
            let min = *all_scores.iter().min().unwrap_or(&0);
            (avg, max, min)
        } else {
            (0.0, 0, 0)
        };

        let pass_rate = if total > 0 {
            pass_count as f64 / total as f64
        } else {
            0.0
        };

        Ok(CourseStatistics {
            course_id: course_id.clone(),
            course_name: course.name.clone(),
            avg_score,
            max_score,
            min_score,
            pass_rate,
            total_submissions: total,
            pass_submissions: pass_count,
        })
    }

    pub fn get_score_distribution(&self, course_id: &CourseId) -> Result<ScoreDistribution, AppError> {
        let state = self.lock();

        let course = state
            .courses
            .get(course_id)
            .ok_or_else(|| AppError::CourseNotFound(course_id.clone()))?;

        let mut dist = ScoreDistribution {
            range_0_59: 0,
            range_60_69: 0,
            range_70_79: 0,
            range_80_89: 0,
            range_90_100: 0,
        };

        for aid in &course.assignment_ids {
            let assignment = match state.assignments.get(aid) {
                Some(a) => a,
                None => continue,
            };
            let max_score = assignment
                .questions
                .iter()
                .map(|q| q.max_score)
                .sum::<u32>();

            for submission in state
                .submissions
                .values()
                .filter(|s| s.assignment_id == *aid)
            {
                if let Some(score) = submission.get_final_score(
                    &submission
                        .grading_ids
                        .iter()
                        .filter_map(|gid| state.gradings.get(gid).cloned())
                        .collect::<Vec<_>>(),
                ) {
                    let percentage = if max_score > 0 {
                        (score as f64 / max_score as f64) * 100.0
                    } else {
                        0.0
                    };

                    match percentage {
                        p if p < 60.0 => dist.range_0_59 += 1,
                        p if p < 70.0 => dist.range_60_69 += 1,
                        p if p < 80.0 => dist.range_70_79 += 1,
                        p if p < 90.0 => dist.range_80_89 += 1,
                        _ => dist.range_90_100 += 1,
                    }
                }
            }
        }

        Ok(dist)
    }

    pub fn get_student_trends(
        &self,
        course_id: &CourseId,
    ) -> Result<Vec<StudentTrend>, AppError> {
        let state = self.lock();

        let course = state
            .courses
            .get(course_id)
            .ok_or_else(|| AppError::CourseNotFound(course_id.clone()))?;

        let mut trends = Vec::new();

        for sid in &course.student_ids {
            let student = match state.users.get(sid) {
                Some(s) => s,
                None => continue,
            };

            let mut assignment_trends = Vec::new();

            for aid in &course.assignment_ids {
                let assignment = match state.assignments.get(aid) {
                    Some(a) => a,
                    None => continue,
                };

                let max_score = assignment
                    .questions
                    .iter()
                    .map(|q| q.max_score)
                    .sum::<u32>();

                let submission = state.submissions.values().find(|s| {
                    s.assignment_id == *aid && s.student_id == *sid
                });

                if let Some(sub) = submission {
                    if let Some(score) = sub.get_final_score(
                        &sub.grading_ids
                            .iter()
                            .filter_map(|gid| state.gradings.get(gid).cloned())
                            .collect::<Vec<_>>(),
                    ) {
                        let percentage = if max_score > 0 {
                            (score as f64 / max_score as f64) * 100.0
                        } else {
                            0.0
                        };

                        assignment_trends.push(AssignmentTrend {
                            assignment_id: aid.clone(),
                            assignment_title: assignment.title.clone(),
                            total_score: score,
                            percentage,
                        });
                    }
                }
            }

            trends.push(StudentTrend {
                student_id: sid.clone(),
                student_name: student.name.clone(),
                assignments: assignment_trends,
            });
        }

        Ok(trends)
    }

    pub fn get_teacher_workload(&self) -> Result<Vec<TeacherWorkload>, AppError> {
        let state = self.lock();

        let teachers: Vec<User> = state
            .users
            .values()
            .filter(|u| u.role == Role::Teacher)
            .cloned()
            .collect();

        let mut workloads = Vec::new();

        for teacher in teachers {
            let gradings: Vec<&Grading> = state
                .gradings
                .values()
                .filter(|g| g.teacher_id == teacher.id)
                .collect();

            let total_gradings = gradings.len();

            let avg_score = if total_gradings > 0 {
                let total: u32 = gradings
                    .iter()
                    .map(|g| g.subjective_scores.values().sum::<u32>())
                    .sum();
                total as f64 / total_gradings as f64
            } else {
                0.0
            };

            workloads.push(TeacherWorkload {
                teacher_id: teacher.id.clone(),
                teacher_name: teacher.name.clone(),
                total_gradings,
                avg_score,
            });
        }

        Ok(workloads)
    }

    pub fn get_submissions_by_student(
        &self,
        student_id: &UserId,
    ) -> Result<Vec<Submission>, AppError> {
        let state = self.lock();

        let student = state
            .users
            .get(student_id)
            .ok_or_else(|| AppError::UserNotFound(student_id.clone()))?;

        if student.role != Role::Student {
            return Err(AppError::NotStudent);
        }

        Ok(state
            .submissions
            .values()
            .filter(|s| s.student_id == *student_id)
            .cloned()
            .collect())
    }
}
