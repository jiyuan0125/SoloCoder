use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use uuid::Uuid;

use crate::models::{
    ApiResponse, Claim, CreateClaimRequest, FinalRejectRequest, ReturnRequest, ResubmitRequest,
    TimeoutCheckResult, UpdateClaimRequest,
};

#[derive(Clone)]
pub struct ClaimService {
    claims: Arc<Mutex<HashMap<Uuid, Claim>>>,
}

impl ClaimService {
    pub fn new() -> Self {
        Self {
            claims: Arc::new(Mutex::new(HashMap::new())),
        }
    }

    pub fn create_claim(&self, req: CreateClaimRequest) -> ApiResponse<Claim> {
        let mut claim = Claim::new(
            req.employee_id,
            req.title,
            req.description,
            req.amount,
        );
        claim.start_audit();
        let id = claim.id;
        let mut map = self.claims.lock().unwrap();
        map.insert(id, claim.clone());
        ApiResponse::success(claim)
    }

    pub fn get_claim(&self, id: Uuid) -> ApiResponse<Claim> {
        let map = self.claims.lock().unwrap();
        match map.get(&id) {
            Some(claim) => ApiResponse::success(claim.clone()),
            None => ApiResponse::error("理赔申请不存在".to_string()),
        }
    }

    pub fn list_claims(&self) -> ApiResponse<Vec<Claim>> {
        let map = self.claims.lock().unwrap();
        let claims: Vec<Claim> = map.values().cloned().collect();
        ApiResponse::success(claims)
    }

    pub fn update_claim(&self, id: Uuid, req: UpdateClaimRequest) -> ApiResponse<Claim> {
        let mut map = self.claims.lock().unwrap();
        match map.get_mut(&id) {
            Some(claim) => {
                if claim.current_stage.is_terminal() {
                    return ApiResponse::error("申请已处于终态，无法修改".to_string());
                }
                claim.update_content(req.title, req.description, req.amount);
                ApiResponse::success(claim.clone())
            }
            None => ApiResponse::error("理赔申请不存在".to_string()),
        }
    }

    pub fn approve(&self, id: Uuid, operator: String) -> ApiResponse<Claim> {
        let mut map = self.claims.lock().unwrap();
        match map.get_mut(&id) {
            Some(claim) => match claim.approve_current(&operator) {
                Ok(()) => ApiResponse::success(claim.clone()),
                Err(e) => ApiResponse::error(e),
            },
            None => ApiResponse::error("理赔申请不存在".to_string()),
        }
    }

    pub fn return_to(&self, id: Uuid, req: ReturnRequest) -> ApiResponse<Claim> {
        if req.reason.trim().is_empty() {
            return ApiResponse::error("退回原因不能为空".to_string());
        }
        let mut map = self.claims.lock().unwrap();
        match map.get_mut(&id) {
            Some(claim) => match claim.return_to(req.target_stage, req.reason, &req.operator) {
                Ok(()) => ApiResponse::success(claim.clone()),
                Err(e) => ApiResponse::error(e),
            },
            None => ApiResponse::error("理赔申请不存在".to_string()),
        }
    }

    pub fn final_reject(&self, id: Uuid, req: FinalRejectRequest) -> ApiResponse<Claim> {
        if req.reason.trim().is_empty() {
            return ApiResponse::error("驳回原因不能为空".to_string());
        }
        let mut map = self.claims.lock().unwrap();
        match map.get_mut(&id) {
            Some(claim) => match claim.final_reject(req.reason, &req.operator) {
                Ok(()) => ApiResponse::success(claim.clone()),
                Err(e) => ApiResponse::error(e),
            },
            None => ApiResponse::error("理赔申请不存在".to_string()),
        }
    }

    pub fn resubmit(&self, id: Uuid, req: ResubmitRequest) -> ApiResponse<Claim> {
        let mut map = self.claims.lock().unwrap();
        match map.get_mut(&id) {
            Some(claim) => match claim.resubmit(&req.employee_id) {
                Ok(()) => ApiResponse::success(claim.clone()),
                Err(e) => ApiResponse::error(e),
            },
            None => ApiResponse::error("理赔申请不存在".to_string()),
        }
    }

    pub fn check_all_timeouts(&self) -> Vec<(Uuid, TimeoutCheckResult)> {
        let mut map = self.claims.lock().unwrap();
        let mut results = Vec::new();
        for (id, claim) in map.iter_mut() {
            let result = claim.check_timeout();
            if !matches!(result, TimeoutCheckResult::NoChange | TimeoutCheckResult::NoDeadline) {
                results.push((*id, result));
            }
        }
        results
    }
}

impl Default for ClaimService {
    fn default() -> Self {
        Self::new()
    }
}
