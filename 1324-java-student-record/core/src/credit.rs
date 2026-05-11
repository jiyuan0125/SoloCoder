use crate::models::*;
use crate::errors::*;

pub fn calculate_current_credits(
    student: &Student,
    current_major: &Major,
) -> CreditSummary {
    let mut public = 0.0;
    let mut professional = 0.0;
    let mut elective = 0.0;

    for course in &student.completed_courses {
        match course.original_course_type {
            CourseType::Public => {
                public += course.credit;
            }
            CourseType::Professional => {
                if current_major.curriculum.contains_key(&course.course_id) {
                    professional += course.credit;
                } else {
                    elective += course.credit;
                }
            }
            CourseType::Elective => {
                elective += course.credit;
            }
        }
    }

    CreditSummary {
        public,
        professional,
        elective,
    }
}

pub fn calculate_transfer_credits(
    student: &Student,
    _from_major: &Major,
    to_major: &Major,
) -> (CreditSummary, f64, f64) {
    let mut public = 0.0;
    let mut professional = 0.0;
    let mut converted_to_elective = 0.0;

    for course in &student.completed_courses {
        match course.original_course_type {
            CourseType::Public => {
                public += course.credit;
            }
            CourseType::Professional => {
                if to_major.curriculum.contains_key(&course.course_id) {
                    professional += course.credit;
                } else {
                    converted_to_elective += course.credit;
                }
            }
            CourseType::Elective => {
                converted_to_elective += course.credit;
            }
        }
    }

    let max_allowed_elective = to_major.requirements.elective * 0.5;
    let excess_elective = if converted_to_elective > max_allowed_elective {
        converted_to_elective - max_allowed_elective
    } else {
        0.0
    };
    
    let elective = converted_to_elective - excess_elective;

    (
        CreditSummary {
            public,
            professional,
            elective,
        },
        converted_to_elective,
        excess_elective,
    )
}

pub fn calculate_graduation_gap(
    credits: &CreditSummary,
    requirements: &CreditRequirements,
) -> GraduationGap {
    let public_gap = (requirements.public - credits.public).max(0.0);
    let professional_gap = (requirements.professional - credits.professional).max(0.0);
    let elective_gap = (requirements.elective - credits.elective).max(0.0);
    let total = public_gap + professional_gap + elective_gap;

    GraduationGap {
        public: public_gap,
        professional: professional_gap,
        elective: elective_gap,
        total,
    }
}

pub fn process_transfer(
    student: &Student,
    from_major: &Major,
    to_major: &Major,
) -> Result<TransferResult> {
    if student.current_major_id == to_major.id {
        return Err(StudentRecordError::AlreadyInTargetMajor);
    }

    let (transferred_credits, converted_to_elective, excess_elective) = 
        calculate_transfer_credits(student, from_major, to_major);

    let graduation_gap = calculate_graduation_gap(
        &transferred_credits,
        &to_major.requirements,
    );

    let message = if excess_elective > 0.0 {
        format!(
            "转专业成功，公共课互认 {} 学分，专业课认定 {} 学分，选修课认定 {} 学分（其中 {} 学分超出50%上限，不予认定）",
            transferred_credits.public,
            transferred_credits.professional,
            transferred_credits.elective,
            excess_elective
        )
    } else {
        format!(
            "转专业成功，公共课互认 {} 学分，专业课认定 {} 学分，选修课认定 {} 学分",
            transferred_credits.public,
            transferred_credits.professional,
            transferred_credits.elective
        )
    };

    Ok(TransferResult {
        student_id: student.id.clone(),
        from_major_id: from_major.id.clone(),
        to_major_id: to_major.id.clone(),
        transferred_credits,
        converted_to_elective,
        excess_elective,
        graduation_gap,
        success: true,
        message,
    })
}
