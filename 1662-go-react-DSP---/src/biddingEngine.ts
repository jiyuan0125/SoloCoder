import {
  Campaign,
  UserProfile,
  Targeting,
  BidParticipant,
  BiddingStrategy
} from './types';

const MINIMUM_BID = 0.01;

function matchTargeting(targeting: Targeting, profile: UserProfile): boolean {
  if (targeting.age !== undefined && targeting.age.length > 0) {
    if (profile.age === undefined) return false;
    if (!targeting.age.includes(profile.age)) return false;
  }

  if (targeting.gender !== undefined && targeting.gender.length > 0) {
    if (profile.gender === undefined) return false;
    if (!targeting.gender.includes(profile.gender)) return false;
  }

  if (targeting.location !== undefined && targeting.location.length > 0) {
    if (profile.location === undefined) return false;
    if (!targeting.location.includes(profile.location)) return false;
  }

  if (targeting.interests !== undefined && targeting.interests.length > 0) {
    if (profile.interests === undefined) return false;
    const hasMatch = targeting.interests.some((interest) =>
      profile.interests!.includes(interest)
    );
    if (!hasMatch) return false;
  }

  return true;
}

function calculateBid(
  campaign: Campaign,
  impressionType: BiddingStrategy
): number | null {
  if (campaign.biddingStrategy !== impressionType) {
    return null;
  }
  return campaign.bidAmount;
}

export interface BiddingResult {
  participants: BidParticipant[];
  winner: { id: string; name: string } | null;
  finalPrice: number;
}

export function runAuction(
  activeCampaigns: Campaign[],
  userProfile: UserProfile,
  impressionType: BiddingStrategy
): BiddingResult {
  const participants: BidParticipant[] = [];

  for (const campaign of activeCampaigns) {
    if (!matchTargeting(campaign.targeting, userProfile)) {
      continue;
    }

    const bid = calculateBid(campaign, impressionType);
    if (bid === null) {
      continue;
    }

    participants.push({
      campaignId: campaign.id,
      campaignName: campaign.name,
      bidAmount: bid
    });
  }

  if (participants.length === 0) {
    return {
      participants: [],
      winner: null,
      finalPrice: 0
    };
  }

  participants.sort((a, b) => b.bidAmount - a.bidAmount);

  let finalPrice: number;
  if (participants.length === 1) {
    finalPrice = MINIMUM_BID;
  } else {
    finalPrice = participants[1].bidAmount + 0.01;
  }

  const winner = participants[0];

  return {
    participants,
    winner: { id: winner.campaignId, name: winner.campaignName },
    finalPrice
  };
}
