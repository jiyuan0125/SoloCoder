use serde::{Serialize, Deserialize};
use chrono::{DateTime, Utc};
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq, Hash)]
pub struct Semester(pub String);

impl Semester {
    pub fn new(year: u32, season: &str) -> Self {
        Semester(format!("{}{}", year, season))
    }
}

impl std::fmt::Display for Semester {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.0)
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Student {
    pub id: Uuid,
    pub name: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TuitionRecord {
    pub student_id: Uuid,
    pub semester: Semester,
    pub total_amount: f64,
    pub first_installment: f64,
    pub second_installment: f64,
    pub is_paid: bool,
}

impl TuitionRecord {
    pub fn new(student_id: Uuid, semester: Semester, total_amount: f64) -> Self {
        Self {
            student_id,
            semester,
            total_amount,
            first_installment: 0.0,
            second_installment: 0.0,
            is_paid: false,
        }
    }

    pub fn paid_amount(&self) -> f64 {
        self.first_installment + self.second_installment
    }

    pub fn owed_amount(&self) -> f64 {
        self.total_amount - self.paid_amount()
    }

    pub fn update_paid_status(&mut self) {
        self.is_paid = self.first_installment > 0.0 && self.second_installment > 0.0 
            && (self.paid_amount() >= self.total_amount - 0.01);
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Course {
    pub id: Uuid,
    pub name: String,
    pub capacity: u32,
    pub semester: Semester,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Enrollment {
    pub course_id: Uuid,
    pub student_id: Uuid,
    pub semester: Semester,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EnrollmentPeriod {
    pub semester: Semester,
    pub start_time: DateTime<Utc>,
    pub end_time: DateTime<Utc>,
}

impl EnrollmentPeriod {
    pub fn is_active(&self) -> bool {
        let now = Utc::now();
        now >= self.start_time && now <= self.end_time
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CourseWithEnrollment {
    pub course: Course,
    pub enrolled_count: u32,
    pub is_full: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StudentOwingInfo {
    pub student: Student,
    pub semester: Semester,
    pub owed_amount: f64,
}
