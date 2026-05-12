import { Router, Request, Response } from 'express';
import {
  createIssue,
  getIssueById,
  listIssues,
  isPublicityPeriodEnded,
  updateIssueStatus,
  checkStatusTransition
} from '../services/issueService';
import { addOpinion, listOpinionsByIssue, canAddOpinion } from '../services/opinionService';
import {
  createVoteSetting,
  getVoteSettingByIssueId,
  isVotingPeriod,
  isVotingEnded,
  castVote,
  hasVoted,
  calculateWeight,
  calculateVoteResult,
  canViewResults
} from '../services/voteService';
import { createTodo, getTodoTypeForStatus, markTodoCompleted } from '../services/todoService';
import { listUsers } from '../services/userService';
import { authenticate, requireVerified, requireCommittee, requireExecutor } from '../middleware/auth';
import { IssueStatus, VoteType, VoteOption } from '../types';
import { sendError } from '../utils';

const router = Router();

router.post('/', authenticate, requireVerified, (req: Request, res: Response) => {
  try {
    const { title, content, category, attachments } = req.body;

    if (!title || title.trim() === '') {
      return sendError(res, { code: 400, message: '标题不能为空' });
    }

    if (!content || content.trim() === '') {
      return sendError(res, { code: 400, message: '内容不能为空' });
    }

    if (!category || category.trim() === '') {
      return sendError(res, { code: 400, message: '类别不能为空' });
    }

    const attachmentList = Array.isArray(attachments) ? attachments : [];

    const issue = createIssue(
      title.trim(),
      content.trim(),
      category.trim(),
      attachmentList,
      req.user.id
    );

    res.status(201).json(issue);
  } catch (error) {
    sendError(res, { code: 500, message: '创建议题失败' });
  }
});

router.get('/', authenticate, (req: Request, res: Response) => {
  try {
    const issues = listIssues();
    res.json(issues);
  } catch (error) {
    sendError(res, { code: 500, message: '获取议题列表失败' });
  }
});

router.get('/:id', authenticate, (req: Request, res: Response) => {
  try {
    const id = parseInt(req.params.id, 10);
    const issue = getIssueById(id);

    if (!issue) {
      return sendError(res, { code: 404, message: '议题不存在' });
    }

    res.json(issue);
  } catch (error) {
    sendError(res, { code: 500, message: '获取议题失败' });
  }
});

router.post('/:id/opinions', authenticate, requireVerified, (req: Request, res: Response) => {
  try {
    const id = parseInt(req.params.id, 10);
    const { content } = req.body;
    const issue = getIssueById(id);

    if (!issue) {
      return sendError(res, { code: 404, message: '议题不存在' });
    }

    if (!canAddOpinion(issue)) {
      return sendError(res, { code: 400, message: '当前阶段不能发表意见' });
    }

    if (!content || content.trim() === '') {
      return sendError(res, { code: 400, message: '意见内容不能为空' });
    }

    const opinion = addOpinion(id, req.user.id, content.trim());
    res.status(201).json(opinion);
  } catch (error) {
    sendError(res, { code: 500, message: '发表意见失败' });
  }
});

router.get('/:id/opinions', authenticate, (req: Request, res: Response) => {
  try {
    const id = parseInt(req.params.id, 10);
    const opinions = listOpinionsByIssue(id);
    res.json(opinions);
  } catch (error) {
    sendError(res, { code: 500, message: '获取意见列表失败' });
  }
});

router.post('/:id/committee-confirm', authenticate, requireCommittee, (req: Request, res: Response) => {
  try {
    const id = parseInt(req.params.id, 10);
    const issue = getIssueById(id);

    if (!issue) {
      return sendError(res, { code: 404, message: '议题不存在' });
    }

    if (!isPublicityPeriodEnded(issue)) {
      return sendError(res, { code: 400, message: '公示期未结束，不能提前投票' });
    }

    if (!checkStatusTransition(issue.status, IssueStatus.WAITING_VOTE_CONFIRM)) {
      return sendError(res, { code: 400, message: '不能跳步流转状态' });
    }

    const updatedIssue = updateIssueStatus(id, IssueStatus.WAITING_VOTE_CONFIRM);

    const users = listUsers();
    const committeeUsers = users.filter(u => u.role === 'committee');
    for (const user of committeeUsers) {
      createTodo(id, user.id, 'committee_vote_confirm', 3);
    }

    res.json(updatedIssue);
  } catch (error) {
    sendError(res, { code: 500, message: '确认失败' });
  }
});

router.post('/:id/vote-settings', authenticate, requireCommittee, (req: Request, res: Response) => {
  try {
    const id = parseInt(req.params.id, 10);
    const { startAt, endAt, voteType, minParticipationRate } = req.body;
    const issue = getIssueById(id);

    if (!issue) {
      return sendError(res, { code: 404, message: '议题不存在' });
    }

    if (!checkStatusTransition(issue.status, IssueStatus.VOTING)) {
      return sendError(res, { code: 400, message: '不能跳步流转状态' });
    }

    if (!startAt || !endAt) {
      return sendError(res, { code: 400, message: '投票起止时间不能为空' });
    }

    if (!voteType || !Object.values(VoteType).includes(voteType)) {
      return sendError(res, { code: 400, message: '无效的投票方式' });
    }

    if (minParticipationRate == null || minParticipationRate <= 0) {
      return sendError(res, { code: 400, message: '最低参与率必须大于0' });
    }

    if (minParticipationRate > 1) {
      return sendError(res, { code: 400, message: '最低参与率不能超过100%' });
    }

    const setting = createVoteSetting(id, startAt, endAt, voteType, minParticipationRate);
    updateIssueStatus(id, IssueStatus.VOTING);

    res.status(201).json(setting);
  } catch (error) {
    sendError(res, { code: 500, message: '设置投票失败' });
  }
});

