import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('board_token'))
  const username = ref<string>(localStorage.getItem('board_username') || '')

  const isLoggedIn = computed(() => !!token.value)

  async function login(user: string, password: string) {
    const res = await authApi.login({ username: user, password })
    const t = res.data.token
    token.value = t
    username.value = user
    localStorage.setItem('board_token', t)
    localStorage.setItem('board_username', user)
  }

  function logout() {
    token.value = null
    username.value = ''
    localStorage.removeItem('board_token')
    localStorage.removeItem('board_username')
  }

  return { token, username, isLoggedIn, login, logout }
})
