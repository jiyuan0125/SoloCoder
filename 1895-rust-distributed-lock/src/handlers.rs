use crate::lock_manager::{
    AcquireError, ForceReleaseError, LockManager, ReleaseError, RenewError,
};
use crate::models::{
    AcquireRequest, ErrorResponse, ForceReleaseRequest, LockInfoResponse, ReleaseRequest,
    RenewRequest,
};
use actix_web::{web, HttpResponse, Responder};

pub async fn acquire_lock(
    lock_manager: web::Data<LockManager>,
    lock_name: web::Path<String>,
    req: web::Json<AcquireRequest>,
) -> impl Responder {
    let lock_name = lock_name.into_inner();
    let client_id = req.client_id.clone();
    let timeout_seconds = req.timeout_seconds;
    let wait_timeout_seconds = req.wait_timeout_seconds;

    match lock_manager
        .try_acquire(&lock_name, &client_id, timeout_seconds, wait_timeout_seconds)
        .await
    {
        Ok(_) => HttpResponse::Ok().json(serde_json::json!({
            "success": true,
            "lock_name": lock_name,
            "client_id": client_id,
            "message": "Lock acquired successfully"
        })),
        Err(e) => match e {
            AcquireError::LockBusy => HttpResponse::Conflict().json(ErrorResponse {
                error: "LOCK_BUSY".to_string(),
                message: "Lock is currently held by another client".to_string(),
            }),
            AcquireError::QueueFull => HttpResponse::ServiceUnavailable().json(ErrorResponse {
                error: "QUEUE_FULL".to_string(),
                message: "Wait queue is full, maximum 20 waiters allowed".to_string(),
            }),
            AcquireError::WaitTimeout => HttpResponse::RequestTimeout().json(ErrorResponse {
                error: "WAIT_TIMEOUT".to_string(),
                message: "Wait timeout exceeded, lock not acquired".to_string(),
            }),
            AcquireError::InvalidTimeout => HttpResponse::BadRequest().json(ErrorResponse {
                error: "INVALID_TIMEOUT".to_string(),
                message: "Timeout must be greater than 0".to_string(),
            }),
        },
    }
}

pub async fn release_lock(
    lock_manager: web::Data<LockManager>,
    lock_name: web::Path<String>,
    req: web::Json<ReleaseRequest>,
) -> impl Responder {
    let lock_name = lock_name.into_inner();
    let client_id = req.client_id.clone();

    match lock_manager.release(&lock_name, &client_id).await {
        Ok(_) => HttpResponse::Ok().json(serde_json::json!({
            "success": true,
            "lock_name": lock_name,
            "client_id": client_id,
            "message": "Lock released successfully"
        })),
        Err(e) => match e {
            ReleaseError::NotFound => HttpResponse::NotFound().json(ErrorResponse {
                error: "LOCK_NOT_FOUND".to_string(),
                message: format!("Lock '{}' not found", lock_name),
            }),
            ReleaseError::NotHolder => HttpResponse::Forbidden().json(ErrorResponse {
                error: "NOT_HOLDER".to_string(),
                message: "Only the lock holder can release the lock".to_string(),
            }),
            ReleaseError::AlreadyExpired => HttpResponse::Gone().json(ErrorResponse {
                error: "ALREADY_EXPIRED".to_string(),
                message: "Lock has already expired and cannot be released".to_string(),
            }),
        },
    }
}

pub async fn renew_lock(
    lock_manager: web::Data<LockManager>,
    lock_name: web::Path<String>,
    req: web::Json<RenewRequest>,
) -> impl Responder {
    let lock_name = lock_name.into_inner();
    let client_id = req.client_id.clone();
    let timeout_seconds = req.timeout_seconds;

    match lock_manager
        .renew(&lock_name, &client_id, timeout_seconds)
        .await
    {
        Ok(_) => HttpResponse::Ok().json(serde_json::json!({
            "success": true,
            "lock_name": lock_name,
            "client_id": client_id,
            "message": "Lock renewed successfully",
            "new_timeout_seconds": timeout_seconds
        })),
        Err(e) => match e {
            RenewError::NotFound => HttpResponse::NotFound().json(ErrorResponse {
                error: "LOCK_NOT_FOUND".to_string(),
                message: format!("Lock '{}' not found", lock_name),
            }),
            RenewError::NotHolder => HttpResponse::Forbidden().json(ErrorResponse {
                error: "NOT_HOLDER".to_string(),
                message: "Only the lock holder can renew the lock".to_string(),
            }),
            RenewError::AlreadyExpired => HttpResponse::Gone().json(ErrorResponse {
                error: "ALREADY_EXPIRED".to_string(),
                message: "Lock has already expired, cannot renew. It may have been acquired by another client.".to_string(),
            }),
            RenewError::InvalidTimeout => HttpResponse::BadRequest().json(ErrorResponse {
                error: "INVALID_TIMEOUT".to_string(),
                message: "Timeout must be greater than 0".to_string(),
            }),
        },
    }
}

pub async fn list_locks(lock_manager: web::Data<LockManager>) -> impl Responder {
    let locks = lock_manager.get_all_locks().await;
    let response: Vec<LockInfoResponse> = locks.iter().map(LockInfoResponse::from).collect();
    HttpResponse::Ok().json(response)
}

pub async fn force_release_lock(
    lock_manager: web::Data<LockManager>,
    lock_name: web::Path<String>,
    req: web::Json<ForceReleaseRequest>,
) -> impl Responder {
    if !req.confirm {
        return HttpResponse::BadRequest().json(ErrorResponse {
            error: "CONFIRM_REQUIRED".to_string(),
            message: "Force release requires 'confirm: true' parameter".to_string(),
        });
    }

    let lock_name = lock_name.into_inner();

    match lock_manager.force_release(&lock_name).await {
        Ok(_) => HttpResponse::Ok().json(serde_json::json!({
            "success": true,
            "lock_name": lock_name,
            "message": "Lock forcefully released"
        })),
        Err(e) => match e {
            ForceReleaseError::NotFound => HttpResponse::NotFound().json(ErrorResponse {
                error: "LOCK_NOT_FOUND".to_string(),
                message: format!("Lock '{}' not found", lock_name),
            }),
            ForceReleaseError::AlreadyFree => HttpResponse::BadRequest().json(ErrorResponse {
                error: "ALREADY_FREE".to_string(),
                message: "Lock is already free".to_string(),
            }),
        },
    }
}

pub async fn health_check() -> impl Responder {
    HttpResponse::Ok().json(serde_json::json!({
        "status": "ok",
        "service": "distributed-lock-service"
    }))
}
