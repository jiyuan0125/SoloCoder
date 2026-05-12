import { v4 as uuidv4 } from 'uuid';
import { ProposalStage, VoteOption } from '../types';

export const generateId = (): string => uuidv4();

export const now = (): number => Date.now();

export const isValidVoteOption = (option: string): boolean => {
  return Object.values(VoteOption).includes(option as VoteOption);
};

export const isValidStageTransition = (
  currentStage: ProposalStage,
  targetStage: ProposalStage
): boolean => {
  const validTransitions: Record<ProposalStage, ProposalStage[]> = {
    [ProposalStage.SUBMITTED]: [ProposalStage.REVIEW],
    [ProposalStage.REVIEW]: [
      ProposalStage.PUBLIC_NOTICE,
      ProposalStage.REJECTED
    ],
    [ProposalStage.PUBLIC_NOTICE]: [ProposalStage.VOTING],
    [ProposalStage.VOTING]: [ProposalStage.EXECUTION, ProposalStage.REJECTED],
    [ProposalStage.EXECUTION]: [ProposalStage.EXECUTED],
    [ProposalStage.EXECUTED]: [],
    [ProposalStage.REJECTED]: []
  };
  
  return validTransitions[currentStage]?.includes(targetStage) ?? false;
};

export const calculateParticipationRate = (
  totalVotes: number,
  totalOwners: number
): number => {
  if (totalOwners === 0) return 0;
  return totalVotes / totalOwners;
};

export const calculateApprovalRate = (
  approveVotes: number,
  totalOwners: number
): number => {
  if (totalOwners === 0) return 0;
  return approveVotes / totalOwners;
};
