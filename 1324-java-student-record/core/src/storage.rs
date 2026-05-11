use crate::models::*;
use crate::errors::*;
use crate::credit::*;
use std::collections::{HashMap, HashSet};
use std::sync::Arc;
use tokio::sync::{RwLock, Mutex};

pub struct InMemoryStorage {
    students: RwLock<HashMap<String, Student>>,
    majors: RwLock<HashMap<String, Major>>,
    transfer_locks: Mutex<HashSet<String>>,
}

impl InMemoryStorage {
    pub fn new() -> Arc<Self> {
        Arc::new(Self {
            students: RwLock::new(HashMap::new()),
            majors: RwLock::new(HashMap::new()),
            transfer_locks: Mutex::new(HashSet::new()),
        })
    }

    pub async fn add_major(&self, major: Major) -> Result<()> {
        let mut majors = self.majors.write().await;
        majors.insert(major.id.clone(), major);
        Ok(())
    }

    pub async fn get_major(&self, major_id: &str) -> Result<Major> {
        let majors = self.majors.read().await;
        majors
            .get(major_id)
            .cloned()
            .ok_or_else(|| StudentRecordError::MajorNotFound(major_id.to_string()))
    }

    pub async fn list_majors(&self) -> Vec<Major> {
        let majors = self.majors.read().await;
        majors.values().cloned().collect()
    }

    pub async fn add_student(&self, student: Student) -> Result<()> {
        let majors = self.majors.read().await;
        if !majors.contains_key(&student.current_major_id) {
            return Err(StudentRecordError::MajorNotFound(student.current_major_id.clone()));
        }
        drop(majors);
        
        let mut students = self.students.write().await;
        students.insert(student.id.clone(), student);
        Ok(())
    }

    pub async fn get_student(&self, student_id: &str) -> Result<Student> {
        let students = self.students.read().await;
        students
            .get(student_id)
            .cloned()
            .ok_or_else(|| StudentRecordError::StudentNotFound(student_id.to_string()))
    }

    pub async fn list_students(&self) -> Vec<Student> {
        let students = self.students.read().await;
        students.values().cloned().collect()
    }

    pub async fn get_student_credit_summary(&self, student_id: &str) -> Result<CreditSummary> {
        let student = self.get_student(student_id).await?;
        let major = self.get_major(&student.current_major_id).await?;
        Ok(calculate_current_credits(&student, &major))
    }

    pub async fn get_student_graduation_gap(&self, student_id: &str) -> Result<GraduationGap> {
        let student = self.get_student(student_id).await?;
        let major = self.get_major(&student.current_major_id).await?;
        let credits = calculate_current_credits(&student, &major);
        Ok(calculate_graduation_gap(&credits, &major.requirements))
    }

    pub async fn process_transfer_request(
        self: Arc<Self>,
        request: TransferRequest,
    ) -> Result<TransferResult> {
        let student_id = request.student_id.clone();
        let target_major_id = request.target_major_id.clone();
        
        let transfer_locks = self.transfer_locks.lock().await;
        
        if transfer_locks.contains(&student_id) {
            return Err(StudentRecordError::TransferInProgress);
        }
        
        drop(transfer_locks);
        
        let mut transfer_locks = self.transfer_locks.lock().await;
        transfer_locks.insert(student_id.clone());
        drop(transfer_locks);

        let storage_clone = Arc::clone(&self);
        let student_id_clone = student_id.clone();

        let result = async move {
            let student = storage_clone.get_student(&student_id_clone).await?;
            let from_major = storage_clone.get_major(&student.current_major_id).await?;
            let to_major = storage_clone.get_major(&target_major_id).await?;
            
            let transfer_result = process_transfer(&student, &from_major, &to_major)?;
            
            let mut students = storage_clone.students.write().await;
            if let Some(s) = students.get_mut(&student_id_clone) {
                s.current_major_id = target_major_id.clone();
            }
            drop(students);
            
            Ok(transfer_result)
        }.await;

        let mut transfer_locks = self.transfer_locks.lock().await;
        transfer_locks.remove(&student_id);
        
        result
    }
}
