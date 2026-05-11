<template>
  <div class="admin-page">
    <el-container>
      <el-header class="header">
        <div class="header-content">
          <h1>管理后台</h1>
          <div class="header-actions">
            <el-tag :type="userRoleTagType">{{ userRoleText }}</el-tag>
            <el-button @click="goBack">返回考试列表</el-button>
          </div>
        </div>
      </el-header>

      <el-main class="main-content">
        <el-tabs v-model="activeTab" class="admin-tabs">
          <el-tab-pane label="题库管理" name="questions">
            <el-card>
              <template #header>
                <div class="card-header">
                  <span>题目列表</span>
                  <el-button type="primary" @click="openQuestionDialog()">
                    新增题目
                  </el-button>
                </div>
              </template>
              
              <el-table :data="questions" border>
                <el-table-column prop="id" label="ID" width="80" />
                <el-table-column prop="content" label="题目内容" min-width="300" show-overflow-tooltip />
                <el-table-column label="类型" width="100">
                  <template #default="scope">
                    <el-tag size="small">{{ getQuestionTypeText(scope.row.type) }}</el-tag>
                  </template>
                </el-table-column>
                <el-table-column label="分类" width="120">
                  <template #default="scope">
                    {{ scope.row.category?.name || '-' }}
                  </template>
                </el-table-column>
                <el-table-column label="难度" width="100">
                  <template #default="scope">
                    {{ getDifficultyText(scope.row.difficulty) }}
                  </template>
                </el-table-column>
                <el-table-column prop="defaultScore" label="分值" width="80" />
                <el-table-column label="操作" width="150">
                  <template #default="scope">
                    <el-button size="small" type="primary" @click="openQuestionDialog(scope.row)">编辑</el-button>
                    <el-button size="small" type="danger" @click="deleteQuestion(scope.row)">删除</el-button>
                  </template>
                </el-table-column>
              </el-table>
            </el-card>
          </el-tab-pane>

          <el-tab-pane label="分类管理" name="categories">
            <el-card>
              <template #header>
                <div class="card-header">
                  <span>知识点分类</span>
                  <el-button type="primary" @click="openCategoryDialog()">
                    新增分类
                  </el-button>
                </div>
              </template>
              
              <el-table :data="categories" border>
                <el-table-column prop="id" label="ID" width="80" />
                <el-table-column prop="name" label="分类名称" width="200" />
                <el-table-column prop="description" label="描述" />
                <el-table-column prop="sortOrder" label="排序" width="80" />
                <el-table-column label="操作" width="150">
                  <template #default="scope">
                    <el-button size="small" type="primary" @click="openCategoryDialog(scope.row)">编辑</el-button>
                    <el-button size="small" type="danger" @click="deleteCategory(scope.row)">删除</el-button>
                  </template>
                </el-table-column>
              </el-table>
            </el-card>
          </el-tab-pane>

          <el-tab-pane label="试卷管理" name="papers">
            <el-card>
              <template #header>
                <div class="card-header">
                  <span>试卷列表</span>
                  <el-button type="primary" @click="openPaperDialog()">
                    新增试卷
                  </el-button>
                </div>
              </template>
              
              <el-table :data="examPapers" border>
                <el-table-column prop="id" label="ID" width="80" />
                <el-table-column prop="title" label="试卷名称" width="200" />
                <el-table-column prop="description" label="描述" />
                <el-table-column prop="totalScore" label="总分" width="80" />
                <el-table-column prop="passScore" label="及格分" width="80" />
                <el-table-column prop="durationMinutes" label="时长(分)" width="100" />
                <el-table-column label="状态" width="100">
                  <template #default="scope">
                    <el-tag :type="scope.row.enabled ? 'success' : 'info'" size="small">
                      {{ scope.row.enabled ? '启用' : '禁用' }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column label="操作" width="150">
                  <template #default="scope">
                    <el-button size="small" type="primary" @click="openPaperDialog(scope.row)">编辑</el-button>
                    <el-button size="small" type="danger" @click="deletePaper(scope.row)">删除</el-button>
                  </template>
                </el-table-column>
              </el-table>
            </el-card>
          </el-tab-pane>

          <el-tab-pane label="考试记录" name="records">
            <el-card>
              <template #header>
                <span>所有考试记录</span>
              </template>
              
              <el-table :data="allRecords" border>
                <el-table-column prop="id" label="记录ID" width="100" />
                <el-table-column label="试卷" width="200">
                  <template #default="scope">
                    {{ scope.row.examPaper?.title || '-' }}
                  </template>
                </el-table-column>
                <el-table-column label="考生" width="120">
                  <template #default="scope">
                    {{ scope.row.user?.realName || '-' }}
                  </template>
                </el-table-column>
                <el-table-column prop="attemptNumber" label="第几次" width="80" />
                <el-table-column label="状态" width="120">
                  <template #default="scope">
                    <el-tag :type="getStatusTagType(scope.row.status)" size="small">
                      {{ getStatusText(scope.row.status) }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="obtainedScore" label="得分" width="80" />
                <el-table-column label="操作" width="200">
                  <template #default="scope">
                    <el-button 
                      size="small" 
                      type="warning"
                      v-if="scope.row.status === 'IN_PROGRESS' || scope.row.status === 'PAUSED'"
                      @click="extendTime(scope.row)"
                    >
                      延长时间
                    </el-button>
                    <el-button 
                      size="small" 
                      type="info"
                      v-if="scope.row.status !== 'NOT_STARTED' && scope.row.status !== 'IN_PROGRESS' && scope.row.status !== 'PAUSED'"
                      @click="viewReport(scope.row.id)"
                    >
                      查看报告
                    </el-button>
                  </template>
                </el-table-column>
              </el-table>
            </el-card>
          </el-tab-pane>
        </el-tabs>
      </el-main>
    </el-container>

    <el-dialog
      v-model="categoryDialogVisible"
      title="分类管理"
      width="500px"
    >
      <el-form :model="categoryForm" label-width="100px">
        <el-form-item label="分类名称">
          <el-input v-model="categoryForm.name" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="categoryForm.description" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="categoryForm.sortOrder" :min="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="categoryDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveCategory">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../utils/api'
import { useUserStore } from '../stores/user'

const router = useRouter()
const userStore = useUserStore()

const activeTab = ref('questions')
const currentUser = computed(() => userStore.currentUser)

const questions = ref([])
const categories = ref([])
const examPapers = ref([])
const allRecords = ref([])

const categoryDialogVisible = ref(false)
const categoryForm = ref({
  id: null,
  name: '',
  description: '',
  sortOrder: 0
})

const userRoleText = computed(() => {
  const roleMap = {
    'ADMIN': '管理员',
    'TEACHER': '教师',
    'STUDENT': '学生'
  }
  return roleMap[currentUser.value?.role] || '未知'
})

const userRoleTagType = computed(() => {
  const typeMap = {
    'ADMIN': 'danger',
    'TEACHER': 'warning',
    'STUDENT': 'success'
  }
  return typeMap[currentUser.value?.role] || 'info'
})

function getQuestionTypeText(type) {
  const typeMap = {
    'SINGLE_CHOICE': '单选',
    'MULTIPLE_CHOICE': '多选',
    'TRUE_FALSE': '判断',
    'FILL_BLANK': '填空'
  }
  return typeMap[type] || type
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

async function loadQuestions() {
  try {
    const response = await api.get('/questions')
    questions.value = response.data || []
  } catch (error) {
    console.error('加载题目失败:', error)
  }
}

async function loadCategories() {
  try {
    const response = await api.get('/categories')
    categories.value = response.data || []
  } catch (error) {
    console.error('加载分类失败:', error)
  }
}

async function loadExamPapers() {
  try {
    const response = await api.get('/exam-papers')
    examPapers.value = response.data || []
  } catch (error) {
    console.error('加载试卷失败:', error)
  }
}

async function loadAllRecords() {
  try {
    const paperPromises = examPapers.value.map(paper => 
      api.get(`/exams/paper/${paper.id}/records`)
    )
    const responses = await Promise.all(paperPromises)
    allRecords.value = responses.flatMap(r => r.data || [])
  } catch (error) {
    console.error('加载考试记录失败:', error)
  }
}

function openQuestionDialog(question = null) {
  ElMessage.info('题目编辑功能开发中')
}

function openCategoryDialog(category = null) {
  if (category) {
    categoryForm.value = { ...category }
  } else {
    categoryForm.value = {
      id: null,
      name: '',
      description: '',
      sortOrder: 0
    }
  }
  categoryDialogVisible.value = true
}

async function saveCategory() {
  try {
    if (categoryForm.value.id) {
      await api.put(`/categories/${categoryForm.value.id}`, categoryForm.value)
      ElMessage.success('更新成功')
    } else {
      await api.post('/categories', categoryForm.value)
      ElMessage.success('创建成功')
    }
    categoryDialogVisible.value = false
    await loadCategories()
  } catch (error) {
    ElMessage.error('保存失败: ' + error.message)
  }
}

async function deleteQuestion(question) {
  try {
    await ElMessageBox.confirm(`确定要删除题目 "${question.content.substring(0, 20)}..." 吗？`, '确认', {
      type: 'warning'
    })
    await api.delete(`/questions/${question.id}`)
    ElMessage.success('删除成功')
    await loadQuestions()
  } catch {
    // 用户取消
  }
}

async function deleteCategory(category) {
  try {
    await ElMessageBox.confirm(`确定要删除分类 "${category.name}" 吗？`, '确认', {
      type: 'warning'
    })
    await api.delete(`/categories/${category.id}`)
    ElMessage.success('删除成功')
    await loadCategories()
  } catch (error) {
    ElMessage.error('删除失败: ' + error.message)
  }
}

function openPaperDialog(paper = null) {
  ElMessage.info('试卷编辑功能开发中')
}

async function deletePaper(paper) {
  try {
    await ElMessageBox.confirm(`确定要删除试卷 "${paper.title}" 吗？`, '确认', {
      type: 'warning'
    })
    await api.delete(`/exam-papers/${paper.id}`)
    ElMessage.success('删除成功')
    await loadExamPapers()
  } catch {
    // 用户取消
  }
}

async function extendTime(record) {
  try {
    const { value } = await ElMessageBox.prompt('请输入延长的分钟数', '延长考试时间', {
      confirmButtonText: '确认',
      cancelButtonText: '取消',
      inputPattern: /^\d+$/,
      inputErrorMessage: '请输入有效的数字'
    })
    
    const minutes = parseInt(value)
    await api.post(`/exams/${record.id}/extend?additionalMinutes=${minutes}`)
    ElMessage.success(`已延长 ${minutes} 分钟`)
  } catch {
    // 用户取消
  }
}

function viewReport(recordId) {
  router.push(`/report/${recordId}`)
}

function goBack() {
  router.push('/exam-list')
}

onMounted(async () => {
  await loadQuestions()
  await loadCategories()
  await loadExamPapers()
  await loadAllRecords()
})
</script>

<style scoped>
.admin-page {
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

.header-actions {
  display: flex;
  gap: 15px;
  align-items: center;
}

.main-content {
  padding: 20px;
}

.admin-tabs {
  background: white;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0,0,0,0.1);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
