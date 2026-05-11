<template>
  <div class="exam-page" :class="{ 'exam-ended': isExamEnded }">
    <el-container>
      <el-header class="exam-header">
        <div class="header-content">
          <div class="exam-info">
            <h2>{{ examDetail?.examPaperTitle }}</h2>
            <el-tag type="info">总分: {{ examDetail?.totalScore }}</el-tag>
            <el-tag type="success">及格分: {{ examDetail?.passScore }}</el-tag>
          </div>
          <div class="timer-section">
            <div class="timer" :class="{ 'timer-warning': remainingSeconds < 300 }">
              <el-icon><Timer /></el-icon>
              <span>{{ formattedTime }}</span>
            </div>
          </div>
        </div>
      </el-header>

      <el-main class="exam-main">
        <div class="exam-content" v-loading="loading">
          <div class="question-area" v-if="currentQuestion && !isExamEnded">
            <div class="question-header">
              <span class="question-number">第 {{ currentIndex + 1 }} / {{ examDetail?.questions?.length }} 题</span>
              <el-tag :type="getQuestionTypeTagType(currentQuestion.questionType)">
                {{ getQuestionTypeText(currentQuestion.questionType) }}
              </el-tag>
              <el-tag type="warning">
                {{ currentQuestion.score }} 分
              </el-tag>
              <el-tag type="info" v-if="currentQuestion.difficulty">
                {{ getDifficultyText(currentQuestion.difficulty) }}
              </el-tag>
            </div>
            
            <div class="question-content">
              <p>{{ currentQuestion.content }}</p>
            </div>

            <div class="question-options" v-if="currentQuestion.questionType === 'SINGLE_CHOICE'">
              <el-radio-group 
                v-model="answers[currentQuestion.examQuestionId]"
                @change="saveAnswer(currentQuestion)"
              >
                <el-radio 
                  v-for="option in currentQuestion.options" 
                  :key="option.optionKey"
                  :value="option.optionKey"
                  class="option-item"
                >
                  <span class="option-key">{{ option.optionKey }}.</span>
                  <span class="option-content">{{ option.content }}</span>
                </el-radio>
              </el-radio-group>
            </div>

            <div class="question-options" v-else-if="currentQuestion.questionType === 'MULTIPLE_CHOICE'">
              <el-checkbox-group 
                v-model="multiAnswers[currentQuestion.examQuestionId]"
                @change="saveMultiAnswer(currentQuestion)"
              >
                <el-checkbox 
                  v-for="option in currentQuestion.options" 
                  :key="option.optionKey"
                  :value="option.optionKey"
                  class="option-item"
                >
                  <span class="option-key">{{ option.optionKey }}.</span>
                  <span class="option-content">{{ option.content }}</span>
                </el-checkbox>
              </el-checkbox-group>
            </div>

            <div class="question-options" v-else-if="currentQuestion.questionType === 'TRUE_FALSE'">
              <el-radio-group 
                v-model="answers[currentQuestion.examQuestionId]"
                @change="saveAnswer(currentQuestion)"
              >
                <el-radio 
                  v-for="option in currentQuestion.options" 
                  :key="option.optionKey"
                  :value="option.optionKey"
                  class="option-item"
                >
                  <span class="option-key">{{ option.optionKey }}.</span>
                  <span class="option-content">{{ option.content }}</span>
                </el-radio>
              </el-radio-group>
            </div>

            <div class="question-fill" v-else-if="currentQuestion.questionType === 'FILL_BLANK'">
              <el-input
                v-model="answers[currentQuestion.examQuestionId]"
                type="textarea"
                :rows="3"
                placeholder="请输入答案"
                @blur="saveAnswer(currentQuestion)"
              />
            </div>

            <div class="question-navigation">
              <el-button 
                :disabled="currentIndex === 0"
                @click="previousQuestion"
              >
                上一题
              </el-button>
              <el-button 
                type="primary"
                :disabled="currentIndex === (examDetail?.questions?.length || 1) - 1"
                @click="nextQuestion"
              >
                下一题
              </el-button>
              <el-button 
                type="danger"
                @click="confirmSubmit"
              >
                交卷
              </el-button>
            </div>
          </div>

          <div class="exam-result" v-if="isExamEnded">
            <el-result
              :icon="examDetail?.obtainedScore >= examDetail?.passScore ? 'success' : 'error'"
              :title="examDetail?.obtainedScore >= examDetail?.passScore ? '恭喜！考试通过' : '考试未通过'"
              :sub-title="`总分: ${examDetail?.totalScore}, 得分: ${examDetail?.obtainedScore}, 及格分: ${examDetail?.passScore}`"
            >
              <template #extra>
                <el-button type="primary" @click="goBack">
                  返回列表
                </el-button>
                <el-button @click="viewReport">
                  查看详细报告
                </el-button>
              </template>
            </el-result>
          </div>
        </div>

        <div class="question-navigator">
          <h3>题目导航</h3>
          <div class="nav-grid">
            <el-button
              v-for="(q, index) in examDetail?.questions"
              :key="q.examQuestionId"
              :type="getNavButtonType(index, q.examQuestionId)"
              size="small"
              @click="goToQuestion(index)"
            >
              {{ index + 1 }}
            </el-button>
          </div>
          <div class="nav-legend">
            <div class="legend-item">
              <span class="legend-dot answered"></span>
              <span>已答</span>
            </div>
            <div class="legend-item">
              <span class="legend-dot current"></span>
              <span>当前</span>
            </div>
            <div class="legend-item">
              <span class="legend-dot unanswered"></span>
              <span>未答</span>
            </div>
          </div>
        </div>
      </el-main>
    </el-container>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessageBox, ElMessage } from 'element-plus'
