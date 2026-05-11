use std::collections::HashMap;
use std::sync::RwLock;

use chrono::{Duration, Utc};
use uuid::Uuid;

use crate::errors::RecruitError;
use crate::models::*;

pub struct RecruitmentService {
    candidates: RwLock<HashMap<Uuid, Candidate>>,
    interviews: RwLock<HashMap<Uuid, Interview>>,
    offers: RwLock<HashMap<Uuid, Offer>>,
}

impl Default for RecruitmentService {
    fn default() -> Self {
        Self::new()
    }
}

impl RecruitmentService {
    pub fn new() -> Self {
        Self {
            candidates: RwLock::new(HashMap::new()),
            interviews: RwLock::new(HashMap::new()),
            offers: RwLock::new(HashMap::new()),
        }
    }

    pub fn create_candidate(&self, req: CreateCandidateRequest) -> Result<Candidate, RecruitError> {
        if req.name.is_empty() {
            return Err(RecruitError::InvalidInput("name cannot be empty".to_string()));
        }
        if req.email.is_empty() {
            return Err(RecruitError::InvalidInput("email cannot be empty".to_string()));
        }

        let mut candidate = Candidate::new(req.name, req.email);
        candidate.phone = req.phone;
        candidate.resume = req.resume;

        let id = candidate.id;
        self.candidates
            .write()
            .map_err(|_| RecruitError::InvalidInput("lock error".to_string()))?
            .insert(id, candidate.clone());

        Ok(candidate)
    }

    pub fn get_candidate(&self, id: Uuid) -> Result<Candidate, RecruitError> {
        self.candidates
            .read()
            .map_err(|_| RecruitError::InvalidInput("lock error".to_string()))?
            .get(&id)
            .cloned()
            .ok_or_else(|| RecruitError::CandidateNotFound(id.to_string()))
    }

    pub fn list_candidates(&self) -> Result<Vec<Candidate>, RecruitError> {
        let mut candidates: Vec<Candidate> = self
            .candidates
            .read()
            .map_err(|_| RecruitError::InvalidInput("lock error".to_string()))?
            .values()
            .cloned()
            .collect();
        candidates.sort_by(|a, b| b.created_at.cmp(&a.created_at));
        Ok(candidates)
    }

    pub fn initial_screening(
        &self,
        candidate_id: Uuid,
        req: InitialScreeningRequest,
    ) -> Result<Candidate, RecruitError> {
        let mut candidates = self
            .candidates
            .write()
            .map_err(|_| RecruitError::InvalidInput("lock error".to_string()))?;

        let candidate = candidates
            .get_mut(&candidate_id)
            .ok_or_else(|| RecruitError::CandidateNotFound(candidate_id.to_string()))?;

        if candidate.status != CandidateStatus::InitialScreening {
            return Err(RecruitError::InvalidStateTransition(format!(
                "candidate is not in initial screening state: {:?}",
                candidate.status
            )));
        }

        candidate.status = if req.passed {
            CandidateStatus::SecondScreening
        } else {
            CandidateStatus::InitialScreeningRejected
        };
        candidate.updated_at = Utc::now();

        Ok(candidate.clone())
    }

    pub fn second_screening(
        &self,
        candidate_id: Uuid,
        req: SecondScreeningRequest,
    ) -> Result<Candidate, RecruitError> {
        let mut candidates = self
            .candidates
            .write()
            .map_err(|_| RecruitError::InvalidInput("lock error".to_string()))?;

        let candidate = candidates
            .get_mut(&candidate_id)
            .ok_or_else(|| RecruitError::CandidateNotFound(candidate_id.to_string()))?;

        if candidate.status != CandidateStatus::SecondScreening {
            return Err(RecruitError::InvalidStateTransition(format!(
                "candidate is not in second screening state: {:?}",
                candidate.status
            )));
        }

        let now = Utc::now();
        candidate.status = match req.decision {
            SecondScreeningDecision::RecommendInterview => {
                CandidateStatus::RecommendedForInterview
            }
            SecondScreeningDecision::Pending => {
                candidate.pending_since = Some(now);
                CandidateStatus::Pending
            }
            SecondScreeningDecision::NotSuitable => CandidateStatus::NotSuitable,
        };
        candidate.updated_at = now;

        Ok(candidate.clone())
    }

    pub fn check_and_update_expired_pending(&self) -> Result<Vec<Uuid>, RecruitError> {
        let mut candidates = self
            .candidates
            .write()
            .map_err(|_| RecruitError::InvalidInput("lock error".to_string()))?;

        let now = Utc::now();
        let mut expired = Vec::new();

        for candidate in candidates.values_mut() {
            if candidate.status == CandidateStatus::Pending {
                if let Some(pending_since) = candidate.pending_since {
                    if now.signed_duration_since(pending_since) > Duration::days(14) {
                        candidate.status = CandidateStatus::PendingExpired;
                        candidate.updated_at = now;
                        expired.push(candidate.id);
                    }
                }
            }
        }

        Ok(expired)
    }

