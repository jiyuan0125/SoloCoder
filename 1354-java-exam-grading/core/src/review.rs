use crate::models::*;
use crate::errors::GradingError;

pub fn review_for_errors(
    student_exam: &StudentExam,
    exam: &ExamPaper,
) -> ReviewResult {
    let original_total = student_exam.total_score.unwrap_or(0);
    
    let mut recalculated_objective: u32 = 0;
    let mut objective_missing = false;
    
    for question in &exam.questions {
        if question.question_type != QuestionType::Subjective {
            match student_exam.objective_scores.get(&question.id) {
                Some(score) => recalculated_objective += *score,
                None => objective_missing = true,
            }
        }
    }
    
    let mut recalculated_subjective: u32 = 0;
    let mut subjective_missing = false;
    
    for question in &exam.questions {
        if question.question_type == QuestionType::Subjective {
            match student_exam.subjective_scores.get(&question.id) {
                Some(score_record) => {
                    if let Some(final_score) = score_record.final_score {
                        recalculated_subjective += final_score;
                    } else {
                        subjective_missing = true;
                    }
                }
                None => subjective_missing = true,
            }
        }
    }
    
    let recalculated_total = recalculated_objective + recalculated_subjective;
    
    let mut error_descriptions = Vec::new();
    
    if objective_missing {
        error_descriptions.push("发现未评分的客观题".to_string());
    }
    
    if subjective_missing {
        error_descriptions.push("发现未完成评分的主观题".to_string());
    }
    
    if original_total != recalculated_total {
        error_descriptions.push(format!(
            "总分计算错误: 原始总分{}，重新计算总分{}",
            original_total, recalculated_total
        ));
    }
    
    let found_errors = !error_descriptions.is_empty();
    let error_description = if found_errors {
        error_descriptions.join("; ")
    } else {
        "未发现评分或计算错误".to_string()
    };
    
    ReviewResult {
        found_errors,
        error_description,
        original_total,
        corrected_total: recalculated_total,
    }
}

pub fn can_request_review(student_exam: &StudentExam) -> Result<(), GradingError> {
    if !student_exam.is_published {
        return Err(GradingError::ExamNotPublished);
    }
    Ok(())
}
