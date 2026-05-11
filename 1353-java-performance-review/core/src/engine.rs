use crate::models::*;
use std::collections::HashMap;

const MIN_PEERS: usize = 3;
const SELF_WEIGHT: f64 = 0.20;
const MANAGER_WEIGHT: f64 = 0.50;
const PEER_WEIGHT: f64 = 0.30;

const A_RATIO: f64 = 0.15;
const B_RATIO: f64 = 0.35;
const C_RATIO: f64 = 0.40;
const D_RATIO: f64 = 0.10;

const FORCED_DISTRIBUTION_MIN_SIZE: usize = 10;

pub fn calculate_weighted_scores(
    data: &PerformanceData,
    employee_id: &EmployeeId,
    cycle_id: &ReviewCycleId,
) -> Option<WeightedScores> {
    let key = (employee_id.clone(), cycle_id.clone());

    let self_score = data.self_assessments.get(&key).map(|sa| sa.score);

    let manager_score = data.manager_reviews.get(&key).map(|mr| mr.score);

    let peer_reviews = data.peer_reviews.get(&key);

    let peer_count = peer_reviews.map(|prs| prs.len()).unwrap_or(0);
    let peer_avg = peer_reviews.map(|prs| {
        let sum: f64 = prs.iter().map(|pr| pr.score).sum();
        if prs.is_empty() { 0.0 } else { sum / prs.len() as f64 }
    });

    let mut weights = ScoreWeights {
        self_weight: SELF_WEIGHT,
        manager_weight: MANAGER_WEIGHT,
        peer_weight: if peer_count >= MIN_PEERS { PEER_WEIGHT } else { 0.0 },
    };

    if weights.peer_weight == 0.0 && peer_count < MIN_PEERS {
        weights.manager_weight += PEER_WEIGHT;
    }

    let self_score = self_score.unwrap_or(0.0);
    let manager_score = manager_score.unwrap_or(0.0);
    let peer_score = peer_avg.unwrap_or(0.0);

    let total_score = self_score * weights.self_weight
        + manager_score * weights.manager_weight
        + peer_score * weights.peer_weight;

    Some(WeightedScores {
        self_score,
        manager_score,
        peer_score,
        peer_count,
        total_score,
        weights_used: weights,
    })
}

pub fn parse_date(date_str: &str) -> Option<chrono::NaiveDate> {
    chrono::NaiveDate::parse_from_str(date_str, "%Y-%m-%d").ok()
}

pub fn count_days(start: chrono::NaiveDate, end: chrono::NaiveDate) -> i64 {
    (end - start).num_days() + 1
}

pub fn get_employee_department_segments(
    employee: &Employee,
    cycle_start: chrono::NaiveDate,
    cycle_end: chrono::NaiveDate,
) -> Vec<(DepartmentId, u32)> {
    let mut segments = Vec::new();

    for assignment in &employee.department_assignments {
        let assign_start = match parse_date(&assignment.start_date) {
            Some(d) => d,
            None => continue,
        };

        let assign_end = match &assignment.end_date {
            Some(end_str) => match parse_date(end_str) {
                Some(d) => d,
                None => continue,
            },
            None => cycle_end,
        };

        let overlap_start = std::cmp::max(assign_start, cycle_start);
        let overlap_end = std::cmp::min(assign_end, cycle_end);

        if overlap_start <= overlap_end {
            let days = count_days(overlap_start, overlap_end) as u32;
            segments.push((assignment.department_id.clone(), days));
        }
    }

    segments.sort_by(|a, b| b.1.cmp(&a.1));
    segments
}

pub fn group_employees_by_primary_department(
    data: &PerformanceData,
    cycle_id: &ReviewCycleId,
) -> HashMap<DepartmentId, Vec<EmployeeId>> {
    let cycle = match data.review_cycles.get(cycle_id) {
        Some(c) => c,
        None => return HashMap::new(),
    };

    let cycle_start = match parse_date(&cycle.start_date) {
        Some(d) => d,
        None => return HashMap::new(),
    };
    let cycle_end = match parse_date(&cycle.end_date) {
        Some(d) => d,
        None => return HashMap::new(),
    };

    let mut dept_map: HashMap<DepartmentId, Vec<EmployeeId>> = HashMap::new();

    for (emp_id, employee) in &data.employees {
        let segments = get_employee_department_segments(employee, cycle_start, cycle_end);
        if let Some((primary_dept, _)) = segments.first() {
            dept_map
                .entry(primary_dept.clone())
                .or_default()
                .push(emp_id.clone());
        }
    }

    dept_map
}

