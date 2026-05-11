use thiserror::Error;

#[derive(Error, Debug)]
pub enum TuitionError {
    #[error("Student not found")]
    StudentNotFound,
    
    #[error("Course not found")]
    CourseNotFound,
    
    #[error("Tuition record not found for semester {0}")]
    TuitionRecordNotFound(String),
    
    #[error("Enrollment period not set for semester {0}")]
    EnrollmentPeriodNotSet(String),
    
    #[error("Enrollment period is not active for semester {0}")]
    EnrollmentPeriodNotActive(String),
    
    #[error("Course {0} is full")]
    CourseFull(String),
    
    #[error("Already enrolled in course {0}")]
    AlreadyEnrolled(String),
    
    #[error("Not enrolled in course {0}")]
    NotEnrolled(String),
    
    #[error("Student has unpaid tuition for semester {0}")]
    UnpaidTuition(String),
    
    #[error("Payment amount {0} is invalid: cannot be negative or exceed owed amount")]
    InvalidPaymentAmount(f64),
    
    #[error("First installment already paid")]
    FirstInstallmentAlreadyPaid,
    
    #[error("Second installment already paid")]
    SecondInstallmentAlreadyPaid,
    
    #[error("Must pay first installment before second")]
    MustPayFirstInstallmentFirst,
    
    #[error("Internal error: {0}")]
    Internal(String),
}

impl From<TuitionError> for String {
    fn from(err: TuitionError) -> String {
        err.to_string()
    }
}