    pub fn confirm_expired_candidate(
        &self,
        candidate_id: Uuid,
        req: ConfirmExpiredCandidateRequest,
    ) -> Result<Candidate, RecruitError> {
        let mut candidates = self
            .candidates
            .write()
            .map_err(|_| RecruitError::InvalidInput("lock error".to_string()))?;

        let candidate = candidates
            .get_mut(&candidate_id)
            .ok_or_else(|| RecruitError::CandidateNotFound(candidate_id.to_string()))?;

        if candidate.status != CandidateStatus::PendingExpired {
            return Err(RecruitError::InvalidStateTransition(format!(
                "candidate is not in expired state: {:?}",
                candidate.status
            )));
        }

        let now = Utc::now();
        if req.confirmed {
            candidate.status = CandidateStatus::RecommendedForInterview;
            candidate.pending_since = None;
        } else {
            candidate.status = CandidateStatus::NotSuitable;
        }
        candidate.updated_at = now;

        Ok(candidate.clone())
    }

    pub fn schedule_interview(
        &self,
        candidate_id: Uuid,
        req: ScheduleInterviewRequest,
    ) -> Result<Interview, RecruitError> {
        {
            let candidates = self
                .candidates
                .read()
                .map_err(|_| RecruitError::InvalidInput("lock error".to_string()))?;

            let candidate = candidates
                .get(&candidate_id)
                .ok_or_else(|| RecruitError::CandidateNotFound(candidate_id.to_string()))?;

            if candidate.status == CandidateStatus::Pending {
                return Err(RecruitError::CandidatePending);
            }
            if candidate.status == CandidateStatus::PendingExpired {
                return Err(RecruitError::CandidateExpired);
            }
        }

        let duration = req
            .duration_minutes
            .unwrap_or_else(|| req.interview_type.default_duration());

        let now = Utc::now();
        let new_interview = Interview {
            id: Uuid::new_v4(),
            candidate_id,
            interview_type: req.interview_type,
            round: req.round,
            interviewer: req.interviewer.clone(),
            start_time: req.start_time,
            duration_minutes: duration,
            status: InterviewStatus::Scheduled,
            notes: None,
            created_at: now,
            updated_at: now,
        };

        let new_start = new_interview.start_time;
        let new_end = new_interview.end_time();

        let interviews = self
            .interviews
            .read()
            .map_err(|_| RecruitError::InvalidInput("lock error".to_string()))?;

        for interview in interviews.values() {
            if interview.status == InterviewStatus::Cancelled {
                continue;
            }

            if interview.interviewer == new_interview.interviewer {
                let existing_start = interview.start_time;
                let existing_end = interview.end_time();

                let with_buffer_start = existing_start - Duration::minutes(10);
                let with_buffer_end = existing_end + Duration::minutes(10);

                if !(new_end <= with_buffer_start || new_start >= with_buffer_end) {
                    return Err(RecruitError::InterviewerTimeConflict(
                        req.interviewer.clone(),
                    ));
                }
            }

            if interview.candidate_id == candidate_id {
                let existing_start = interview.start_time;
                let existing_end = interview.end_time();

                let with_buffer_start = existing_start - Duration::minutes(10);
                let with_buffer_end = existing_end + Duration::minutes(10);

                if !(new_end <= with_buffer_start || new_start >= with_buffer_end) {
                    return Err(RecruitError::NoBufferTime);
                }
            }
        }

        self.interviews
            .write()
            .map_err(|_| RecruitError::InvalidInput("lock error".to_string()))?
            .insert(new_interview.id, new_interview.clone());

        let mut candidates = self
            .candidates
            .write()
            .map_err(|_| RecruitError::InvalidInput("lock error".to_string()))?;

        if let Some(candidate) = candidates.get_mut(&candidate_id) {
            if candidate.status == CandidateStatus::RecommendedForInterview {
                candidate.status = CandidateStatus::Interviewing;
                candidate.updated_at = now;
            }
        }

        Ok(new_interview)
    }

    pub fn get_interview(&self, id: Uuid) -> Result<Interview, RecruitError> {
        self.interviews
            .read()
            .map_err(|_| RecruitError::InvalidInput("lock error".to_string()))?
            .get(&id)
            .cloned()
            .ok_or_else(|| RecruitError::InterviewNotFound(id.to_string()))
    }

