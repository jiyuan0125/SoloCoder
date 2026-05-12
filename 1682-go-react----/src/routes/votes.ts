import { Router, Request, Response } from 'express';
import { voteService } from '../services/voteService';
import { proposalService } from '../services/proposalService';
import { VoteOption, ProposalStage } from '../types';
import { isValidVoteOption, now } from '../utils';
import { OBJECTION_THRESHOLD } from '../constants';

const router = Router();

router.post('/register', (req: Request, res: Response) => {
  const { phone, name } = req.body;
  
  if (!phone || !name) {
    return res.status(400).json({ error: 'Phone and name are required' });
  }
  
  const owner = voteService.registerOwner(phone, name);
  res.status(201).json(owner);
});

router.post('/proposals/:proposalId/vote', (req: Request, res: Response) => {
  const { proposalId } = req.params;
  const { phone, option } = req.body;
  
  const proposal = proposalService.getProposal(proposalId);
  
  if (!proposal) {
    return res.status(404).json({ error: 'Proposal not found' });
  }
  
  if (proposal.stage !== ProposalStage.VOTING) {
    return res.status(400).json({ error: 'Voting is not open for this proposal' });
  }
  
  if (proposal.votingEndTime && now() > proposal.votingEndTime) {
    return res.status(400).json({ error: 'Voting period has ended' });
  }
  
  if (!isValidVoteOption(option)) {
    return res.status(400).json({ 
      error: 'Invalid vote option. Must be approve, oppose, or abstain' 
    });
  }
  
  const owner = voteService.getOwnerByPhone(phone);
  if (!owner) {
    return res.status(403).json({ error: 'Phone number not registered' });
  }
  
  const existingVotes = voteService.getVotesByProposal(proposalId);
  const hasVoted = existingVotes.some(v => v.phone === phone);
  
  if (hasVoted) {
    return res.status(409).json({ error: 'You have already voted on this proposal' });
  }
  
  const vote = voteService.castVote(proposalId, phone, option as VoteOption);
  res.status(201).json(vote);
});

router.get('/proposals/:proposalId/result', (req: Request, res: Response) => {
  const { proposalId } = req.params;
  
  const proposal = proposalService.getProposal(proposalId);
  
  if (!proposal) {
    return res.status(404).json({ error: 'Proposal not found' });
  }
  
  if (proposal.stage === ProposalStage.VOTING) {
    if (proposal.votingEndTime && now() <= proposal.votingEndTime) {
      return res.status(403).json({ 
        error: 'Voting results are not available during voting period' 
      });
    }
  }
  
  const result = voteService.getVotingResult(proposalId);
  res.json(result);
});

router.post('/proposals/:proposalId/suggestions', (req: Request, res: Response) => {
  const { proposalId } = req.params;
  const { ownerId, content } = req.body;
  
  const proposal = proposalService.getProposal(proposalId);
  
  if (!proposal) {
    return res.status(404).json({ error: 'Proposal not found' });
  }
  
  if (proposal.stage !== ProposalStage.PUBLIC_NOTICE) {
    return res.status(400).json({ error: 'Suggestions are only accepted during public notice period' });
  }
  
  if (!ownerId || !content) {
    return res.status(400).json({ error: 'Owner ID and content are required' });
  }
  
  const suggestion = voteService.addSuggestion(proposalId, ownerId, content);
  res.status(201).json(suggestion);
});

router.get('/proposals/:proposalId/suggestions', (req: Request, res: Response) => {
  const { proposalId } = req.params;
  
  const proposal = proposalService.getProposal(proposalId);
  
  if (!proposal) {
    return res.status(404).json({ error: 'Proposal not found' });
  }
  
  const suggestions = voteService.getSuggestions(proposalId);
  res.json(suggestions);
});

router.post('/proposals/:proposalId/objections', (req: Request, res: Response) => {
  const { proposalId } = req.params;
  const { ownerId, reason } = req.body;
  
  const proposal = proposalService.getProposal(proposalId);
  
  if (!proposal) {
    return res.status(404).json({ error: 'Proposal not found' });
  }
  
  if (proposal.stage !== ProposalStage.EXECUTION) {
    return res.status(400).json({ error: 'Objections are only accepted during execution period' });
  }
  
  if (!ownerId || !reason) {
    return res.status(400).json({ error: 'Owner ID and reason are required' });
  }
  
  try {
    const objection = voteService.addObjection(proposalId, ownerId, reason);
    
    const totalOwners = voteService.getTotalOwners();
    const objectionCount = voteService.getObjectionCount(proposalId);
    const objectionRate = totalOwners > 0 ? objectionCount / totalOwners : 0;
    
    if (objectionRate > OBJECTION_THRESHOLD) {
      proposalService.suspendExecution(proposalId);
    }
    
    res.status(201).json(objection);
  } catch (error) {
    res.status(409).json({ error: 'You have already submitted an objection for this proposal' });
  }
});

router.get('/proposals/:proposalId/objections', (req: Request, res: Response) => {
  const { proposalId } = req.params;
  
  const proposal = proposalService.getProposal(proposalId);
  
  if (!proposal) {
    return res.status(404).json({ error: 'Proposal not found' });
  }
  
  const objections = voteService.getObjections(proposalId);
  res.json(objections);
});

export default router;
