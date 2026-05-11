<template>
  <div class="exam-list">
    <el-container>
      <el-header class="header">
        <div class="header-content">
          <h1>在线考试系统</h1>
          <div class="user-info">
            <el-tag type="info">{{ currentUser.realName }}</el-tag>
            <el-tag :type="userRoleTagType">{{ userRoleText }}</el-tag>
          </div>
        </div>
      </el-header>
      
      <el-main class="main-content">
        <el-card class="welcome-card">
          <template #header>
            <span>欢迎使用在线考试系统</span>
          </template>
          <p>请选择要参加的考试或查看历史记录。</p>
        </el-card>

        <el-card class="section-card">
          <template #header>
            <span>可用考试</span>
          </template>
          <el-table :data="examPapers" v-loading="loading" border>
            <el-table-column prop="title" label="考试名称" width="250" />
            <el-table-column prop="description" label="描述" />
            <el-table-column prop="totalScore" label="总分" width="80" />
            <el-table-column prop="passScore" label="及格分" width="80" />
            <el-table-column prop="durationMinutes" label="时长(分钟)" width="100" />
            <el-table-column label="操作" width="200">
              <template #default="scope">
                <el-button 
                  type="primary" 
                  size="small"
                  @click="startExam(scope.row.id)"
                >
                  开始考试
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>

        <el-card class="section-card" v-if="currentUser.id">
          <template #header>
            <span>我的考试记录</span>
          </template>
          <el-table :data="examRecords" border>
            <el-table-column prop="examPaper.title" label="考试名称" width="200" />
            <el-table-column prop="attemptNumber" label="第几次" width="80" />
            <el-table-column label="状态" width="100">
              <template #default="scope">
                <el-tag :type="getStatusTagType(scope.row.status)">
                  {{ getStatusText(scope.row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="obtainedScore" label="得分" width="80" />
            <el-table-column label="操作" width="150">
              <template #default="scope">
                <el-button 
                  type="info" 
                  size="small"
                  v-if="scope.row.status === 'IN_PROGRESS' || scope.row.status === 'PAUSED'"
                  @click="resumeExam(scope.row.id)"
                >
                  继续考试
                </el-button>
                <el-button 
                  type="primary" 
                  size="small"
                  v-else-if="scope.row.status !== 'NOT_STARTED'"
                  @click="viewReport(scope.row.id)"
                >
                  查看报告
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>

        <el-card class="section-card" v-if="currentUser.role === 'ADMIN' || currentUser.role === 'TEACHER'">
          <template #header>
            <span>管理功能</span>
          </template>
          <el-button type="primary" @click="goToAdmin">
            进入管理后台
          </el-button>
        </el-card>
      </el-main>
    </el-container>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import api from '../utils/api'
import { ElMessage } from 'element-plus'

const router = useRouter()
const userStore = useUserStore()

const currentUser = computed(() => userStore.currentUser)
const examPapers = ref([])
const examRecords = ref([])
const loading = ref(false)

const userRoleText = computed(() => {
  const roleMap = {
    'ADMIN': '管理员',
    'TEACHER': '教师',
    'STUDENT': '学生'
  }
  return roleMap[currentUser.value.role] || '未知'
})

const userRoleTagType = computed(() => {
  const typeMap = {
    'ADMIN': 'danger',
    'TEACHER': 'warning',
    'STUDENT': 'success'
  }
  return typeMap[currentUser.value.role] || 'info'
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

async function loadExamPapers() {
  loading.value = true
  try {
    const response = await api.get('/exam-papers')
    examPapers.value = response.data || []
  } catch (error) {
    console.error('加载考试列表失败:', error)
  } finally {
    loading.value = false
  }
}

async function loadExamRecords() {
  try {
    if (currentUser.value.id) {
      const response = await api.get(`/exams/user/${currentUser.value.id}/records`)
      examRecords.value = response.data || []
    }
  } catch (error) {
    console.error('加载考试记录失败:', error)
  }
}

async function startExam(examPaperId) {
  try {
    const response = await api.post('/exams/start', {
      examPaperId,
      userId: currentUser.value.id
    })
    router.push(`/exam/${response.data.examRecordId}`)
  } catch (error) {
    ElMessage.error('开始考试失败: ' + error.message)
  }
}

function resumeExam(examRecordId) {
  router.push(`/exam/${examRecordId}`)
}

function viewReport(examRecordId) {
  router.push(`/report/${examRecordId}`)
}

function goToAdmin() {
  router.push('/admin')
}

onMounted(() => {
  loadExamPapers()
  loadExamRecords()
})
</script>

<style scoped>
.exam-list {
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
  font-size: 24px;
}

.user-info {
  display: flex;
  gap: 10px;
}

.main-content {
  padding: 20px;
}

.welcome-card {
  margin-bottom: 20px;
}

.section-card {
  margin-bottom: 20px;
}
</style>