    pub fn list_interviews(&self, candidate_id: Option<Uuid>) -> Result<Vec<Interview>, RecruitError> {
        let interviews: Vec<Interview> = self
            .interviews
            .read()
            .map_err(|_| RecruitError::InvalidInput("lock error".to_string()))?
            .values()
            .cloned()
            .filter(|i| candidate_id.map_or(true, |id| i.candidate_id == id))
            .collect();
        Ok(interviews)
    }

    pub fn submit_interview_result(
        &self,
        interview_id: Uuid,
        req: InterviewResultRequest,
    ) -> Result<Interview, RecruitError> {
        let mut interviews = self
            .interviews
            .write()
            .map_err(|_| RecruitError::InvalidInput("lock error".to_string()))?;

        let interview = interviews
            .get_mut(&interview_id)
            .ok_or_else(|| RecruitError::InterviewNotFound(interview_id.to_string()))?;

        if interview.status != InterviewStatus::Scheduled
            && interview.status != InterviewStatus::Completed
        {
            return Err(RecruitError::InvalidStateTransition(format!(
                "interview is not in scheduled/completed state: {:?}",
                interview.status
            )));
        }

        interview.status = if req.passed {
            InterviewStatus::Passed
        } else {
            InterviewStatus::Rejected
        };
        interview.notes = req.notes;
        interview.updated_at = Utc::now();

        let candidate_id = interview.candidate_id;
        let passed = req.passed;
        let interview_clone = interview.clone();

        drop(interviews);

        let mut candidates = self
            .candidates
            .write()
            .map_err(|_| RecruitError::InvalidInput("lock error".to_string()))?;

        if let Some(candidate) = candidates.get_mut(&candidate_id) {
            candidate.updated_at = Utc::now();
            if !passed {
                candidate.status = CandidateStatus::InterviewRejected;
            }
        }

        Ok(interview_clone)
    }

    pub fn create_offer(
        &self,
        candidate_id: Uuid,
        req: CreateOfferRequest,
    ) -> Result<Offer, RecruitError> {
        {
            let candidates = self
                .candidates
                .read()
                .map_err(|_| RecruitError::InvalidInput("lock error".to_string()))?;

            let candidate = candidates
                .get(&candidate_id)
                .ok_or_else(|| RecruitError::CandidateNotFound(candidate_id.to_string()))?;

            if candidate.status != CandidateStatus::Interviewing
                && candidate.status != CandidateStatus::RecommendedForInterview
                && candidate.status != CandidateStatus::InterviewPassed
            {
                return Err(RecruitError::InvalidStateTransition(format!(
                    "candidate is not eligible for offer: {:?}",
                    candidate.status
                )));
            }
        }

        if req.salary == 0 {
            return Err(RecruitError::InvalidInput("salary must be greater than 0".to_string()));
        }
        if req.position.is_empty() {
            return Err(RecruitError::InvalidInput("position cannot be empty".to_string()));
        }
        if req.department.is_empty() {
            return Err(RecruitError::InvalidInput("department cannot be empty".to_string()));
        }

        let offer = Offer::new(candidate_id, req.salary, req.position, req.department);
        let offer_id = offer.id;

        self.offers
            .write()
            .map_err(|_| RecruitError::InvalidInput("lock error".to_string()))?
            .insert(offer_id, offer.clone());

        Ok(offer)
    }

    pub fn get_offer(&self, id: Uuid) -> Result<Offer, RecruitError> {
        self.offers
            .read()
            .map_err(|_| RecruitError::InvalidInput("lock error".to_string()))?
            .get(&id)
            .cloned()
            .ok_or_else(|| RecruitError::OfferNotFound(id.to_string()))
    }

    pub fn list_offers(&self, candidate_id: Option<Uuid>) -> Result<Vec<Offer>, RecruitError> {
        let offers: Vec<Offer> = self
            .offers
            .read()
            .map_err(|_| RecruitError::InvalidInput("lock error".to_string()))?
            .values()
            .cloned()
            .filter(|o| candidate_id.map_or(true, |id| o.candidate_id == id))
            .collect();
        Ok(offers)
    }

    pub fn submit_offer_for_approval(&self, offer_id: Uuid) -> Result<Offer, RecruitError> {
        let mut offers = self
            .offers
            .write()
            .map_err(|_| RecruitError::InvalidInput("lock error".to_string()))?;

        let offer = offers
            .get_mut(&offer_id)
            .ok_or_else(|| RecruitError::OfferNotFound(offer_id.to_string()))?;

        if offer.status != OfferStatus::Draft && offer.status != OfferStatus::Revised {
            return Err(RecruitError::InvalidStateTransition(format!(
                "offer is not in draft/revised state: {:?}",
                offer.status
            )));
        }

        offer.status = OfferStatus::PendingApproval;
        offer.updated_at = Utc::now();

        let candidate_id = offer.candidate_id;
        let offer_clone = offer.clone();

        drop(offers);

        let mut candidates = self
            .candidates
            .write()
            .map_err(|_| RecruitError::InvalidInput("lock error".to_string()))?;

        if let Some(candidate) = candidates.get_mut(&candidate_id) {
            candidate.status = CandidateStatus::OfferPending;
            candidate.updated_at = Utc::now();
        }

        Ok(offer_clone)
    }

