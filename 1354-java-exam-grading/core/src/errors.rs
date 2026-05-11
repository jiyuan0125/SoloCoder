use thiserror::Error;
use uuid::Uuid;

#[derive(Debug, Error)]
pub enum GradingError {
    #[error("Exam not found: {0}")]
    ExamNotFound(Uuid),
    
    #[error("Student exam not found: {0}")]
    StudentExamNotFound(Uuid),
    
    #[error("Question not found: {0}")]
    QuestionNotFound(Uuid),
    
    #[error("Teacher not found: {0}")]
    TeacherNotFound(Uuid),
    
    #[error("Student not found: {0}")]
    StudentNotFound(Uuid),
    
    #[error("Task not found: {0}")]
    TaskNotFound(Uuid),
    
    #[error("Review request not found: {0}")]
    ReviewRequestNotFound(Uuid),
    
    #[error("Answer not found for question: {0}")]
    AnswerNotFound(Uuid),
    
    #[error("Invalid question type for operation")]
    InvalidQuestionType,
    
    #[error("No eligible teachers available for assignment")]
    NoEligibleTeachers,
    
    #[error("Score out of range: {0} (max: {1})")]
    ScoreOutOfRange(u32, u32),
    
    #[error("Task already completed")]
    TaskAlreadyCompleted,
    
    #[error("Exam already published")]
    ExamAlreadyPublished,
    
    #[error("Exam not published yet")]
    ExamNotPublished,
    
    #[error("Invalid state transition")]
    InvalidStateTransition,
    
    #[error("Plagiarism detected, both students get 0 points")]
    PlagiarismDetected,
}
