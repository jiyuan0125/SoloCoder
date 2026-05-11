<template>
  <div class="report-page">
    <el-container>
      <el-header class="header">
        <div class="header-content">
          <h1>考试报告</h1>
          <el-button @click="goBack">返回列表</el-button>
        </div>
      </el-header>

      <el-main class="main-content">
        <div v-loading="loading">
          <el-card class="summary-card" v-if="report">
            <template #header>
              <div class="card-header">
                <span>考试概况</span>
                <el-tag :type="report.passed ? 'success' : 'danger'" size="large">
                  {{ report.passed ? '及格' : '不及格' }}
                </el-tag>
              </div>
            </template>
            
            <div class="summary-grid">
              <div class="summary-item">
                <div class="summary-label">考试名称</div>
                <div class="summary-value">{{ report.examPaperTitle }}</div>
              </div>
              <div class="summary-item">
                <div class="summary-label">考试状态</div>
                <div class="summary-value">
                  <el-tag :type="getStatusTagType(report.status)">
                    {{ getStatusText(report.status) }}
                  </el-tag>
                </div>
              </div>
              <div class="summary-item">
                <div class="summary-label">总分</div>
                <div class="summary-value score">
                  <span class="current-score">{{ report.obtainedScore }}</span>
                  <span class="total-score"> / {{ report.totalScore }}</span>
                </div>
              </div>
              <div class="summary-item">
                <div class="summary-label">及格分</div>
                <div class="summary-value">{{ report.passScore }}</div>
              </div>
              <div class="summary-item">
                <div class="summary-label">得分率</div>
                <div class="summary-value percentage" :class="report.passed ? 'pass' : 'fail'">
                  {{ report.scorePercentage.toFixed(1) }}%
                </div>
              </div>
              <div class="summary-item">
                <div class="summary-label">考试时长</div>
                <div class="summary-value">{{ report.durationMinutes + report.extendedMinutes }} 分钟</div>
              </div>
              <div class="summary-item">
                <div class="summary-label">实际用时</div>
                <div class="summary-value">
                  {{ report.actualDurationSeconds ? formatDuration(report.actualDurationSeconds) : '-' }}
                </div>
              </div>
              <div class="summary-item">
                <div class="summary-label">切出次数</div>
                <div class="summary-value">
                  <el-tag :type="report.switchCount > 3 ? 'danger' : 'info'">
                    {{ report.switchCount }} 次
                  </el-tag>
                </div>
              </div>
            </div>
          </el-card>

          <el-card class="stats-card" v-if="report">
            <template #header>
              <span>答题统计</span>
            </template>
            
            <div class="stats-grid">
              <div class="stat-item">
                <el-statistic title="总题数" :value="report.totalQuestions" />
              </div>
              <div class="stat-item">
                <el-statistic title="已答题数" :value="report.answeredQuestions" />
              </div>
              <div class="stat-item correct">
                <el-statistic title="正确题数" :value="report.correctQuestions" />
              </div>
              <div class="stat-item wrong">
                <el-statistic title="错误题数" :value="report.incorrectQuestions" />
              </div>
            </div>
          </el-card>

          <el-card class="category-card" v-if="report && report.categoryScores.length > 0">
            <template #header>
              <span>知识点分析</span>
            </template>
            
            <el-table :data="report.categoryScores" border>
              <el-table-column prop="categoryName" label="知识点分类" width="200" />
              <el-table-column prop="totalQuestions" label="总题数" width="80" align="center" />
              <el-table-column prop="correctCount" label="正确" width="80" align="center">
                <template #default="scope">
                  <span class="correct">{{ scope.row.correctCount }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="incorrectCount" label="错误" width="80" align="center">
                <template #default="scope">
                  <span class="wrong">{{ scope.row.incorrectCount }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="unansweredCount" label="未答" width="80" align="center">
                <template #default="scope">
                  <span class="unanswered">{{ scope.row.unansweredCount }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="totalScore" label="总分" width="80" align="center" />
              <el-table-column prop="obtainedScore" label="得分" width="80" align="center">
                <template #default="scope">
                  <strong>{{ scope.row.obtainedScore }}</strong>
                </template>
              </el-table-column>
              <el-table-column label="得分率" width="150" align="center">
                <template #default="scope">
                  <el-progress 
                    :percentage="scope.row.scoreRate" 
                    :color="getProgressColor(scope.row.scoreRate)"
                    :stroke-width="18"
                  />
                </template>
              </el-table-column>
            </el-table>

            <div class="weak-areas" v-if="!report.passed">
              <el-alert 
                title="薄弱知识点提示" 
                type="warning" 
                :closable="false"
                show-icon
              >
                <p>以下知识点得分率较低，建议重点复习：</p>
                <ul>
                  <li v-for="cat in weakCategories" :key="cat.categoryName">
                    <strong>{{ cat.categoryName }}</strong> - 得分率: {{ cat.scoreRate.toFixed(1) }}%
                  </li>
                </ul>
              </el-alert>
            </div>
          </el-card>

          <el-card class="exceptions-card" v-if="report && report.exceptions.length > 0">
            <template #header>
              <span>异常记录 
                <el-tag type="danger" size="small" style="margin-left: 10px;">
                  {{ report.exceptions.length }} 条记录
                </el-tag>
              </span>
            </template>
            
            <el-timeline>
              <el-timeline-item
                v-for="(exception, index) in report.exceptions"
                :key="index"
                :timestamp="formatDateTime(exception.timestamp)"
                placement="top"
                :type="getExceptionType(exception.type)"
              >
                <el-card class="exception-card">
                  <h4>{{ getExceptionTitle(exception.type) }}</h4>
                  <p>{{ exception.description }}</p>
                  <p v-if="exception.durationSeconds" class="duration">
                    持续时间: {{ exception.durationSeconds }} 秒
                  </p>
                </el-card>
              </el-timeline-item>
            </el-timeline>
          </el-card>
        </div>
      </el-main>
    </el-container>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import api from '../utils/api'
import { ElMessage } from 'element-plus'

const router = useRouter()
const route = useRoute()

const examRecordId = route.params.examRecordId
const report = ref(null)
const loading = ref(false)

const weakCategories = computed(() => {
  if (!report.value || !report.value.categoryScores) return []
  return report.value.categoryScores.filter(c => c.scoreRate < 60)
})

function getStatusText(status) {
  const statusMap = {
    'NOT_STARTED': '未开始',
    'IN_PROGRESS': '进行中',
    'PAUSED': '已暂停',
    'SUBMITTED': '已提交',
    'TIMEOUT_SUBMITTED': '超时交卷',
    'ABNORMAL_SUBMITTED': '异常交卷'
  }
  return statusMap[status] || status
}

function getStatusTagType(status) {
  const typeMap = {
    'NOT_STARTED': 'info',
    'IN_PROGRESS': 'primary',
    'PAUSED': 'warning',
    'SUBMITTED': 'success',
    'TIMEOUT_SUBMITTED': 'warning',
    'ABNORMAL_SUBMITTED': 'danger'
  }
  return typeMap[status] || 'info'
}

function getProgressColor(percentage) {
  if (percentage >= 80) return '#67c23a'
  if (percentage >= 60) return '#e6a23c'
  return '#f56c6c'
}

function getExceptionType(type) {
  const typeMap = {
    'SINGLE_QUESTION_TIMEOUT': 'warning',
    'PAGE_SWITCH_OUT': 'info',
    'SWITCH_COUNT_EXCEEDED': 'danger',
    'SWITCH_DURATION_EXCEEDED': 'danger'
  }
  return typeMap[type] || 'info'
}

function getExceptionTitle(type) {
  const titleMap = {
    'SINGLE_QUESTION_TIMEOUT': '单题答题超时',
    'PAGE_SWITCH_OUT': '切出考试页面',
    'SWITCH_COUNT_EXCEEDED': '切出次数超限',
    'SWITCH_DURATION_EXCEEDED': '单次切出超时'
  }
  return titleMap[type] || type
}

function formatDuration(seconds) {
  const minutes = Math.floor(seconds / 60)
  const remainingSeconds = seconds % 60
  return `${minutes}分${remainingSeconds}秒`
}

function formatDateTime(datetime) {
  if (!datetime) return '-'
  return new Date(datetime).toLocaleString('zh-CN')
}

async function loadReport() {
  loading.value = true
  try {
    const response = await api.get(`/exams/${examRecordId}/report`)
    report.value = response.data
  } catch (error) {
    ElMessage.error('加载报告失败: ' + error.message)
  } finally {
    loading.value = false
  }
}

function goBack() {
  router.push('/exam-list')
}

onMounted(() => {
  loadReport()
})
</script>

<style scoped>
.report-page {
  min-height: 100vh;
}

.header {
  background-color: #409eff;
  color: white;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 100%;
}

.header-content h1 {
  margin: 0;
  font-size: 20px;
}

.main-content {
  padding: 20px;
}

.summary-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
}

.summary-item {
  padding: 15px;
  background: #f5f7fa;
  border-radius: 8px;
}

.summary-label {
  font-size: 14px;
  color: #909399;
  margin-bottom: 8px;
}

.summary-value {
  font-size: 18px;
  font-weight: bold;
}

.summary-value.score .current-score {
  font-size: 28px;
  color: #409eff;
}

.summary-value.score .total-score {
  font-size: 16px;
  color: #909399;
}

.summary-value.percentage {
  font-size: 24px;
}

.summary-value.percentage.pass {
  color: #67c23a;
}

.summary-value.percentage.fail {
  color: #f56c6c;
}

.stats-card {
  margin-bottom: 20px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
}

.stat-item {
  text-align: center;
  padding: 20px;
  background: #f5f7fa;
  border-radius: 8px;
}

.stat-item.correct {
  background: #f0f9eb;
}

.stat-item.wrong {
  background: #fef0f0;
}

.category-card {
  margin-bottom: 20px;
}

.correct {
  color: #67c23a;
  font-weight: bold;
}

.wrong {
  color: #f56c6c;
  font-weight: bold;
}

.unanswered {
  color: #909399;
}

.weak-areas {
  margin-top: 20px;
}

.weak-areas ul {
  margin: 10px 0 0 20px;
}

.weak-areas li {
  margin: 5px 0;
}

.exceptions-card {
  margin-bottom: 20px;
}

.exception-card {
  box-shadow: none !important;
}

.exception-card h4 {
  margin: 0 0 10px 0;
}

.exception-card p {
  margin: 5px 0;
  color: #606266;
}

.duration {
  color: #e6a23c !important;
  font-weight: bold;
}
</style>
