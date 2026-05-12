export const MEETING_STATUS = {
  PREPARING: 'preparing',
  ACCEPTING_SUBMISSIONS: 'accepting_submissions',
  REVIEWING: 'reviewing',
  NOTIFYING: 'notifying',
  ENDED: 'ended',
};

export const MEETING_STATUS_LABELS = {
  [MEETING_STATUS.PREPARING]: '筹备中',
  [MEETING_STATUS.ACCEPTING_SUBMISSIONS]: '征稿中',
  [MEETING_STATUS.REVIEWING]: '审稿中',
  [MEETING_STATUS.NOTIFYING]: '录用通知中',
  [MEETING_STATUS.ENDED]: '已结束',
};

export const PAPER_STATUS = {
  DRAFT: 'draft',
  SUBMITTED: 'submitted',
  FORMAT_CHECKING: 'format_checking',
  UNDER_REVIEW: 'under_review',
  REVISION: 'revision',
  ACCEPTED: 'accepted',
  REJECTED: 'rejected',
  PUBLISHED: 'published',
};

export const PAPER_STATUS_LABELS = {
  [PAPER_STATUS.DRAFT]: '草稿',
  [PAPER_STATUS.SUBMITTED]: '已提交',
  [PAPER_STATUS.FORMAT_CHECKING]: '格式审查中',
  [PAPER_STATUS.UNDER_REVIEW]: '审稿中',
  [PAPER_STATUS.REVISION]: '修改中',
  [PAPER_STATUS.ACCEPTED]: '已录用',
  [PAPER_STATUS.REJECTED]: '已拒稿',
  [PAPER_STATUS.PUBLISHED]: '已发表',
};

export const REVIEW_RECOMMENDATIONS = {
  STRONG_ACCEPT: 'strong_accept',
  WEAK_ACCEPT: 'weak_accept',
  BORDERLINE: 'borderline',
  WEAK_REJECT: 'weak_reject',
  STRONG_REJECT: 'strong_reject',
};

export const REVIEW_RECOMMENDATION_LABELS = {
  [REVIEW_RECOMMENDATIONS.STRONG_ACCEPT]: '强接收',
  [REVIEW_RECOMMENDATIONS.WEAK_ACCEPT]: '弱接收',
  [REVIEW_RECOMMENDATIONS.BORDERLINE]: '边界',
  [REVIEW_RECOMMENDATIONS.WEAK_REJECT]: '弱拒稿',
  [REVIEW_RECOMMENDATIONS.STRONG_REJECT]: '强拒稿',
};