router.get('/:id/vote-settings', authenticate, (req: Request, res: Response) => {
  try {
    const id = parseInt(req.params.id, 10);
    const setting = getVoteSettingByIssueId(id);
    res.json(setting);
  } catch (error) {
    sendError(res, { code: 500, message: '获取投票设置失败' });
  }
});

router.post('/:id/votes', authenticate, requireVerified, (req: Request, res: Response) => {
  try {
    const id = parseInt(req.params.id, 10);
    const { option } = req.body;
    const issue = getIssueById(id);
    const setting = getVoteSettingByIssueId(id);

    if (!issue) {
      return sendError(res, { code: 404, message: '议题不存在' });
    }

    if (!setting) {
      return sendError(res, { code: 400, message: '投票未设置' });
    }

    if (!isVotingPeriod(setting)) {
      return sendError(res, { code: 400, message: '不在投票时间段内' });
    }

    if (hasVoted(id, req.user.id)) {
      return sendError(res, { code: 409, message: '不能重复投票' });
    }

    if (!option || !Object.values(VoteOption).includes(option)) {
      return sendError(res, { code: 400, message: '无效的投票选项' });
    }

    if (setting.voteType === VoteType.BY_AREA) {
      const area = req.user.houseArea;
      if (area == null || area <= 0) {
        return sendError(res, { code: 400, message: '房屋面积不能为零或负数' });
      }
    }

    const weight = calculateWeight(req.user, setting.voteType);
    const vote = castVote(id, req.user.id, option, weight);

    res.status(201).json(vote);
  } catch (error: any) {
    if (error.code === 'SQLITE_CONSTRAINT_UNIQUE') {
      return sendError(res, { code: 409, message: '不能重复投票' });
    }
    sendError(res, { code: 500, message: '投票失败' });
  }
});

router.get('/:id/results', authenticate, (req: Request, res: Response) => {
  try {
    const id = parseInt(req.params.id, 10);
    const issue = getIssueById(id);
    const setting = getVoteSettingByIssueId(id);

    if (!issue) {
      return sendError(res, { code: 404, message: '议题不存在' });
    }

    if (!canViewResults(issue, setting)) {
      return sendError(res, { code: 403, message: '投票未结束，不能查询结果' });
    }

    const result = calculateVoteResult(id);

    if (issue.status === IssueStatus.VOTING && isVotingEnded(setting!)) {
      updateIssueStatus(id, IssueStatus.VOTED);
    }

    res.json(result);
  } catch (error) {
    sendError(res, { code: 500, message: '获取投票结果失败' });
  }
});

router.post('/:id/executor-confirm', authenticate, requireExecutor, (req: Request, res: Response) => {
  try {
    const id = parseInt(req.params.id, 10);
    const issue = getIssueById(id);

    if (!issue) {
      return sendError(res, { code: 404, message: '议题不存在' });
    }

    if (issue.status !== IssueStatus.VOTED) {
      return sendError(res, { code: 400, message: '当前状态不能执行确认' });
    }

    const result = calculateVoteResult(id);

    if (!result.isPassed) {
      return sendError(res, { code: 400, message: '投票未通过' });
    }

    if (!checkStatusTransition(issue.status, IssueStatus.WAITING_EXEC_CONFIRM)) {
      return sendError(res, { code: 400, message: '不能跳步流转状态' });
    }

    const updatedIssue = updateIssueStatus(id, IssueStatus.WAITING_EXEC_CONFIRM);

    const users = listUsers();
    const executorUsers = users.filter(u => u.role === 'executor');
    for (const user of executorUsers) {
      createTodo(id, user.id, 'executor_confirm', 3);
    }

    res.json(updatedIssue);
  } catch (error) {
    sendError(res, { code: 500, message: '确认失败' });
  }
});

router.post('/:id/start-execute', authenticate, requireExecutor, (req: Request, res: Response) => {
  try {
    const id = parseInt(req.params.id, 10);
    const issue = getIssueById(id);

    if (!issue) {
      return sendError(res, { code: 404, message: '议题不存在' });
    }

    if (!checkStatusTransition(issue.status, IssueStatus.EXECUTING)) {
      return sendError(res, { code: 400, message: '不能跳步流转状态' });
    }

    const updatedIssue = updateIssueStatus(id, IssueStatus.EXECUTING);
    createTodo(id, req.user.id, 'execute_task', 7);

    res.json(updatedIssue);
  } catch (error) {
    sendError(res, { code: 500, message: '开始执行失败' });
  }
});

router.post('/:id/complete', authenticate, requireExecutor, (req: Request, res: Response) => {
  try {
    const id = parseInt(req.params.id, 10);
    const issue = getIssueById(id);

    if (!issue) {
      return sendError(res, { code: 404, message: '议题不存在' });
    }

    if (!checkStatusTransition(issue.status, IssueStatus.COMPLETED)) {
      return sendError(res, { code: 400, message: '不能跳步流转状态' });
    }

    const updatedIssue = updateIssueStatus(id, IssueStatus.COMPLETED);

    const todos = require('../services/todoService');
    const userTodos = todos.listTodosByUser(req.user.id);
    const issueTodo = userTodos.find((t: any) => t.issueId === id && t.type === 'execute_task');
    if (issueTodo) {
      markTodoCompleted(issueTodo.id);
    }

    res.json(updatedIssue);
  } catch (error) {
    sendError(res, { code: 500, message: '完成确认失败' });
  }
});

export { router as issuesRouter };
