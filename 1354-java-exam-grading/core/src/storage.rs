use crate::models::*;
use crate::errors::GradingError;
use crate::grading::*;
use crate::task_assignment::*;
use crate::review::*;
use std::collections::HashMap;
use std::sync::Mutex;
use uuid::Uuid;

pub struct InMemoryStorage {
    pub teachers: Mutex<HashMap<Uuid, Teacher>>,
    pub students: Mutex<HashMap<Uuid, Student>>,
    pub exams: Mutex<HashMap<Uuid, ExamPaper>>,
    pub student_exams: Mutex<HashMap<Uuid, StudentExam>>,
    pub tasks: Mutex<HashMap<Uuid, GradingTask>>,
    pub review_requests: Mutex<HashMap<Uuid, ReviewRequest>>,
    pub anomalies: Mutex<HashMap<Uuid, Anomaly>>,
    pub task_assigner: Mutex<TaskAssignmentSystem>,
    pub score_diff_threshold: u32,
}

impl InMemoryStorage {
    pub fn new() -> Self {
        Self {
            teachers: Mutex::new(HashMap::new()),
            students: Mutex::new(HashMap::new()),
            exams: Mutex::new(HashMap::new()),
            student_exams: Mutex::new(HashMap::new()),
            tasks: Mutex::new(HashMap::new()),
            review_requests: Mutex::new(HashMap::new()),
            anomalies: Mutex::new(HashMap::new()),
            task_assigner: Mutex::new(TaskAssignmentSystem::new()),
            score_diff_threshold: DEFAULT_SCORE_DIFF_THRESHOLD,
        }
    }

    pub fn with_threshold(threshold: u32) -> Self {
        Self {
            score_diff_threshold: threshold,
            ..Self::new()
        }
    }

    pub fn add_teacher(&self, teacher: Teacher) {
        self.teachers
            .lock()
            .unwrap()
            .insert(teacher.id, teacher.clone());
        self.task_assigner
            .lock()
            .unwrap()
            .teacher_workload
            .entry(teacher.id)
            .or_insert(0);
    }

    pub fn add_student(&self, student: Student) {
        self.students
            .lock()
            .unwrap()
            .insert(student.id, student);
    }

    pub fn add_exam(&self, exam: ExamPaper) {
        self.exams.lock().unwrap().insert(exam.id, exam);
    }

    pub fn get_exam(&self, exam_id: Uuid) -> Result<ExamPaper, GradingError> {
        self.exams
            .lock()
            .unwrap()
            .get(&exam_id)
            .cloned()
            .ok_or(GradingError::ExamNotFound(exam_id))
    }

    pub fn submit_student_exam(
        &self,
        student_id: Uuid,
        exam_id: Uuid,
        answers: HashMap<Uuid, Answer>,
    ) -> Result<Uuid, GradingError> {
        if !self.students.lock().unwrap().contains_key(&student_id) {
            return Err(GradingError::StudentNotFound(student_id));
        }
        
        if !self.exams.lock().unwrap().contains_key(&exam_id) {
            return Err(GradingError::ExamNotFound(exam_id));
        }

        let student_exam = StudentExam::new(student_id, exam_id, answers);
        let id = student_exam.id;
        self.student_exams
            .lock()
            .unwrap()
            .insert(id, student_exam);
        Ok(id)
    }

    pub fn grade_all_objective(&self, student_exam_id: Uuid) -> Result<(), GradingError> {
        let mut student_exams = self.student_exams.lock().unwrap();
        let student_exam = student_exams
            .get_mut(&student_exam_id)
            .ok_or(GradingError::StudentExamNotFound(student_exam_id))?;

        let exams = self.exams.lock().unwrap();
        let exam = exams
            .get(&student_exam.exam_id)
            .ok_or(GradingError::ExamNotFound(student_exam.exam_id))?;

        for question in &exam.questions {
            if question.question_type != QuestionType::Subjective {
                let answer = student_exam
                    .answers
                    .get(&question.id)
                    .ok_or(GradingError::AnswerNotFound(question.id))?;
                
                let score = grade_objective_question(question, answer)?;
                student_exam.objective_scores.insert(question.id, score);
            } else {
                student_exam.subjective_scores.insert(
                    question.id,
                    SubjectiveScore::new(question.id),
                );
            }
        }

        student_exam.status = ExamStatus::PendingSubjectiveGrading;
        Ok(())
    }

    pub fn assign_all_subjective_tasks(&self, exam_id: Uuid) -> Result<Vec<Uuid>, GradingError> {
        let teachers: Vec<Teacher> = self
            .teachers
            .lock()
            .unwrap()
            .values()
            .cloned()
            .collect();

        if teachers.is_empty() {
            return Err(GradingError::NoEligibleTeachers);
        }

        let exam = self.get_exam(exam_id)?;

        let student_exams_list: Vec<StudentExam> = self
            .student_exams
            .lock()
            .unwrap()
            .values()
            .filter(|se| se.exam_id == exam_id)
            .cloned()
            .collect();

        let students = self.students.lock().unwrap().clone();

        let mut task_assigner = self.task_assigner.lock().unwrap();
        let tasks = task_assigner.assign_subjective_tasks(
            &teachers,
            &exam,
            &student_exams_list,
            &students,
        )?;

        let task_ids: Vec<Uuid> = tasks.iter().map(|t| t.id).collect();
        
        let mut all_tasks = self.tasks.lock().unwrap();
        for task in tasks {
            all_tasks.insert(task.id, task);
        }

        Ok(task_ids)
    }

