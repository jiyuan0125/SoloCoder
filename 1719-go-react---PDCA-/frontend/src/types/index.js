export const INDICATOR_CATEGORIES = [
  { value: 'safety', label: '医疗安全指标' },
  { value: 'efficiency', label: '医疗效率指标' },
  { value: 'quality', label: '医疗质量指标' },
  { value: 'experience', label: '患者体验指标' },
]

export const PDCA_STATUS = [
  { value: 'planning', label: '计划中', color: 'status-info' },
  { value: 'executing', label: '执行中', color: 'status-warning' },
  { value: 'checking', label: '检查中', color: 'status-warning' },
  { value: 'completed', label: '已完成', color: 'status-success' },
  { value: 'closed', label: '已关闭', color: 'status-success' },
]

export const PDCA_PHASES = [
  { value: 'plan', label: 'Plan', fullLabel: '计划阶段' },
  { value: 'do', label: 'Do', fullLabel: '执行阶段' },
  { value: 'check', label: 'Check', fullLabel: '检查阶段' },
  { value: 'act', label: 'Act', fullLabel: '处理阶段' },
]

export const TODO_TYPES = [
  { value: 'improvement', label: '改进待办', color: 'status-danger' },
  { value: 'action', label: '行动项', color: 'status-info' },
]

export const TODO_STATUS = [
  { value: 'pending', label: '待处理', color: 'status-warning' },
  { value: 'in_progress', label: '进行中', color: 'status-info' },
  { value: 'completed', label: '已完成', color: 'status-success' },
  { value: 'overdue', label: '已逾期', color: 'status-danger' },
  { value: 'archived', label: '已归档', color: '' },
]

export const getStatusLabel = (status, statusList) => {
  const item = statusList.find(s => s.value === status)
  return item ? item.label : status
}

export const getStatusColor = (status, statusList) => {
  const item = statusList.find(s => s.value === status)
  return item ? item.color : ''
}
