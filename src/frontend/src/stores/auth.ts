import { defineStore } from 'pinia'
import { ref } from 'vue'
import backend from '@/api/backend'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(localStorage.getItem('zq_token') || '')
  const username = ref<string>(localStorage.getItem('zq_username') || '')

  const isLoggedIn = () => !!token.value

  async function login(phone: string, password: string) {
    const res: any = await backend.post('/auth/login', { username: phone, password })
    token.value = res.token
    username.value = res.username || phone
    localStorage.setItem('zq_token', res.token)
    localStorage.setItem('zq_username', username.value)
  }

  function logout() {
    token.value = ''
    username.value = ''
    localStorage.removeItem('zq_token')
    localStorage.removeItem('zq_username')
  }

  return { token, username, isLoggedIn, login, logout }
})
