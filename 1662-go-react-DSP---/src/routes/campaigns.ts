import express, { Request, Response } from 'express';
import {
  getCampaignById,
  getCampaignByName,
  getAllCampaigns,
  createCampaign as dbCreateCampaign,
  updateCampaign as dbUpdateCampaign,
  updateCampaignStatus
} from '../database';
import {
  CampaignStatus,
  CampaignCreateRequest,
  CampaignUpdateRequest,
  BiddingStrategy
} from '../types';

const router = express.Router();

function isValidBiddingStrategy(value: any): boolean {
  return Object.values(BiddingStrategy).includes(value as BiddingStrategy);
}

function isValidCampaignStatus(value: any): boolean {
  return Object.values(CampaignStatus).includes(value as CampaignStatus);
}

function isValidTargeting(targeting: any): boolean {
  if (targeting.age !== undefined) {
    if (!Array.isArray(targeting.age) || !targeting.age.every((n: any) => typeof n === 'number')) {
      return false;
    }
  }
  if (targeting.gender !== undefined) {
    if (!Array.isArray(targeting.gender)) {
      return false;
    }
  }
  if (targeting.location !== undefined) {
    if (!Array.isArray(targeting.location)) {
      return false;
    }
  }
  if (targeting.interests !== undefined) {
    if (!Array.isArray(targeting.interests)) {
      return false;
    }
  }
  return true;
}

router.get('/', async (req: Request, res: Response) => {
  try {
    const campaigns = await getAllCampaigns();
    res.json(campaigns);
  } catch (err) {
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.get('/:id', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const campaign = await getCampaignById(id);
    if (!campaign) {
      return res.status(404).json({ error: 'Campaign not found' });
    }
    res.json(campaign);
  } catch (err) {
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.post('/', async (req: Request, res: Response) => {
  try {
    const body = req.body;
    const {
      name,
      dailyBudget,
      totalBudget,
      targeting,
      biddingStrategy,
      bidAmount
    } = body;

    if (typeof name !== 'string' || name.trim() === '') {
      return res.status(400).json({ error: 'Invalid name' });
    }

    if (typeof dailyBudget !== 'number' || dailyBudget <= 0) {
      return res.status(400).json({ error: 'Daily budget must be greater than 0' });
    }

    if (typeof totalBudget !== 'number' || totalBudget <= 0) {
      return res.status(400).json({ error: 'Total budget must be greater than 0' });
    }

    if (dailyBudget > totalBudget) {
      return res.status(400).json({ error: 'Daily budget cannot exceed total budget' });
    }

    if (typeof bidAmount !== 'number' || bidAmount <= 0) {
      return res.status(400).json({ error: 'Bid amount must be greater than 0' });
    }

    if (!isValidBiddingStrategy(biddingStrategy)) {
      return res.status(400).json({ error: 'Invalid bidding strategy' });
    }

    const existing = await getCampaignByName(name);
    if (existing) {
      return res.status(409).json({ error: 'Campaign name already exists' });
    }

    const reqData: CampaignCreateRequest = {
      name,
      dailyBudget,
      totalBudget,
      targeting: targeting || {},
      biddingStrategy,
      bidAmount
    };

    const created = await dbCreateCampaign(reqData);
    res.status(201).json(created);
  } catch (err) {
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.put('/:id', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const body = req.body;

    const existing = await getCampaignById(id);
    if (!existing) {
      return res.status(404).json({ error: 'Campaign not found' });
    }

    const updateData: CampaignUpdateRequest = {};

    if (body.name !== undefined) {
      if (typeof body.name !== 'string' || body.name.trim() === '') {
        return res.status(400).json({ error: 'Invalid name' });
      }
      if (body.name !== existing.name) {
        const duplicate = await getCampaignByName(body.name);
        if (duplicate) {
          return res.status(409).json({ error: 'Campaign name already exists' });
        }
      }
      updateData.name = body.name;
    }

    if (body.dailyBudget !== undefined) {
      if (typeof body.dailyBudget !== 'number' || body.dailyBudget <= 0) {
        return res.status(400).json({ error: 'Daily budget must be greater than 0' });
      }
      updateData.dailyBudget = body.dailyBudget;
    }

    if (body.totalBudget !== undefined) {
      if (typeof body.totalBudget !== 'number' || body.totalBudget <= 0) {
        return res.status(400).json({ error: 'Total budget must be greater than 0' });
      }
      updateData.totalBudget = body.totalBudget;
    }

    if (body.bidAmount !== undefined) {
      if (typeof body.bidAmount !== 'number' || body.bidAmount <= 0) {
        return res.status(400).json({ error: 'Bid amount must be greater than 0' });
      }
      updateData.bidAmount = body.bidAmount;
    }

    if (body.biddingStrategy !== undefined) {
      if (!isValidBiddingStrategy(body.biddingStrategy)) {
        return res.status(400).json({ error: 'Invalid bidding strategy' });
      }
      updateData.biddingStrategy = body.biddingStrategy;
    }

    if (body.targeting !== undefined) {
      if (!isValidTargeting(body.targeting)) {
        return res.status(400).json({ error: 'Invalid targeting' });
      }
      updateData.targeting = body.targeting;
    }

    if (body.status !== undefined) {
      if (!isValidCampaignStatus(body.status)) {
        return res.status(400).json({ error: 'Invalid status' });
      }
      if (existing.status === CampaignStatus.ENDED && body.status === CampaignStatus.ACTIVE) {
        return res.status(400).json({ error: 'Ended campaigns cannot be reactivated' });
      }
      updateData.status = body.status;
    }

    const updated = await dbUpdateCampaign(id, updateData);
    if (!updated) {
      return res.status(404).json({ error: 'Campaign not found' });
    }

    res.json(updated);
  } catch (err) {
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.delete('/:id', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const existing = await getCampaignById(id);
    if (!existing) {
      return res.status(404).json({ error: 'Campaign not found' });
    }
    await updateCampaignStatus(id, CampaignStatus.ENDED);
    res.status(204).send();
  } catch (err) {
    res.status(500).json({ error: 'Internal server error' });
  }
});

export default router;