import { Timer } from '@element-plus/icons-vue'
import api from '../utils/api'

const router = useRouter()
const route = useRoute()

const examRecordId = route.params.examRecordId
const examDetail = ref(null)
const currentIndex = ref(0)
const answers = ref({})
const multiAnswers = ref({})
const loading = ref(false)
const remainingSeconds = ref(0)
const timer = ref(null)
const isExamEnded = ref(false)

const currentQuestion = computed(() => {
  if (examDetail.value?.questions && examDetail.value.questions.length > currentIndex.value) {
    return examDetail.value.questions[currentIndex.value]
  }
  return null
})

const formattedTime = computed(() => {
  const minutes = Math.floor(remainingSeconds.value / 60)
  const seconds = remainingSeconds.value % 60
  return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
})

function getQuestionTypeText(type) {
  const typeMap = {
    'SINGLE_CHOICE': '单选题',
    'MULTIPLE_CHOICE': '多选题',
    'TRUE_FALSE': '判断题',
    'FILL_BLANK': '填空题'
  }
  return typeMap[type] || type
}

function getQuestionTypeTagType(type) {
  const typeMap = {
    'SINGLE_CHOICE': 'primary',
    'MULTIPLE_CHOICE': 'warning',
    'TRUE_FALSE': 'success',
    'FILL_BLANK': 'danger'
  }
  return typeMap[type] || 'info'
}

function getDifficultyText(difficulty) {
  const diffMap = {
    'LEVEL_1': '1星',
    'LEVEL_2': '2星',
    'LEVEL_3': '3星',
    'LEVEL_4': '4星',
    'LEVEL_5': '5星'
  }
  return diffMap[difficulty] || difficulty
}

function getNavButtonType(index, questionId) {
  if (index === currentIndex.value) {
    return 'primary'
  }
  if (isAnswered(questionId)) {
    return 'success'
  }
  return 'default'
}

function isAnswered(questionId) {
  const answer = answers.value[questionId]
  const multiAnswer = multiAnswers.value[questionId]
  
  if (multiAnswer && multiAnswer.length > 0) {
    return true
  }
  if (answer && answer.trim() !== '') {
    return true
  }
  return false
}

async function loadExamDetail() {
  loading.value = true
  try {
    const response = await api.get(`/exams/${examRecordId}`)
    examDetail.value = response.data
    remainingSeconds.value = response.data.remainingSeconds || 0
    
    if (response.data.status === 'SUBMITTED' || 
        response.data.status === 'TIMEOUT_SUBMITTED' ||
        response.data.status === 'ABNORMAL_SUBMITTED') {
      isExamEnded.value = true
      stopTimer()
    } else {
      startTimer()
    }
  } catch (error) {
    ElMessage.error('加载考试详情失败: ' + error.message)
  } finally {
    loading.value = false
  }
}

function startTimer() {
  timer.value = setInterval(async () => {
    if (remainingSeconds.value > 0) {
      remainingSeconds.value--
    } else {
      stopTimer()
      await autoSubmit()
    }
  }, 1000)
}

function stopTimer() {
  if (timer.value) {
    clearInterval(timer.value)
    timer.value = null
  }
}

async function autoSubmit() {
  try {
    ElMessage.warning('考试时间到，自动交卷')
    const response = await api.post(`/exams/${examRecordId}/auto-submit`)
    examDetail.value = response.data
    isExamEnded.value = true
  } catch (error) {
    ElMessage.error('自动交卷失败: ' + error.message)
  }
}

async function saveAnswer(question) {
  try {
    await api.post('/exams/answer', {
      examRecordId,
      examQuestionId: question.examQuestionId,
      userAnswer: answers.value[question.examQuestionId] || ''
    })
  } catch (error) {
    console.error('保存答案失败:', error)
  }
}

