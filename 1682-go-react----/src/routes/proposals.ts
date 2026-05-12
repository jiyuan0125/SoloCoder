import { Router, Request, Response } from 'express';
import { proposalService } from '../services/proposalService';
import { todoService } from '../services/todoService';
import { TodoType, ReviewResult, ProposalStage } from '../types';
import { now } from '../utils';

const router = Router();

router.post('/', (req: Request, res: Response) => {
  const { title, content, type, attachmentDescription, createdBy } = req.body;
  
  if (!title || title.trim() === '') {
    return res.status(400).json({ error: 'Title cannot be empty' });
  }
  
  if (!content || !type || attachmentDescription === undefined) {
    return res.status(400).json({ error: 'Missing required fields' });
  }
  
  const proposal = proposalService.submitProposal(
    title.trim(),
    content,
    type,
    attachmentDescription,
    createdBy || 'anonymous'
  );
  
  todoService.createTodo(proposal.id, TodoType.REVIEW);
  
  res.status(201).json(proposal);
});

router.get('/', (req: Request, res: Response) => {
  const proposals = proposalService.listProposals();
  res.json(proposals);
});

router.get('/:id', (req: Request, res: Response) => {
  const proposal = proposalService.getProposal(req.params.id);
  
  if (!proposal) {
    return res.status(404).json({ error: 'Proposal not found' });
  }
  
  res.json(proposal);
});

router.post('/:id/review/start', (req: Request, res: Response) => {
  const proposal = proposalService.getProposal(req.params.id);
  
  if (!proposal) {
    return res.status(404).json({ error: 'Proposal not found' });
  }
  
  if (proposal.stage !== ProposalStage.SUBMITTED) {
    return res.status(400).json({ error: 'Invalid stage transition' });
  }
  
  const updated = proposalService.startReview(req.params.id);
  res.json(updated);
});

router.post('/:id/review/complete', (req: Request, res: Response) => {
  const { result, comments } = req.body;
  
  const proposal = proposalService.getProposal(req.params.id);
  
  if (!proposal) {
    return res.status(404).json({ error: 'Proposal not found' });
  }
  
  if (proposal.stage !== ProposalStage.REVIEW) {
    return res.status(400).json({ error: 'Invalid stage transition' });
  }
  
  if (!result || (result !== ReviewResult.PASS && result !== ReviewResult.REJECT)) {
    return res.status(400).json({ error: 'Invalid review result' });
  }
  
  if (result === ReviewResult.PASS) {
    const isComplete = proposal.content && proposal.type;
    if (!isComplete) {
      return res.status(400).json({ error: 'Content is incomplete' });
    }
    
    const isDuplicate = proposalService.isDuplicate(
      proposal.title,
      proposal.content,
      proposal.id
    );
    if (isDuplicate) {
      return res.status(400).json({ error: 'Duplicate proposal' });
    }
  }
  
  const updated = proposalService.completeReview(req.params.id, result);
  
  const reviewTodos = todoService.getTodosByProposal(req.params.id)
    .filter(t => t.type === TodoType.REVIEW && t.status === 'pending');
  
  for (const todo of reviewTodos) {
    todoService.completeTodo(todo.id);
  }
  
  if (result === ReviewResult.PASS) {
    todoService.createTodo(req.params.id, TodoType.VOTING_ADMIN);
  }
  
  res.json(updated);
});

router.post('/:id/voting/start', (req: Request, res: Response) => {
  const proposal = proposalService.getProposal(req.params.id);
  
  if (!proposal) {
    return res.status(404).json({ error: 'Proposal not found' });
  }
  
  if (proposal.stage !== ProposalStage.PUBLIC_NOTICE) {
    return res.status(400).json({ error: 'Invalid stage transition' });
  }
  
  const updated = proposalService.startVoting(req.params.id);
  res.json(updated);
});

router.post('/:id/voting/complete', (req: Request, res: Response) => {
  const { isApproved } = req.body;
  
  const proposal = proposalService.getProposal(req.params.id);
  
  if (!proposal) {
    return res.status(404).json({ error: 'Proposal not found' });
  }
  
  if (proposal.stage !== ProposalStage.VOTING) {
    return res.status(400).json({ error: 'Invalid stage transition' });
  }
  
  if (isApproved === undefined) {
    return res.status(400).json({ error: 'Missing isApproved field' });
  }
  
  const updated = proposalService.completeVoting(req.params.id, isApproved);
  
  const votingTodos = todoService.getTodosByProposal(req.params.id)
    .filter(t => t.type === TodoType.VOTING_ADMIN && t.status === 'pending');
  
  for (const todo of votingTodos) {
    todoService.completeTodo(todo.id);
  }
  
  res.json(updated);
});

router.post('/:id/execution/complete', (req: Request, res: Response) => {
  const proposal = proposalService.getProposal(req.params.id);
  
  if (!proposal) {
    return res.status(404).json({ error: 'Proposal not found' });
  }
  
  if (proposal.stage !== ProposalStage.EXECUTION) {
    return res.status(400).json({ error: 'Invalid stage transition' });
  }
  
  const updated = proposalService.completeExecution(req.params.id);
  res.json(updated);
});

router.post('/:id/rerun', (req: Request, res: Response) => {
  const original = proposalService.getProposal(req.params.id);
  
  if (!original) {
    return res.status(404).json({ error: 'Proposal not found' });
  }
  
  if (original.stage !== ProposalStage.VOTING && 
      original.stage !== ProposalStage.REJECTED) {
    return res.status(400).json({ error: 'Cannot rerun this proposal' });
  }
  
  const rerun = proposalService.createRerunProposal(req.params.id);
  if (rerun) {
    todoService.createTodo(rerun.id, TodoType.REVIEW);
  }
  
  res.json(rerun);
});

export default router;
