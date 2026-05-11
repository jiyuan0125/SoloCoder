use crate::models::*;
use crate::errors::GradingError;
use std::collections::HashSet;
use uuid::Uuid;

pub const DEFAULT_SCORE_DIFF_THRESHOLD: u32 = 2;

pub fn grade_objective_question(
    question: &Question,
    answer: &Answer,
) -> Result<u32, GradingError> {
    match question.question_type {
        QuestionType::SingleChoice => grade_single_choice(question, answer),
        QuestionType::MultipleChoice => grade_multiple_choice(question, answer),
        QuestionType::Subjective => Err(GradingError::InvalidQuestionType),
    }
}

fn grade_single_choice(question: &Question, answer: &Answer) -> Result<u32, GradingError> {
    let correct = question.correct_answer.as_ref()
        .ok_or(GradingError::InvalidQuestionType)?;
    
    let selected = answer.selected_options.as_ref()
        .ok_or(GradingError::AnswerNotFound(question.id))?;
    
    if selected.len() == 1 && selected[0] == correct[0] {
        Ok(question.max_score)
    } else {
        Ok(0)
    }
}

fn grade_multiple_choice(question: &Question, answer: &Answer) -> Result<u32, GradingError> {
    let correct = question.correct_answer.as_ref()
        .ok_or(GradingError::InvalidQuestionType)?;
    let correct_set: HashSet<&str> = correct.iter().map(|s| s.as_str()).collect();
    
    let selected = answer.selected_options.as_ref()
        .ok_or(GradingError::AnswerNotFound(question.id))?;
    let selected_set: HashSet<&str> = selected.iter().map(|s| s.as_str()).collect();
    
    let has_wrong = !selected_set.is_subset(&correct_set);
    
    if has_wrong {
        Ok(0)
    } else if selected_set == correct_set {
        Ok(question.max_score)
    } else if selected_set.is_subset(&correct_set) && !selected_set.is_empty() {
        let half_score = question.max_score / 2;
        Ok(half_score)
    } else {
        Ok(0)
    }
}

pub fn grade_subjective_first(
    score_record: &mut SubjectiveScore,
    teacher_id: Uuid,
    score: u32,
    max_score: u32,
    comments: Option<String>,
    threshold: u32,
) -> Result<(), GradingError> {
    if score > max_score {
        return Err(GradingError::ScoreOutOfRange(score, max_score));
    }
    
    if score_record.first_score.is_some() {
        return Err(GradingError::InvalidStateTransition);
    }
    
    score_record.first_score = Some(TeacherScore {
        teacher_id,
        score,
        comments,
    });
    score_record.status = SubjectiveScoreStatus::PendingSecondReview;
    
    Ok(())
}

pub fn grade_subjective_second(
    score_record: &mut SubjectiveScore,
    teacher_id: Uuid,
    score: u32,
    max_score: u32,
    comments: Option<String>,
    threshold: u32,
) -> Result<(), GradingError> {
    if score > max_score {
        return Err(GradingError::ScoreOutOfRange(score, max_score));
    }
    
    if score_record.second_score.is_some() || score_record.first_score.is_none() {
        return Err(GradingError::InvalidStateTransition);
    }
    
    let first = score_record.first_score.as_ref().unwrap();
    score_record.second_score = Some(TeacherScore {
        teacher_id,
        score,
        comments,
    });
    
    let diff = (first.score as i32 - score as i32).abs() as u32;
    
    if diff <= threshold {
        let avg = ((first.score + score) as f64 / 2.0).round() as u32;
        score_record.final_score = Some(avg);
        score_record.status = SubjectiveScoreStatus::Completed;
    } else {
        score_record.status = SubjectiveScoreStatus::PendingThirdReview;
    }
    
    Ok(())
}

pub fn grade_subjective_third(
    score_record: &mut SubjectiveScore,
    teacher_id: Uuid,
    score: u32,
    max_score: u32,
    comments: Option<String>,
) -> Result<(), GradingError> {
    if score > max_score {
        return Err(GradingError::ScoreOutOfRange(score, max_score));
    }
    
    if score_record.third_score.is_some() 
        || score_record.first_score.is_none() 
        || score_record.second_score.is_none()
    {
        return Err(GradingError::InvalidStateTransition);
    }
    
    score_record.third_score = Some(TeacherScore {
        teacher_id,
        score,
        comments,
    });
    
    let s1 = score_record.first_score.as_ref().unwrap().score;
    let s2 = score_record.second_score.as_ref().unwrap().score;
    let s3 = score;
    
    let median = calculate_median(s1, s2, s3);
    score_record.final_score = Some(median);
    score_record.status = SubjectiveScoreStatus::Completed;
    
    Ok(())
}

fn calculate_median(a: u32, b: u32, c: u32) -> u32 {
    let mut scores = vec![a, b, c];
    scores.sort();
    scores[1]
}

pub fn calculate_total_score(
    student_exam: &StudentExam,
) -> Result<u32, GradingError> {
    let objective_total: u32 = student_exam.objective_scores.values().sum();
    
    let subjective_total: u32 = student_exam.subjective_scores
        .values()
        .filter_map(|s| s.final_score)
        .sum();
    
    Ok(objective_total + subjective_total)
}

pub fn is_teacher_eligible(
    teacher: &Teacher,
    question: &Question,
    student: &Student,
    already_assigned: &HashSet<Uuid>,
) -> bool {
    if teacher.id == question.author_id {
        return false;
    }
    
    if let Some(mentor_id) = student.mentor_id {
        if teacher.id == mentor_id {
            return false;
        }
    }
    
    if teacher.mentor_students.contains(&student.id) {
        return false;
    }
    
    if already_assigned.contains(&teacher.id) {
        return false;
    }
    
    true
}
