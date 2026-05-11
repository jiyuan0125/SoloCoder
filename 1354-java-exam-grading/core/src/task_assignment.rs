use crate::models::*;
use crate::errors::GradingError;
use crate::grading::is_teacher_eligible;
use std::collections::{HashMap, HashSet};

pub struct TaskAssignmentSystem {
    teacher_workload: HashMap<Uuid, u32>,
}

impl TaskAssignmentSystem {
    pub fn new() -> Self {
        Self {
            teacher_workload: HashMap::new(),
        }
    }

    pub fn from_teachers(teachers: &[Teacher]) -> Self {
        let mut workload = HashMap::new();
        for teacher in teachers {
            workload.insert(teacher.id, 0);
        }
        Self {
            teacher_workload: workload,
        }
    }

    pub fn get_workload(&self, teacher_id: Uuid) -> u32 {
        *self.teacher_workload.get(&teacher_id).unwrap_or(&0)
    }

    pub fn add_workload(&mut self, teacher_id: Uuid) {
        *self.teacher_workload.entry(teacher_id).or_insert(0) += 1;
    }

    pub fn select_eligible_teacher(
        &mut self,
        teachers: &[Teacher],
        question: &Question,
        student: &Student,
        already_assigned: &HashSet<Uuid>,
    ) -> Result<Uuid, GradingError> {
        let eligible: Vec<&Teacher> = teachers
            .iter()
            .filter(|t| is_teacher_eligible(t, question, student, already_assigned))
            .collect();

        if eligible.is_empty() {
            return Err(GradingError::NoEligibleTeachers);
        }

        let selected = eligible
            .iter()
            .min_by_key(|t| self.get_workload(t.id))
            .ok_or(GradingError::NoEligibleTeachers)?;

        self.add_workload(selected.id);
        Ok(selected.id)
    }

    pub fn select_two_eligible_teachers(
        &mut self,
        teachers: &[Teacher],
        question: &Question,
        student: &Student,
    ) -> Result<(Uuid, Uuid), GradingError> {
        let mut assigned = HashSet::new();

        let first = self.select_eligible_teacher(teachers, question, student, &assigned)?;
        assigned.insert(first);

        let second = self.select_eligible_teacher(teachers, question, student, &assigned)?;

        Ok((first, second))
    }

    pub fn select_third_eligible_teacher(
        &mut self,
        teachers: &[Teacher],
        question: &Question,
        student: &Student,
        first_teacher: Uuid,
        second_teacher: Uuid,
    ) -> Result<Uuid, GradingError> {
        let mut assigned = HashSet::new();
        assigned.insert(first_teacher);
        assigned.insert(second_teacher);

        self.select_eligible_teacher(teachers, question, student, &assigned)
    }

    pub fn assign_subjective_tasks(
        &mut self,
        teachers: &[Teacher],
        exam: &ExamPaper,
        student_exams: &[StudentExam],
        students: &HashMap<Uuid, Student>,
    ) -> Result<Vec<GradingTask>, GradingError> {
        let mut tasks = Vec::new();

        let subjective_questions: Vec<&Question> = exam
            .questions
            .iter()
            .filter(|q| q.question_type == QuestionType::Subjective)
            .collect();

        for student_exam in student_exams {
            let student = students
                .get(&student_exam.student_id)
                .ok_or(GradingError::StudentNotFound(student_exam.student_id))?;

            for question in &subjective_questions {
                let (first, second) = self.select_two_eligible_teachers(
                    teachers,
                    question,
                    student,
                )?;

                let first_task = GradingTask {
                    id: Uuid::new_v4(),
                    student_exam_id: student_exam.id,
                    question_id: question.id,
                    teacher_id: first,
                    status: TaskStatus::Assigned,
                };

                let second_task = GradingTask {
                    id: Uuid::new_v4(),
                    student_exam_id: student_exam.id,
                    question_id: question.id,
                    teacher_id: second,
                    status: TaskStatus::Assigned,
                };

                tasks.push(first_task);
                tasks.push(second_task);
            }
        }

        Ok(tasks)
    }
}

impl Default for TaskAssignmentSystem {
    fn default() -> Self {
        Self::new()
    }
}
