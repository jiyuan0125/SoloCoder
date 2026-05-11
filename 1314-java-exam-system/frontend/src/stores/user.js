import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUserStore = defineStore('user', () => {
  const currentUser = ref({
    id: 3,
    username: 'student',
    realName: '李同学',
    role: 'STUDENT'
  })

  const users = ref([
    { id: 1, username: 'admin', realName: '系统管理员', role: 'ADMIN', password: 'admin123' },
    { id: 2, username: 'teacher', realName: '张老师', role: 'TEACHER', password: 'teacher123' },
    { id: 3, username: 'student', realName: '李同学', role: 'STUDENT', password: 'student123' }
  ])

  function login(username, password) {
    const user = users.value.find(u => u.username === username && u.password === password)
    if (user) {
      currentUser.value = { ...user }
      return true
    }
    return false
  }

  function logout() {
    currentUser.value = null
  }

  return {
    currentUser,
    users,
    login,
    logout
  }
})
