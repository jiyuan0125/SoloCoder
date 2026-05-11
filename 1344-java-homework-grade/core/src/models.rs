use serde::{Deserialize, Serialize};
use std::collections::HashMap;

pub type UserId = String;
pub type CourseId = String;
pub type AssignmentId = String;
pub type QuestionId = String;
pub type SubmissionId = String;
pub type GradingId = String;

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum Role {
    Student,
    Teacher,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct User {
    pub id: UserId,
    pub name: String,
    pub role: Role,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Course {
    pub id: CourseId,
    pub name: String,
    pub teacher_ids: Vec<UserId>,
    pub student_ids: Vec<UserId>,
    pub assignment_ids: Vec<AssignmentId>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum QuestionType {
    Objective,
    Subjective,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Question {
    pub id: QuestionId,
    pub question_type: QuestionType,
    pub title: String,
    pub max_score: u32,
    pub answer: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Assignment {
    pub id: AssignmentId,
    pub course_id: CourseId,
    pub title: String,
    pub questions: Vec<Question>,
    pub created_at: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Answer {
    pub question_id: QuestionId,
    pub answer: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Submission {
    pub id: SubmissionId,
    pub assignment_id: AssignmentId,
    pub student_id: UserId,
    pub answers: Vec<Answer>,
    pub submitted_at: u64,
    pub objective_scores: HashMap<QuestionId, u32>,
    pub is_completed: bool,
    pub grading_ids: Vec<GradingId>,
    pub has_appealed: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum GradingStatus {
    Initial,
    Appeal,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Grading {
    pub id: GradingId,
    pub submission_id: SubmissionId,
    pub teacher_id: UserId,
    pub status: GradingStatus,
    pub subjective_scores: HashMap<QuestionId, u32>,
    pub graded_at: Option<u64>,
    pub is_final: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CourseStatistics {
    pub course_id: CourseId,
    pub course_name: String,
    pub avg_score: f64,
    pub max_score: u32,
    pub min_score: u32,
    pub pass_rate: f64,
    pub total_submissions: usize,
    pub pass_submissions: usize,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ScoreDistribution {
    pub range_0_59: usize,
    pub range_60_69: usize,
    pub range_70_79: usize,
    pub range_80_89: usize,
    pub range_90_100: usize,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StudentTrend {
    pub student_id: UserId,
    pub student_name: String,
    pub assignments: Vec<AssignmentTrend>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AssignmentTrend {
    pub assignment_id: AssignmentId,
    pub assignment_title: String,
    pub total_score: u32,
    pub percentage: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TeacherWorkload {
    pub teacher_id: UserId,
    pub teacher_name: String,
    pub total_gradings: usize,
    pub avg_score: f64,
}

impl Submission {
    pub fn get_final_grading<'a>(&self, gradings: &'a [Grading]) -> Option<&'a Grading> {
        gradings.iter().find(|g| g.is_final)
    }

    pub fn get_objective_total(&self) -> u32 {
        self.objective_scores.values().sum()
    }

    pub fn get_final_score(&self, gradings: &[Grading]) -> Option<u32> {
        let final_grading = self.get_final_grading(gradings)?;
        let subjective_total: u32 = final_grading.subjective_scores.values().sum();
        Some(self.get_objective_total() + subjective_total)
    }

    pub fn get_max_score(&self, assignment: &Assignment) -> u32 {
        assignment.questions.iter().map(|q| q.max_score).sum()
    }
}
