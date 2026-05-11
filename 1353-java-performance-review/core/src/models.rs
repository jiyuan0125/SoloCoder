use serde::{Deserialize, Serialize};
use std::collections::HashMap;

pub type EmployeeId = String;
pub type DepartmentId = String;
pub type ReviewCycleId = String;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum Grade {
    A,
    B,
    C,
    D,
}

impl Grade {
    pub fn coefficient(&self) -> f64 {
        match self {
            Grade::A => 1.3,
            Grade::B => 1.1,
            Grade::C => 1.0,
            Grade::D => 0.7,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Department {
    pub id: DepartmentId,
    pub name: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DepartmentAssignment {
    pub department_id: DepartmentId,
    pub start_date: String,
    pub end_date: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Employee {
    pub id: EmployeeId,
    pub name: String,
    pub manager_id: Option<EmployeeId>,
    pub base_performance: f64,
    pub department_assignments: Vec<DepartmentAssignment>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ReviewCycle {
    pub id: ReviewCycleId,
    pub name: String,
    pub start_date: String,
    pub end_date: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SelfAssessment {
    pub employee_id: EmployeeId,
    pub cycle_id: ReviewCycleId,
    pub score: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ManagerReview {
    pub employee_id: EmployeeId,
    pub cycle_id: ReviewCycleId,
    pub manager_id: EmployeeId,
    pub score: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PeerReview {
    pub reviewer_id: EmployeeId,
    pub reviewee_id: EmployeeId,
    pub cycle_id: ReviewCycleId,
    pub score: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct WeightedScores {
    pub self_score: f64,
    pub manager_score: f64,
    pub peer_score: f64,
    pub peer_count: usize,
    pub total_score: f64,
    pub weights_used: ScoreWeights,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ScoreWeights {
    pub self_weight: f64,
    pub manager_weight: f64,
    pub peer_weight: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DepartmentSegment {
    pub department_id: DepartmentId,
    pub days_in_cycle: u32,
    pub segment_score: f64,
    pub final_grade: Grade,
    pub segment_bonus: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EmployeePerformanceResult {
    pub employee_id: EmployeeId,
    pub cycle_id: ReviewCycleId,
    pub weighted_scores: WeightedScores,
    pub primary_department_id: DepartmentId,
    pub department_segments: Vec<DepartmentSegment>,
    pub final_grade: Grade,
    pub total_bonus: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DepartmentDistribution {
    pub department_id: DepartmentId,
    pub cycle_id: ReviewCycleId,
    pub total_employees: usize,
    pub grade_counts: HashMap<Grade, usize>,
    pub employee_results: Vec<EmployeePerformanceResult>,
}

#[derive(Debug, Clone, Default)]
pub struct PerformanceData {
    pub departments: HashMap<DepartmentId, Department>,
    pub employees: HashMap<EmployeeId, Employee>,
    pub review_cycles: HashMap<ReviewCycleId, ReviewCycle>,
    pub self_assessments: HashMap<(EmployeeId, ReviewCycleId), SelfAssessment>,
    pub manager_reviews: HashMap<(EmployeeId, ReviewCycleId), ManagerReview>,
    pub peer_reviews: HashMap<(EmployeeId, ReviewCycleId), Vec<PeerReview>>,
}

impl PerformanceData {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn add_department(&mut self, dept: Department) {
        self.departments.insert(dept.id.clone(), dept);
    }

    pub fn add_employee(&mut self, emp: Employee) {
        self.employees.insert(emp.id.clone(), emp);
    }

    pub fn add_review_cycle(&mut self, cycle: ReviewCycle) {
        self.review_cycles.insert(cycle.id.clone(), cycle);
    }

    pub fn add_self_assessment(&mut self, sa: SelfAssessment) {
        self.self_assessments
            .insert((sa.employee_id.clone(), sa.cycle_id.clone()), sa);
    }

    pub fn add_manager_review(&mut self, mr: ManagerReview) {
        self.manager_reviews
            .insert((mr.employee_id.clone(), mr.cycle_id.clone()), mr);
    }

    pub fn add_peer_review(&mut self, pr: PeerReview) {
        let key = (pr.reviewee_id.clone(), pr.cycle_id.clone());
        self.peer_reviews
            .entry(key)
            .or_default()
            .push(pr);
    }
}
