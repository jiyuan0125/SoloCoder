use parking_lot::RwLock;
use std::collections::HashMap;
use std::sync::Arc;
use uuid::Uuid;

use crate::models::*;

pub type SharedStore = Arc<InMemoryStore>;

pub struct InMemoryStore {
    students: RwLock<HashMap<Uuid, Student>>,
    courses: RwLock<HashMap<Uuid, Course>>,
    tuition_records: RwLock<HashMap<(Uuid, Semester), TuitionRecord>>,
    enrollments: RwLock<HashMap<(Uuid, Uuid, Semester), Enrollment>>,
    enrollment_periods: RwLock<HashMap<Semester, EnrollmentPeriod>>,
}

impl InMemoryStore {
    pub fn new() -> Self {
        Self {
            students: RwLock::new(HashMap::new()),
            courses: RwLock::new(HashMap::new()),
            tuition_records: RwLock::new(HashMap::new()),
            enrollments: RwLock::new(HashMap::new()),
            enrollment_periods: RwLock::new(HashMap::new()),
        }
    }

    pub fn shared() -> SharedStore {
        Arc::new(Self::new())
    }
}

impl InMemoryStore {
    pub fn add_student(&self, student: Student) {
        self.students.write().insert(student.id, student);
    }

    pub fn get_student(&self, id: Uuid) -> Option<Student> {
        self.students.read().get(&id).cloned()
    }

    pub fn list_students(&self) -> Vec<Student> {
        self.students.read().values().cloned().collect()
    }

    pub fn add_course(&self, course: Course) {
        self.courses.write().insert(course.id, course);
    }

    pub fn get_course(&self, id: Uuid) -> Option<Course> {
        self.courses.read().get(&id).cloned()
    }

    pub fn list_courses(&self, semester: &Semester) -> Vec<Course> {
        self.courses
            .read()
            .values()
            .filter(|c| c.semester == *semester)
            .cloned()
            .collect()
    }

    pub fn set_tuition_record(&self, record: TuitionRecord) {
        let key = (record.student_id, record.semester.clone());
        self.tuition_records.write().insert(key, record);
    }

    pub fn get_tuition_record(&self, student_id: Uuid, semester: &Semester) -> Option<TuitionRecord> {
        self.tuition_records
            .read()
            .get(&(student_id, semester.clone()))
            .cloned()
    }

    pub fn update_tuition_record<F>(&self, student_id: Uuid, semester: &Semester, f: F) -> bool
    where
        F: FnOnce(&mut TuitionRecord),
    {
        let mut records = self.tuition_records.write();
        if let Some(record) = records.get_mut(&(student_id, semester.clone())) {
            f(record);
            true
        } else {
            false
        }
    }

    pub fn list_tuition_records(&self) -> Vec<TuitionRecord> {
        self.tuition_records.read().values().cloned().collect()
    }

    pub fn add_enrollment(&self, enrollment: Enrollment) -> bool {
        let key = (enrollment.course_id, enrollment.student_id, enrollment.semester.clone());
        let mut enrollments = self.enrollments.write();
        if enrollments.contains_key(&key) {
            false
        } else {
            enrollments.insert(key, enrollment);
            true
        }
    }

    pub fn remove_enrollment(&self, course_id: Uuid, student_id: Uuid, semester: &Semester) -> bool {
        self.enrollments
            .write()
            .remove(&(course_id, student_id, semester.clone()))
            .is_some()
    }

    pub fn is_enrolled(&self, course_id: Uuid, student_id: Uuid, semester: &Semester) -> bool {
        self.enrollments
            .read()
            .contains_key(&(course_id, student_id, semester.clone()))
    }

    pub fn get_enrollment_count(&self, course_id: Uuid, semester: &Semester) -> u32 {
        self.enrollments
            .read()
            .iter()
            .filter(|(k, _)| k.0 == course_id && &k.2 == semester)
            .count() as u32
    }

    pub fn list_student_enrollments(&self, student_id: Uuid, semester: &Semester) -> Vec<Uuid> {
        self.enrollments
            .read()
            .iter()
            .filter(|(k, _)| k.1 == student_id && &k.2 == semester)
            .map(|(k, _)| k.0)
            .collect()
    }

    pub fn set_enrollment_period(&self, period: EnrollmentPeriod) {
        self.enrollment_periods
            .write()
            .insert(period.semester.clone(), period);
    }

    pub fn get_enrollment_period(&self, semester: &Semester) -> Option<EnrollmentPeriod> {
        self.enrollment_periods.read().get(semester).cloned()
    }
}