pub fn calculate_forced_distribution(total: usize) -> HashMap<Grade, usize> {
    if total < FORCED_DISTRIBUTION_MIN_SIZE {
        return HashMap::new();
    }

    let a_count = (total as f64 * A_RATIO).round() as usize;
    let b_count = (total as f64 * B_RATIO).round() as usize;
    let c_count = (total as f64 * C_RATIO).round() as usize;
    let d_count = (total as f64 * D_RATIO).round() as usize;

    let mut counts = HashMap::new();
    counts.insert(Grade::A, a_count);
    counts.insert(Grade::B, b_count);
    counts.insert(Grade::C, c_count);
    counts.insert(Grade::D, d_count);

    let sum: usize = counts.values().sum();
    let diff = total as isize - sum as isize;

    if diff != 0 {
        let current_b = *counts.get(&Grade::B).unwrap_or(&0);
        let new_b = (current_b as isize + diff) as usize;
        counts.insert(Grade::B, new_b);
    }

    counts
}

pub fn assign_grades_for_department(
    data: &PerformanceData,
    _dept_id: &DepartmentId,
    cycle_id: &ReviewCycleId,
    employees: &[EmployeeId],
) -> Vec<(EmployeeId, Grade)> {
    let total = employees.len();

    let mut scored: Vec<_> = employees
        .iter()
        .filter_map(|emp_id| {
            let scores = calculate_weighted_scores(data, emp_id, cycle_id)?;
            Some((emp_id.clone(), scores.total_score))
        })
        .collect();

    scored.sort_by(|a, b| b.1.partial_cmp(&a.1).unwrap_or(std::cmp::Ordering::Equal));

    let distribution = calculate_forced_distribution(total);

    if distribution.is_empty() {
        return scored
            .into_iter()
            .map(|(emp_id, _)| (emp_id, Grade::C))
            .collect();
    }

    let a_count = *distribution.get(&Grade::A).unwrap_or(&0);
    let b_count = *distribution.get(&Grade::B).unwrap_or(&0);
    let c_count = *distribution.get(&Grade::C).unwrap_or(&0);
    let d_count = *distribution.get(&Grade::D).unwrap_or(&0);

    let mut result = Vec::new();
    let mut idx = 0usize;

    for _ in 0..a_count {
        if let Some((emp_id, _)) = scored.get(idx) {
            result.push((emp_id.clone(), Grade::A));
            idx += 1;
        }
    }

    for _ in 0..b_count {
        if let Some((emp_id, _)) = scored.get(idx) {
            result.push((emp_id.clone(), Grade::B));
            idx += 1;
        }
    }

    for _ in 0..c_count {
        if let Some((emp_id, _)) = scored.get(idx) {
            result.push((emp_id.clone(), Grade::C));
            idx += 1;
        }
    }

    for _ in 0..d_count {
        if let Some((emp_id, _)) = scored.get(idx) {
            result.push((emp_id.clone(), Grade::D));
            idx += 1;
        }
    }

    while idx < scored.len() {
        if let Some((emp_id, _)) = scored.get(idx) {
            result.push((emp_id.clone(), Grade::C));
        }
        idx += 1;
    }

    result
}

pub fn calculate_employee_result(
    data: &PerformanceData,
    emp_id: &EmployeeId,
    cycle_id: &ReviewCycleId,
    assigned_grade: Grade,
) -> Option<EmployeePerformanceResult> {
    let employee = data.employees.get(emp_id)?;
    let cycle = data.review_cycles.get(cycle_id)?;

    let weighted_scores = calculate_weighted_scores(data, emp_id, cycle_id)?;

    let cycle_start = parse_date(&cycle.start_date)?;
    let cycle_end = parse_date(&cycle.end_date)?;
    let cycle_total_days = count_days(cycle_start, cycle_end) as u32;

    let segments = get_employee_department_segments(employee, cycle_start, cycle_end);
    let primary_dept_id = segments.first().map(|(d, _)| d.clone())?;

    let department_segments: Vec<_> = segments
        .iter()
        .map(|(dept_id, days)| {
            let proportion = *days as f64 / cycle_total_days as f64;
            let segment_score = weighted_scores.total_score;
            let segment_bonus = employee.base_performance * assigned_grade.coefficient() * proportion;
            DepartmentSegment {
                department_id: dept_id.clone(),
                days_in_cycle: *days,
                segment_score,
                final_grade: assigned_grade,
                segment_bonus,
            }
        })
        .collect();

    let total_bonus = employee.base_performance * assigned_grade.coefficient();

    Some(EmployeePerformanceResult {
        employee_id: emp_id.clone(),
        cycle_id: cycle_id.clone(),
        weighted_scores,
        primary_department_id: primary_dept_id,
        department_segments,
        final_grade: assigned_grade,
        total_bonus,
    })
}

