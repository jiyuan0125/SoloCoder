use std::collections::HashMap;
use std::sync::Arc;
use chrono::{DateTime, Duration, Utc};
use tokio::sync::RwLock;
use uuid::Uuid;

use crate::errors::ReviewError;
use crate::models::{
    CreateFollowUpRequest, CreateInitialReviewRequest, CreateReplyRequest, DeleteReviewRequest,
    FollowUpReview, InitialReview, MerchantReply, Rating, Review, ReviewFilter, ReplyTarget,
};

const FOLLOW_UP_WINDOW_DAYS: i64 = 7;

pub struct ReviewService {
    reviews: RwLock<HashMap<Uuid, Review>>,
    product_reviews: RwLock<HashMap<String, Vec<Uuid>>>,
}

impl ReviewService {
    pub fn new() -> Arc<Self> {
        Arc::new(Self {
            reviews: RwLock::new(HashMap::new()),
            product_reviews: RwLock::new(HashMap::new()),
        })
    }

    pub async fn create_initial_review(
        &self,
        req: CreateInitialReviewRequest,
    ) -> Result<Review, ReviewError> {
        if req.product_id.trim().is_empty() {
            return Err(ReviewError::EmptyProductId);
        }
        if req.customer_id.trim().is_empty() {
            return Err(ReviewError::EmptyCustomerId);
        }
        if req.content.trim().is_empty() {
            return Err(ReviewError::EmptyContent);
        }

        let rating = Rating::from_u32(req.rating).ok_or(ReviewError::InvalidRating)?;

        let now = Utc::now();
        let review_id = Uuid::new_v4();

        let review = Review {
            id: review_id,
            product_id: req.product_id.clone(),
            customer_id: req.customer_id,
            initial: InitialReview {
                rating,
                content: req.content,
                created_at: now,
            },
            follow_up: None,
            replies: Vec::new(),
            deleted_by_customer: false,
            created_at: now,
        };

        {
            let mut reviews = self.reviews.write().await;
            reviews.insert(review_id, review.clone());
        }

        {
            let mut product_reviews = self.product_reviews.write().await;
            product_reviews
                .entry(req.product_id)
                .or_insert_with(Vec::new)
                .push(review_id);
        }

        Ok(review)
    }

    pub async fn create_follow_up_review(
        &self,
        req: CreateFollowUpRequest,
    ) -> Result<Review, ReviewError> {
        if req.content.trim().is_empty() {
            return Err(ReviewError::EmptyContent);
        }

        let mut reviews = self.reviews.write().await;
        let review = reviews
            .get_mut(&req.review_id)
            .ok_or(ReviewError::ReviewNotFound)?;

        if review.customer_id != req.customer_id {
            return Err(ReviewError::PermissionDenied);
        }

        if review.follow_up.is_some() {
            return Err(ReviewError::FollowUpAlreadyExists);
        }

        let now = Utc::now();
        if !Self::is_within_follow_up_window(review.created_at, now) {
            return Err(ReviewError::FollowUpWindowClosed);
        }

        review.follow_up = Some(FollowUpReview {
            content: req.content,
            created_at: now,
        });

        Ok(review.clone())
    }

    fn is_within_follow_up_window(created_at: DateTime<Utc>, now: DateTime<Utc>) -> bool {
        let window_end = created_at + Duration::days(FOLLOW_UP_WINDOW_DAYS);
        now <= window_end
    }

    pub async fn create_merchant_reply(
        &self,
        req: CreateReplyRequest,
    ) -> Result<Review, ReviewError> {
        if req.content.trim().is_empty() {
            return Err(ReviewError::EmptyContent);
        }

        let mut reviews = self.reviews.write().await;
        let review = reviews
            .get_mut(&req.review_id)
            .ok_or(ReviewError::ReviewNotFound)?;

        let has_reply_to_target = match req.target_type {
            ReplyTarget::InitialReview => review.replies.iter().any(|r| !r.is_supplement),
            ReplyTarget::FollowUpReview => {
                if review.follow_up.is_none() {
                    return Err(ReviewError::ReviewNotFound);
                }
                review.replies.iter().filter(|r| !r.is_supplement).count() >= 2
            }
        };

        if has_reply_to_target {
            return Err(ReviewError::ReplyAlreadyExists);
        }

        let now = Utc::now();
        review.replies.push(MerchantReply {
            content: req.content,
            is_supplement: false,
            created_at: now,
        });

        Ok(review.clone())
    }

    pub async fn create_supplement_reply(
        &self,
        req: CreateReplyRequest,
    ) -> Result<Review, ReviewError> {
        if req.content.trim().is_empty() {
            return Err(ReviewError::EmptyContent);
        }

        let mut reviews = self.reviews.write().await;
        let review = reviews
            .get_mut(&req.review_id)
            .ok_or(ReviewError::ReviewNotFound)?;

        if let ReplyTarget::FollowUpReview = req.target_type {
            if review.follow_up.is_none() {
                return Err(ReviewError::ReviewNotFound);
            }
        }

        let now = Utc::now();
        review.replies.push(MerchantReply {
            content: req.content,
            is_supplement: true,
            created_at: now,
        });

        Ok(review.clone())
    }

    pub async fn delete_review(&self, req: DeleteReviewRequest) -> Result<(), ReviewError> {
        let mut reviews = self.reviews.write().await;
        let review = reviews
            .get_mut(&req.review_id)
            .ok_or(ReviewError::ReviewNotFound)?;

        if review.customer_id != req.customer_id {
            return Err(ReviewError::PermissionDenied);
        }

        review.deleted_by_customer = true;
        Ok(())
    }

    pub async fn get_reviews_for_product(
        &self,
        product_id: &str,
        filter: Option<ReviewFilter>,
        as_merchant: bool,
    ) -> Vec<Review> {
        let product_reviews = self.product_reviews.read().await;
        let review_ids = product_reviews.get(product_id).cloned().unwrap_or_default();
        
        let reviews = self.reviews.read().await;
        
        review_ids
            .into_iter()
            .filter_map(|id| reviews.get(&id).cloned())
            .filter(|review| {
                if !as_merchant && review.deleted_by_customer {
                    return false;
                }
                true
            })
            .filter(|review| Self::matches_filter(review, &filter))
            .collect()
    }

    fn matches_filter(review: &Review, filter: &Option<ReviewFilter>) -> bool {
        let Some(filter) = filter else {
            return true;
        };

        if let Some(rating) = filter.rating {
            if review.initial.rating != rating {
                return false;
            }
        }

        if let Some(has_follow_up) = filter.has_follow_up {
            if review.follow_up.is_some() != has_follow_up {
                return false;
            }
        }

        if let Some(has_reply) = filter.has_reply {
            let has_any_reply = !review.replies.is_empty();
            if has_any_reply != has_reply {
                return false;
            }
        }

        true
    }

    pub async fn get_review_by_id(
        &self,
        review_id: Uuid,
        as_merchant: bool,
    ) -> Option<Review> {
        let reviews = self.reviews.read().await;
        reviews.get(&review_id).and_then(|review| {
            if !as_merchant && review.deleted_by_customer {
                None
            } else {
                Some(review.clone())
            }
        })
    }
}