async function saveMultiAnswer(question) {
  try {
    const answer = multiAnswers.value[question.examQuestionId]?.join(',') || ''
    await api.post('/exams/answer', {
      examRecordId,
      examQuestionId: question.examQuestionId,
      userAnswer: answer
    })
  } catch (error) {
    console.error('保存答案失败:', error)
  }
}

function previousQuestion() {
  if (currentIndex.value > 0) {
    currentIndex.value--
  }
}

function nextQuestion() {
  if (currentIndex.value < examDetail.value.questions.length - 1) {
    currentIndex.value++
  }
}

function goToQuestion(index) {
  currentIndex.value = index
}

async function confirmSubmit() {
  const unanswered = examDetail.value.questions.filter(
    q => !isAnswered(q.examQuestionId)
  ).length

  let message = '确认要提交试卷吗？'
  if (unanswered > 0) {
    message = `您还有 ${unanswered} 道题未作答，确认要提交吗？`
  }

  try {
    await ElMessageBox.confirm(message, '提交确认', {
      confirmButtonText: '确认提交',
      cancelButtonText: '继续答题',
      type: 'warning'
    })
    
    await submitExam()
  } catch {
    // 用户取消
  }
}

async function submitExam() {
  try {
    stopTimer()
    const response = await api.post(`/exams/${examRecordId}/submit`)
    examDetail.value = response.data
    isExamEnded.value = true
    ElMessage.success('交卷成功')
  } catch (error) {
    ElMessage.error('交卷失败: ' + error.message)
    startTimer()
  }
}

function goBack() {
  router.push('/exam-list')
}

function viewReport() {
  router.push(`/report/${examRecordId}`)
}

async function reportSwitchOut() {
  try {
    await api.post('/exams/switch-out', {
      examRecordId,
      durationSeconds: 1
    })
  } catch (error) {
    console.error('报告切出失败:', error)
  }
}

function handleVisibilityChange() {
  if (document.hidden && !isExamEnded.value) {
    reportSwitchOut()
  }
}

onMounted(() => {
  loadExamDetail()
  document.addEventListener('visibilitychange', handleVisibilityChange)
})

onUnmounted(() => {
  stopTimer()
  document.removeEventListener('visibilitychange', handleVisibilityChange)
})
</script>

<style scoped>
.exam-page {
  min-height: 100vh;
  background-color: #f5f7fa;
}

.exam-header {
  background-color: #fff;
  border-bottom: 1px solid #ebeef5;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 100%;
}

.exam-info {
  display: flex;
  align-items: center;
  gap: 15px;
}

.exam-info h2 {
  margin: 0;
  font-size: 20px;
}

.timer-section {
  display: flex;
  align-items: center;
}

.timer {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 24px;
  font-weight: bold;
  color: #409eff;
}

.timer-warning {
  color: #f56c6c;
  animation: blink 1s infinite;
}

@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.exam-main {
  display: flex;
  gap: 20px;
  padding: 20px;
}

.exam-content {
  flex: 1;
  background: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 12px rgba(0,0,0,0.1);
}

.question-area {
  max-width: 800px;
  margin: 0 auto;
}

.question-header {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
  align-items: center;
}

.question-number {
  font-weight: bold;
  font-size: 16px;
}

.question-content {
  margin-bottom: 30px;
  padding: 20px;
  background: #f5f7fa;
  border-radius: 8px;
  font-size: 16px;
  line-height: 1.8;
}

.question-options {
  margin-bottom: 30px;
}

.option-item {
  display: block !important;
  margin-bottom: 15px !important;
  padding: 15px;
  background: #fafafa;
  border-radius: 6px;
  transition: all 0.3s;
}

.option-item:hover {
  background: #ecf5ff;
}

.option-key {
  font-weight: bold;
  margin-right: 8px;
}

.option-content {
  vertical-align: middle;
}

.question-fill {
  margin-bottom: 30px;
}

.question-navigation {
  display: flex;
  gap: 15px;
  justify-content: center;
  padding-top: 20px;
  border-top: 1px solid #ebeef5;
}

.question-navigator {
  width: 280px;
  background: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 12px rgba(0,0,0,0.1);
  max-height: calc(100vh - 140px);
  overflow-y: auto;
}

.question-navigator h3 {
  margin: 0 0 15px 0;
  font-size: 16px;
}

.nav-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 8px;
  margin-bottom: 20px;
}

.nav-legend {
  display: flex;
  gap: 15px;
  justify-content: center;
  padding-top: 15px;
  border-top: 1px solid #ebeef5;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;
}

.legend-dot {
  width: 20px;
  height: 20px;
  border-radius: 4px;
}

.legend-dot.answered {
  background: #67c23a;
}

.legend-dot.current {
  background: #409eff;
}

.legend-dot.unanswered {
  background: #e4e7ed;
}

.exam-result {
  padding: 40px;
}

.exam-ended .exam-header {
  background: #f0f2f5;
}
</style>
