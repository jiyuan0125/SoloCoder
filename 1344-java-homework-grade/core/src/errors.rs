use thiserror::Error;

#[derive(Error, Debug)]
pub enum AppError {
    #[error("User not found: {0}")]
    UserNotFound(String),

    #[error("Course not found: {0}")]
    CourseNotFound(String),

    #[error("Assignment not found: {0}")]
    AssignmentNotFound(String),

    #[error("Question not found: {0}")]
    QuestionNotFound(String),

    #[error("Submission not found: {0}")]
    SubmissionNotFound(String),

    #[error("Grading not found: {0}")]
    GradingNotFound(String),

    #[error("Not a teacher")]
    NotTeacher,

    #[error("Not a student")]
    NotStudent,

    #[error("Not enrolled in course")]
    NotEnrolled,

    #[error("Not a course teacher")]
    NotCourseTeacher,

    #[error("Score exceeds max score {max}")]
    ScoreExceedsMax { max: u32 },

    #[error("Already submitted")]
    AlreadySubmitted,

    #[error("Already graded")]
    AlreadyGraded,

    #[error("Appeal already submitted")]
    AlreadyAppealed,

    #[error("Must be graded by different teacher")]
    MustBeDifferentTeacher,

    #[error("Cannot modify after submission")]
    CannotModifyAfterSubmission,

    #[error("Cannot modify after grading")]
    CannotModifyAfterGrading,

    #[error("Invalid role")]
    InvalidRole,

    #[error("Internal error: {0}")]
    Internal(String),
}