    pub fn approve_offer(
        &self,
        offer_id: Uuid,
        req: ApproveOfferRequest,
    ) -> Result<Offer, RecruitError> {
        let mut offers = self
            .offers
            .write()
            .map_err(|_| RecruitError::InvalidInput("lock error".to_string()))?;

        let offer = offers
            .get_mut(&offer_id)
            .ok_or_else(|| RecruitError::OfferNotFound(offer_id.to_string()))?;

        if offer.status != OfferStatus::PendingApproval {
            return Err(RecruitError::OfferNotPending);
        }

        if !req.approver_level.can_approve(&offer.approval_level) {
            let required = match offer.approval_level {
                ApprovalLevel::HRDirector => "HR Director",
                ApprovalLevel::VP => "VP",
                ApprovalLevel::CEO => "CEO",
            };
            return Err(RecruitError::InsufficientApprovalLevel(required.to_string()));
        }

        offer.status = OfferStatus::Approved;
        offer.updated_at = Utc::now();

        let candidate_id = offer.candidate_id;
        let offer_clone = offer.clone();

        drop(offers);

        let mut candidates = self
            .candidates
            .write()
            .map_err(|_| RecruitError::InvalidInput("lock error".to_string()))?;

        if let Some(candidate) = candidates.get_mut(&candidate_id) {
            candidate.status = CandidateStatus::OfferApproved;
            candidate.updated_at = Utc::now();
        }

        Ok(offer_clone)
    }

    pub fn reject_offer(
        &self,
        offer_id: Uuid,
        req: ApproveOfferRequest,
    ) -> Result<Offer, RecruitError> {
        let mut offers = self
            .offers
            .write()
            .map_err(|_| RecruitError::InvalidInput("lock error".to_string()))?;

        let offer = offers
            .get_mut(&offer_id)
            .ok_or_else(|| RecruitError::OfferNotFound(offer_id.to_string()))?;

        if offer.status != OfferStatus::PendingApproval {
            return Err(RecruitError::OfferNotPending);
        }

        if !req.approver_level.can_approve(&offer.approval_level) {
            let required = match offer.approval_level {
                ApprovalLevel::HRDirector => "HR Director",
                ApprovalLevel::VP => "VP",
                ApprovalLevel::CEO => "CEO",
            };
            return Err(RecruitError::InsufficientApprovalLevel(required.to_string()));
        }

        offer.status = OfferStatus::Rejected;
        offer.updated_at = Utc::now();

        let candidate_id = offer.candidate_id;
        let offer_clone = offer.clone();

        drop(offers);

        let mut candidates = self
            .candidates
            .write()
            .map_err(|_| RecruitError::InvalidInput("lock error".to_string()))?;

        if let Some(candidate) = candidates.get_mut(&candidate_id) {
            candidate.status = CandidateStatus::OfferRejected;
            candidate.updated_at = Utc::now();
        }

        Ok(offer_clone)
    }

    pub fn revise_offer(
        &self,
        offer_id: Uuid,
        req: ReviseOfferRequest,
    ) -> Result<Offer, RecruitError> {
        let mut offers = self
            .offers
            .write()
            .map_err(|_| RecruitError::InvalidInput("lock error".to_string()))?;

        let offer = offers
            .get_mut(&offer_id)
            .ok_or_else(|| RecruitError::OfferNotFound(offer_id.to_string()))?;

        if offer.status != OfferStatus::PendingApproval
            && offer.status != OfferStatus::Rejected
            && offer.status != OfferStatus::Approved
        {
            return Err(RecruitError::InvalidStateTransition(format!(
                "offer cannot be revised in current state: {:?}",
                offer.status
            )));
        }

        if req.salary == 0 {
            return Err(RecruitError::InvalidInput("salary must be greater than 0".to_string()));
        }
        if req.position.is_empty() {
            return Err(RecruitError::InvalidInput("position cannot be empty".to_string()));
        }
        if req.department.is_empty() {
            return Err(RecruitError::InvalidInput("department cannot be empty".to_string()));
        }

        offer.revise(req.salary, req.position, req.department);
        Ok(offer.clone())
    }
}
