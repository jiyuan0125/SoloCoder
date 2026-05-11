use serde::{Deserialize, Serialize};
use uuid::Uuid;
use std::collections::{HashMap, HashSet};

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum QuestionType {
    SingleChoice,
    MultipleChoice,
    Subjective,
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct Question {
    pub id: Uuid,
    pub question_type: QuestionType,
    pub content: String,
    pub options: Option<Vec<String>>,
    pub correct_answer: Option<Vec<String>>,
    pub max_score: u32,
    pub author_id: Uuid,
}

impl Question {
    pub fn new_single_choice(
        content: String,
        options: Vec<String>,
        correct_answer: String,
        max_score: u32,
        author_id: Uuid,
    ) -> Self {
        Self {
            id: Uuid::new_v4(),
            question_type: QuestionType::SingleChoice,
            content,
            options: Some(options),
            correct_answer: Some(vec![correct_answer]),
            max_score,
            author_id,
        }
    }

    pub fn new_multiple_choice(
        content: String,
        options: Vec<String>,
        correct_answers: Vec<String>,
        max_score: u32,
        author_id: Uuid,
    ) -> Self {
        Self {
            id: Uuid::new_v4(),
            question_type: QuestionType::MultipleChoice,
            content,
            options: Some(options),
            correct_answer: Some(correct_answers),
            max_score,
            author_id,
        }
    }

    pub fn new_subjective(
        content: String,
        max_score: u32,
        author_id: Uuid,
    ) -> Self {
        Self {
            id: Uuid::new_v4(),
            question_type: QuestionType::Subjective,
            content,
            options: None,
            correct_answer: None,
            max_score,
            author_id,
        }
    }
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct ExamPaper {
    pub id: Uuid,
    pub name: String,
    pub questions: Vec<Question>,
    pub total_score: u32,
}

impl ExamPaper {
    pub fn new(name: String, questions: Vec<Question>) -> Self {
        let total_score = questions.iter().map(|q| q.max_score).sum();
        Self {
            id: Uuid::new_v4(),
            name,
            questions,
            total_score,
        }
    }
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct Teacher {
    pub id: Uuid,
    pub name: String,
    pub mentor_students: HashSet<Uuid>,
}

impl Teacher {
    pub fn new(name: String) -> Self {
        Self {
            id: Uuid::new_v4(),
            name,
            mentor_students: HashSet::new(),
        }
    }

    pub fn with_students(name: String, students: Vec<Uuid>) -> Self {
        Self {
            id: Uuid::new_v4(),
            name,
            mentor_students: students.into_iter().collect(),
        }
    }
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct Student {
    pub id: Uuid,
    pub name: String,
    pub mentor_id: Option<Uuid>,
}

impl Student {
    pub fn new(name: String) -> Self {
        Self {
            id: Uuid::new_v4(),
            name,
            mentor_id: None,
        }
    }

    pub fn with_mentor(name: String, mentor_id: Uuid) -> Self {
        Self {
            id: Uuid::new_v4(),
            name,
            mentor_id: Some(mentor_id),
        }
    }
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct Answer {
    pub question_id: Uuid,
    pub content: Option<String>,
    pub selected_options: Option<Vec<String>>,
}

impl Answer {
    pub fn new_objective(question_id: Uuid, selected_options: Vec<String>) -> Self {
        Self {
            question_id,
            content: None,
            selected_options: Some(selected_options),
        }
    }

    pub fn new_subjective(question_id: Uuid, content: String) -> Self {
        Self {
            question_id,
            content: Some(content),
            selected_options: None,
        }
    }
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct StudentExam {
    pub id: Uuid,
    pub student_id: Uuid,
    pub exam_id: Uuid,
    pub answers: HashMap<Uuid, Answer>,
    pub objective_scores: HashMap<Uuid, u32>,
    pub subjective_scores: HashMap<Uuid, SubjectiveScore>,
    pub total_score: Option<u32>,
    pub status: ExamStatus,
    pub anomalies: Vec<Anomaly>,
    pub is_published: bool,
}

impl StudentExam {
    pub fn new(student_id: Uuid, exam_id: Uuid, answers: HashMap<Uuid, Answer>) -> Self {
        Self {
            id: Uuid::new_v4(),
            student_id,
            exam_id,
            answers,
            objective_scores: HashMap::new(),
            subjective_scores: HashMap::new(),
            total_score: None,
            status: ExamStatus::PendingObjectiveGrading,
            anomalies: Vec::new(),
            is_published: false,
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ExamStatus {
    PendingObjectiveGrading,
    PendingSubjectiveGrading,
    GradingCompleted,
    NeedsReview,
    Completed,
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct SubjectiveScore {
    pub question_id: Uuid,
    pub first_score: Option<TeacherScore>,
    pub second_score: Option<TeacherScore>,
    pub third_score: Option<TeacherScore>,
    pub final_score: Option<u32>,
    pub status: SubjectiveScoreStatus,
}

impl SubjectiveScore {
    pub fn new(question_id: Uuid) -> Self {
        Self {
            question_id,
            first_score: None,
            second_score: None,
            third_score: None,
            final_score: None,
            status: SubjectiveScoreStatus::PendingFirstReview,
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum SubjectiveScoreStatus {
    PendingFirstReview,
    PendingSecondReview,
    PendingThirdReview,
    Completed,
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct TeacherScore {
    pub teacher_id: Uuid,
    pub score: u32,
    pub comments: Option<String>,
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct GradingTask {
    pub id: Uuid,
    pub student_exam_id: Uuid,
    pub question_id: Uuid,
    pub teacher_id: Uuid,
    pub status: TaskStatus,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum TaskStatus {
    Assigned,
    Completed,
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct Anomaly {
    pub id: Uuid,
    pub anomaly_type: AnomalyType,
    pub reported_by: Uuid,
    pub description: String,
    pub reviewed: bool,
    pub related_student_exam_id: Option<Uuid>,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum AnomalyType {
    Plagiarism,
    Illegible,
    Other,
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct ReviewRequest {
    pub id: Uuid,
    pub student_exam_id: Uuid,
    pub requested_by: Uuid,
    pub reason: String,
    pub status: ReviewStatus,
    pub result: Option<ReviewResult>,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ReviewStatus {
    Pending,
    InProgress,
    Completed,
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct ReviewResult {
    pub found_errors: bool,
    pub error_description: String,
    pub original_total: u32,
    pub corrected_total: u32,
}
