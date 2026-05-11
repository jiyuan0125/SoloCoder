use thiserror::Error;

#[derive(Debug, Error)]
pub enum RecruitError {
    #[error("Candidate not found: {0}")]
    CandidateNotFound(String),

    #[error("Interview not found: {0}")]
    InterviewNotFound(String),

    #[error("Offer not found: {0}")]
    OfferNotFound(String),

    #[error("Invalid state transition: {0}")]
    InvalidStateTransition(String),

    #[error("Interview time conflict for interviewer: {0}")]
    InterviewerTimeConflict(String),

    #[error("Interview time conflict for candidate")]
    CandidateTimeConflict,

    #[error("No buffer time between interviews")]
    NoBufferTime,

    #[error("Offer requires approval from: {0}")]
    InsufficientApprovalLevel(String),

    #[error("Offer is not in pending approval state")]
    OfferNotPending,

    #[error("Candidate is in pending status, cannot proceed")]
    CandidatePending,

    #[error("Candidate has expired, HR confirmation required")]
    CandidateExpired,

    #[error("Invalid input: {0}")]
    InvalidInput(String),
}
