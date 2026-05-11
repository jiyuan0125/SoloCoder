use rust_decimal::Decimal;
use chrono::Utc;
use crate::error::ClaimError;
use crate::models::*;
use crate::storage::InMemoryStorage;

#[derive(Debug, Clone)]
pub struct ClaimService {
    storage: InMemoryStorage,
}

impl ClaimService {
    pub fn new(storage: InMemoryStorage) -> Self {
        ClaimService { storage }
    }

    pub async fn create_policy(&self, req: CreatePolicyRequest) -> Result<Policy, ClaimError> {
        if req.deductible < Decimal::ZERO {
            return Err(ClaimError::InvalidAmount);
        }
        
        let policy = Policy::new(
            req.policy_number,
            req.holder_name,
            req.deductible.round_dp(2),
            req.policy_year,
        );
        
        self.storage.insert_policy(policy.clone()).await;
        Ok(policy)
    }

    pub async fn get_policy(&self, id: &str) -> Result<Policy, ClaimError> {
        self.storage
            .get_policy(id)
            .await
            .ok_or(ClaimError::PolicyNotFound)
    }

    pub async fn get_all_policies(&self) -> Vec<Policy> {
        self.storage.get_all_policies().await
    }

    pub async fn create_claim(&self, req: CreateClaimRequest) -> Result<Claim, ClaimError> {
        if req.total_loss <= Decimal::ZERO {
            return Err(ClaimError::InvalidAmount);
        }

        let parties = self.validate_and_build_parties(req.parties).await?;
        let claim = Claim::new(
            req.case_number,
            req.total_loss.round_dp(2),
            parties,
        );

        self.storage.insert_claim(claim.clone()).await;
        Ok(claim)
    }

    async fn validate_and_build_parties(&self, req_parties: Vec<CreatePartyRequest>) -> Result<Vec<Party>, ClaimError> {
        let mut total_ratio = Decimal::ZERO;
        let mut parties: Vec<Party> = Vec::new();
        let mut party_ids: Vec<String> = Vec::new();

        for req in req_parties {
            if req.liability_ratio <= Decimal::ZERO || req.liability_ratio > Decimal::ONE_HUNDRED {
                return Err(ClaimError::InvalidLiabilityRatio);
            }

            if req.own_loss < Decimal::ZERO {
                return Err(ClaimError::InvalidAmount);
            }

            if self.storage.get_policy(&req.policy_id).await.is_none() {
                return Err(ClaimError::PolicyNotFound);
            }

            if party_ids.contains(&req.policy_id) {
                return Err(ClaimError::DuplicateParty);
            }
            party_ids.push(req.policy_id.clone());

            total_ratio += req.liability_ratio;
            
            parties.push(Party::new(
                req.name,
                req.policy_id,
                req.liability_ratio.round_dp(2),
                req.own_loss.round_dp(2),
            ));
        }

        if !total_ratio.round_dp(2).eq(&Decimal::ONE_HUNDRED) {
            return Err(ClaimError::InvalidLiabilityRatioSum(total_ratio.to_string()));
        }

        Ok(parties)
    }

    pub async fn get_claim(&self, id: &str) -> Result<Claim, ClaimError> {
        self.storage
            .get_claim(id)
            .await
            .ok_or(ClaimError::ClaimNotFound)
    }

    pub async fn get_all_claims(&self) -> Vec<Claim> {
        self.storage.get_all_claims().await
    }

    pub async fn close_claim(&self, claim_id: &str) -> Result<Claim, ClaimError> {
        let mut claim = self.get_claim(claim_id).await?;
        
        if claim.is_closed() {
            return Err(ClaimError::ClaimClosed);
        }

        let settlement = self.calculate_settlement(&claim).await?;
        
        claim.status = ClaimStatus::Closed;
        claim.closed_at = Some(Utc::now());
        claim.settlement = Some(settlement);

        self.storage.update_claim(claim.clone()).await;
        Ok(claim)
    }

    async fn calculate_settlement(&self, claim: &Claim) -> Result<SettlementSummary, ClaimError> {
        let mut party_settlements = Vec::new();
        let mut total_payout = Decimal::ZERO;
        let mut total_self_bear = Decimal::ZERO;

        for party in &claim.parties {
            let policy = self.storage
                .get_policy(&party.policy_id)
                .await
                .ok_or(ClaimError::PolicyNotFound)?;

            let ratio_percent = party.liability_ratio / Decimal::ONE_HUNDRED;
            let assumed_amount = (claim.total_loss * ratio_percent).round_dp(2);
            
            let deductible_applied = policy.deductible;
            
            let payout_amount = if assumed_amount <= deductible_applied {
                Decimal::ZERO
            } else {
                (assumed_amount - deductible_applied).round_dp(2)
            };

            let self_bear_amount = (assumed_amount - payout_amount).round_dp(2);

            total_payout += payout_amount;
            total_self_bear += self_bear_amount;

            party_settlements.push(PartySettlement {
                party_id: party.id.clone(),
                party_name: party.name.clone(),
                liability_ratio: party.liability_ratio,
                own_loss: party.own_loss,
                assumed_amount,
                deductible_applied,
                payout_amount,
                self_bear_amount,
            });
        }

        Ok(SettlementSummary {
            claim_id: claim.id.clone(),
            total_loss: claim.total_loss,
            party_settlements,
            total_payout: total_payout.round_dp(2),
            total_self_bear: total_self_bear.round_dp(2),
        })
    }

    pub async fn update_claim_parties(
        &self,
        claim_id: &str,
        parties: Vec<CreatePartyRequest>,
    ) -> Result<Claim, ClaimError> {
        let mut claim = self.get_claim(claim_id).await?;
        
        if claim.is_closed() {
            return Err(ClaimError::ClaimClosed);
        }

        claim.parties = self.validate_and_build_parties(parties).await?;
        self.storage.update_claim(claim.clone()).await;
        Ok(claim)
    }

    pub async fn update_claim_total_loss(
        &self,
        claim_id: &str,
        total_loss: Decimal,
    ) -> Result<Claim, ClaimError> {
        let mut claim = self.get_claim(claim_id).await?;
        
        if claim.is_closed() {
            return Err(ClaimError::ClaimClosed);
        }

        if total_loss <= Decimal::ZERO {
            return Err(ClaimError::InvalidAmount);
        }

        claim.total_loss = total_loss.round_dp(2);
        self.storage.update_claim(claim.clone()).await;
        Ok(claim)
    }
}