    pub fn get_teacher_tasks(&self, teacher_id: Uuid) -> Vec<GradingTask> {
        self.tasks
            .lock()
            .unwrap()
            .values()
            .filter(|t| t.teacher_id == teacher_id && t.status == TaskStatus::Assigned)
            .cloned()
            .collect()
    }

    pub fn submit_grade(
        &self,
        task_id: Uuid,
        score: u32,
        comments: Option<String>,
    ) -> Result<(), GradingError> {
        let task_info = {
            let tasks = self.tasks.lock().unwrap();
            let task = tasks
                .get(&task_id)
                .ok_or(GradingError::TaskNotFound(task_id))?;

            if task.status == TaskStatus::Completed {
                return Err(GradingError::TaskAlreadyCompleted);
            }
            
            (task.student_exam_id, task.question_id, task.teacher_id)
        };
        
        let (student_exam_id, question_id, teacher_id) = task_info;
        
        let (question_clone, student_clone, need_third_review_info) = {
            let exams = self.exams.lock().unwrap();
            let students = self.students.lock().unwrap();
            let mut student_exams = self.student_exams.lock().unwrap();
            
            let student_exam = student_exams
                .get_mut(&student_exam_id)
                .ok_or(GradingError::StudentExamNotFound(student_exam_id))?;

            let exam = exams
                .get(&student_exam.exam_id)
                .ok_or(GradingError::ExamNotFound(student_exam.exam_id))?;

            let question = exam
                .questions
                .iter()
                .find(|q| q.id == question_id)
                .ok_or(GradingError::QuestionNotFound(question_id))?;

            let score_record = student_exam
                .subjective_scores
                .get_mut(&question_id)
                .ok_or(GradingError::QuestionNotFound(question_id))?;

            let threshold = self.score_diff_threshold;
            let question_clone = question.clone();
            let student_clone = students
                .get(&student_exam.student_id)
                .ok_or(GradingError::StudentNotFound(student_exam.student_id))?
                .clone();
            
            let need_third = match score_record.status {
                SubjectiveScoreStatus::PendingFirstReview => {
                    grade_subjective_first(
                        score_record,
                        teacher_id,
                        score,
                        question.max_score,
                        comments.clone(),
                        threshold,
                    )?;
                    None
                }
                SubjectiveScoreStatus::PendingSecondReview => {
                    grade_subjective_second(
                        score_record,
                        teacher_id,
                        score,
                        question.max_score,
                        comments.clone(),
                        threshold,
                    )?;
                    
                    if score_record.status == SubjectiveScoreStatus::PendingThirdReview {
                        let first = score_record.first_score.as_ref().unwrap().teacher_id;
                        let second = score_record.second_score.as_ref().unwrap().teacher_id;
                        Some((first, second, student_exam.id, question.id))
                    } else {
                        None
                    }
                }
                SubjectiveScoreStatus::PendingThirdReview => {
                    grade_subjective_third(
                        score_record,
                        teacher_id,
                        score,
                        question.max_score,
                        comments.clone(),
                    )?;
                    None
                }
                SubjectiveScoreStatus::Completed => {
                    return Err(GradingError::TaskAlreadyCompleted);
                }
            };
            
            let all_completed = student_exam
                .subjective_scores
                .values()
                .all(|s| s.status == SubjectiveScoreStatus::Completed);

            if all_completed {
                let total = calculate_total_score(student_exam)?;
                student_exam.total_score = Some(total);
                student_exam.status = ExamStatus::GradingCompleted;
            }
            
            (question_clone, student_clone, need_third)
        };
        
        if let Some((first_teacher, second_teacher, se_id, q_id)) = need_third_review_info {
            let teachers: Vec<Teacher> = self
                .teachers
                .lock()
                .unwrap()
                .values()
                .cloned()
                .collect();
            
            let third_teacher_id = self
                .task_assigner
                .lock()
                .unwrap()
                .select_third_eligible_teacher(
                    &teachers,
                    &question_clone,
                    &student_clone,
                    first_teacher,
                    second_teacher,
                )?;
            
            let third_task = GradingTask {
                id: Uuid::new_v4(),
                student_exam_id: se_id,
                question_id: q_id,
                teacher_id: third_teacher_id,
                status: TaskStatus::Assigned,
            };
            
            self.tasks
                .lock()
                .unwrap()
                .insert(third_task.id, third_task);
        }
        
        {
            let mut tasks = self.tasks.lock().unwrap();
            let task = tasks
                .get_mut(&task_id)
                .ok_or(GradingError::TaskNotFound(task_id))?;
            task.status = TaskStatus::Completed;
        }

        Ok(())
    }