pub fn calculate_cycle_results(
    data: &PerformanceData,
    cycle_id: &ReviewCycleId,
) -> Vec<DepartmentDistribution> {
    let dept_groups = group_employees_by_primary_department(data, cycle_id);

    let mut results = Vec::new();

    for (dept_id, emp_ids) in dept_groups {
        let grade_assignments = assign_grades_for_department(data, &dept_id, cycle_id, &emp_ids);

        let mut employee_results = Vec::new();
        let mut grade_counts: HashMap<Grade, usize> = HashMap::new();

        for (emp_id, grade) in &grade_assignments {
            *grade_counts.entry(*grade).or_default() += 1;
            if let Some(result) = calculate_employee_result(data, &emp_id, cycle_id, *grade) {
                employee_results.push(result);
            }
        }

        results.push(DepartmentDistribution {
            department_id: dept_id,
            cycle_id: cycle_id.clone(),
            total_employees: emp_ids.len(),
            grade_counts,
            employee_results,
        });
    }

    results
}

#[cfg(test)]
mod tests {
    use super::*;

    fn create_test_data() -> PerformanceData {
        let mut data = PerformanceData::new();

        data.add_department(Department {
            id: "dept1".to_string(),
            name: "技术部".to_string(),
        });

        data.add_review_cycle(ReviewCycle {
            id: "cycle1".to_string(),
            name: "Q1-2024".to_string(),
            start_date: "2024-01-01".to_string(),
            end_date: "2024-03-31".to_string(),
        });

        for i in 1..=10 {
            data.add_employee(Employee {
                id: format!("emp{}", i),
                name: format!("员工{}", i),
                manager_id: Some("mgr1".to_string()),
                base_performance: 10000.0,
                department_assignments: vec![DepartmentAssignment {
                    department_id: "dept1".to_string(),
                    start_date: "2024-01-01".to_string(),
                    end_date: None,
                }],
            });
        }

        for i in 1..=10 {
            data.add_self_assessment(SelfAssessment {
                employee_id: format!("emp{}", i),
                cycle_id: "cycle1".to_string(),
                score: 85.0,
            });

            data.add_manager_review(ManagerReview {
                employee_id: format!("emp{}", i),
                cycle_id: "cycle1".to_string(),
                manager_id: "mgr1".to_string(),
                score: 80.0 + i as f64,
            });

            for j in 1..=3 {
                data.add_peer_review(PeerReview {
                    reviewer_id: format!("emp{}", ((i + j) % 10 + 1)),
                    reviewee_id: format!("emp{}", i),
                    cycle_id: "cycle1".to_string(),
                    score: 75.0 + i as f64,
                });
            }
        }

        data
    }

    #[test]
    fn test_forced_distribution_10_people() {
        let distribution = calculate_forced_distribution(10);
        assert_eq!(*distribution.get(&Grade::A).unwrap(), 2);
        assert_eq!(*distribution.get(&Grade::B).unwrap(), 3);
        assert_eq!(*distribution.get(&Grade::C).unwrap(), 4);
        assert_eq!(*distribution.get(&Grade::D).unwrap(), 1);
    }

    #[test]
    fn test_small_department_no_forced_distribution() {
        let distribution = calculate_forced_distribution(9);
        assert!(distribution.is_empty());
    }

    #[test]
    fn test_weighted_scores_with_enough_peers() {
        let data = create_test_data();
        let scores = calculate_weighted_scores(&data, &"emp1".to_string(), &"cycle1".to_string()).unwrap();

        assert_eq!(scores.weights_used.self_weight, 0.20);
        assert_eq!(scores.weights_used.manager_weight, 0.50);
        assert_eq!(scores.weights_used.peer_weight, 0.30);
        assert_eq!(scores.peer_count, 3);
    }

    #[test]
    fn test_cycle_results() {
        let data = create_test_data();
        let results = calculate_cycle_results(&data, &"cycle1".to_string());

        assert_eq!(results.len(), 1);
        let dept = &results[0];
        assert_eq!(dept.total_employees, 10);
        assert_eq!(dept.grade_counts.get(&Grade::A).unwrap(), &2);
        assert_eq!(dept.grade_counts.get(&Grade::B).unwrap(), &3);
        assert_eq!(dept.grade_counts.get(&Grade::C).unwrap(), &4);
        assert_eq!(dept.grade_counts.get(&Grade::D).unwrap(), &1);
    }
}
