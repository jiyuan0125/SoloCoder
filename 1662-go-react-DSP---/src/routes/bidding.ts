import express, { Request, Response } from 'express';
import {
  getActiveCampaigns,
  checkDailyAndTotalBudget,
  incrementBidCount,
  incrementWinCount,
  deductBudget,
  createBidLog,
  updateCampaignStatus
} from '../database';
import { runAuction } from '../biddingEngine';
import {
  BidRequest,
  BiddingStrategy,
  CampaignStatus,
  BidLogStatus
} from '../types';

const router = express.Router();

function isValidBiddingStrategy(value: any): boolean {
  return Object.values(BiddingStrategy).includes(value as BiddingStrategy);
}

router.post('/bid', async (req: Request, res: Response) => {
  try {
    const body = req.body as Partial<BidRequest>;
    const { requestId, userProfile, impressionType } = body;

    if (typeof requestId !== 'string' || requestId.trim() === '') {
      return res.status(400).json({ error: 'Invalid requestId' });
    }

    if (typeof userProfile !== 'object' || userProfile === null) {
      return res.status(400).json({ error: 'Invalid userProfile' });
    }

    if (!isValidBiddingStrategy(impressionType)) {
      return res.status(400).json({ error: 'Invalid impressionType' });
    }

    await checkDailyAndTotalBudget();

    const activeCampaigns = await getActiveCampaigns();

    const result = runAuction(
      activeCampaigns,
      userProfile,
      impressionType as BiddingStrategy
    );

    const participantIds = result.participants.map((p) => p.campaignId);
    await incrementBidCount(participantIds);

    if (!result.winner) {
      await createBidLog(
        requestId,
        result.participants,
        null,
        0,
        BidLogStatus.SUCCESS
      );
      return res.json({
        success: true,
        winner: null,
        finalPrice: 0,
        participants: result.participants
      });
    }

    const budgetResult = await deductBudget(
      result.winner.id,
      result.finalPrice
    );

    if (!budgetResult.success) {
      await updateCampaignStatus(result.winner.id, CampaignStatus.ENDED);
      await createBidLog(
        requestId,
        result.participants,
        result.winner,
        result.finalPrice,
        BidLogStatus.VOIDED
      );
      return res.status(409).json({
        error: budgetResult.dailyExceeded
          ? 'Daily budget exceeded'
          : 'Total budget exceeded',
        participants: result.participants
      });
    }

    await incrementWinCount(result.winner.id);

    const campaign = activeCampaigns.find((c) => c.id === result.winner!.id);
    if (campaign) {
      const newDailySpent = campaign.dailySpent + result.finalPrice;
      const newTotalSpent = campaign.totalSpent + result.finalPrice;
      if (
        newDailySpent >= campaign.dailyBudget ||
        newTotalSpent >= campaign.totalBudget
      ) {
        await updateCampaignStatus(result.winner.id, CampaignStatus.ENDED);
      }
    }

    await createBidLog(
      requestId,
      result.participants,
      result.winner,
      result.finalPrice,
      BidLogStatus.SUCCESS
    );

    res.json({
      success: true,
      winner: result.winner,
      finalPrice: result.finalPrice,
      participants: result.participants
    });
  } catch (err) {
    res.status(500).json({ error: 'Internal server error' });
  }
});

export default router;