    pub fn mark_anomaly(
        &self,
        student_exam_id: Uuid,
        anomaly_type: AnomalyType,
        reported_by: Uuid,
        description: String,
        related_student_exam_id: Option<Uuid>,
    ) -> Result<Uuid, GradingError> {
        let anomaly = Anomaly {
            id: Uuid::new_v4(),
            anomaly_type,
            reported_by,
            description,
            reviewed: false,
            related_student_exam_id,
        };

        let anomaly_id = anomaly.id;
        
        {
            let mut student_exams = self.student_exams.lock().unwrap();
            let student_exam = student_exams
                .get_mut(&student_exam_id)
                .ok_or(GradingError::StudentExamNotFound(student_exam_id))?;

            student_exam.anomalies.push(anomaly.clone());
            student_exam.status = ExamStatus::NeedsReview;
            
            if anomaly_type == AnomalyType::Plagiarism {
                student_exam.total_score = Some(0);
                for score in student_exam.objective_scores.values_mut() {
                    *score = 0;
                }
                for score in student_exam.subjective_scores.values_mut() {
                    score.final_score = Some(0);
                    score.status = SubjectiveScoreStatus::Completed;
                }
            }
        }
        
        self.anomalies
            .lock()
            .unwrap()
            .insert(anomaly_id, anomaly);

        if anomaly_type == AnomalyType::Plagiarism {
            if let Some(related_id) = related_student_exam_id {
                let mut student_exams = self.student_exams.lock().unwrap();
                if let Some(related_exam) = student_exams.get_mut(&related_id) {
                    related_exam.total_score = Some(0);
                    for score in related_exam.objective_scores.values_mut() {
                        *score = 0;
                    }
                    for score in related_exam.subjective_scores.values_mut() {
                        score.final_score = Some(0);
                        score.status = SubjectiveScoreStatus::Completed;
                    }
                    related_exam.status = ExamStatus::NeedsReview;
                }
            }
        }

        Ok(anomaly_id)
    }

    pub fn publish_scores(&self, exam_id: Uuid) -> Result<(), GradingError> {
        let mut student_exams = self.student_exams.lock().unwrap();
        
        for student_exam in student_exams.values_mut() {
            if student_exam.exam_id == exam_id {
                if student_exam.status != ExamStatus::GradingCompleted {
                    return Err(GradingError::InvalidStateTransition);
                }
                student_exam.is_published = true;
                student_exam.status = ExamStatus::Completed;
            }
        }
        
        Ok(())
    }

    pub fn request_review(
        &self,
        student_exam_id: Uuid,
        requested_by: Uuid,
        reason: String,
    ) -> Result<Uuid, GradingError> {
        let student_exams = self.student_exams.lock().unwrap();
        let student_exam = student_exams
            .get(&student_exam_id)
            .ok_or(GradingError::StudentExamNotFound(student_exam_id))?;

        can_request_review(student_exam)?;

        let review = ReviewRequest {
            id: Uuid::new_v4(),
            student_exam_id,
            requested_by,
            reason,
            status: ReviewStatus::Pending,
            result: None,
        };

        let id = review.id;
        self.review_requests
            .lock()
            .unwrap()
            .insert(id, review);

        Ok(id)
    }

    pub fn process_review(&self, review_id: Uuid) -> Result<ReviewResult, GradingError> {
        let mut reviews = self.review_requests.lock().unwrap();
        let review = reviews
            .get_mut(&review_id)
            .ok_or(GradingError::ReviewRequestNotFound(review_id))?;

        review.status = ReviewStatus::InProgress;

        let student_exams = self.student_exams.lock().unwrap();
        let student_exam = student_exams
            .get(&review.student_exam_id)
            .ok_or(GradingError::StudentExamNotFound(review.student_exam_id))?;

        let exams = self.exams.lock().unwrap();
        let exam = exams
            .get(&student_exam.exam_id)
            .ok_or(GradingError::ExamNotFound(student_exam.exam_id))?;

        let result = review_for_errors(student_exam, exam);

        review.status = ReviewStatus::Completed;
        review.result = Some(result.clone());

        Ok(result)
    }

    pub fn get_student_exam(&self, id: Uuid) -> Result<StudentExam, GradingError> {
        self.student_exams
            .lock()
            .unwrap()
            .get(&id)
            .cloned()
            .ok_or(GradingError::StudentExamNotFound(id))
    }

    pub fn get_teacher(&self, id: Uuid) -> Result<Teacher, GradingError> {
        self.teachers
            .lock()
            .unwrap()
            .get(&id)
            .cloned()
            .ok_or(GradingError::TeacherNotFound(id))
    }

    pub fn get_student(&self, id: Uuid) -> Result<Student, GradingError> {
        self.students
            .lock()
            .unwrap()
            .get(&id)
            .cloned()
            .ok_or(GradingError::StudentNotFound(id))
    }
}

impl Default for InMemoryStorage {
    fn default() -> Self {
        Self::new()
    }
}
