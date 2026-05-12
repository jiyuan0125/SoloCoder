const BASE_URL = '/api'

function getHeaders(role, unitID) {
  const headers = {
    'Content-Type': 'application/json',
  }
  if (role) {
    headers['X-Role'] = role
  }
  if (unitID) {
    headers['X-Unit-ID'] = unitID
  }
  return headers
}

export async function apiRequest(endpoint, options = {}, role, unitID) {
  const response = await fetch(`${BASE_URL}${endpoint}`, {
    ...options,
    headers: {
      ...getHeaders(role, unitID),
      ...options.headers,
    },
  })

  if (!response.ok) {
    const text = await response.text()
    throw new Error(`${response.status}: ${text}`)
  }

  const text = await response.text()
  if (!text) {
    return null
  }
  return JSON.parse(text)
}

export const api = {
  get: (endpoint, role, unitID) => apiRequest(endpoint, { method: 'GET' }, role, unitID),
  post: (endpoint, body, role, unitID) =>
    apiRequest(endpoint, { method: 'POST', body: JSON.stringify(body) }, role, unitID),
  patch: (endpoint, body, role, unitID) =>
    apiRequest(endpoint, { method: 'PATCH', body: JSON.stringify(body) }, role, unitID),
  put: (endpoint, body, role, unitID) =>
    apiRequest(endpoint, { method: 'PUT', body: JSON.stringify(body) }, role, unitID),
  delete: (endpoint, role, unitID) =>
    apiRequest(endpoint, { method: 'DELETE' }, role, unitID),
}

export const STATUS_LABELS = {
  received: '已接收',
  testing: '检测中',
  test_complete: '检测完成',
  report_generating: '报告生成中',
  reported: '已出报告',
}

export const SAMPLE_TYPE_LABELS = {
  saliva: '唾液',
  blood: '血液',
  swab: '口腔拭子',
  tissue: '组织',
}

export const REPORT_STATUS_LABELS = {
  draft: '草稿',
  pending: '待审核',
  published: '已发布',
}

export const APPROVAL_LABELS = {
  submitted: '已提交',
  first_review: '一审中',
  second_review: '二审中',
  final_review: '终审中',
  approved: '已通过',
  rejected: '已驳回',
}
