<template>
  <div class="login-page">
    <div class="login-card">
      <div class="login-logo">
        <el-icon class="logo-icon"><Monitor /></el-icon>
        <h1>客户管理系统</h1>
        <p class="login-subtitle">直播客户管理平台</p>
      </div>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-position="top"
        @submit.prevent="handleLogin"
      >
        <el-form-item label="用户名" prop="username">
          <el-input
            v-model="form.username"
            placeholder="请输入用户名"
            size="large"
            prefix-icon="User"
            clearable
          />
        </el-form-item>

        <el-form-item label="密码" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="请输入密码"
            size="large"
            prefix-icon="Lock"
            show-password
            @keyup.enter="handleLogin"
          />
        </el-form-item>

        <el-alert
          v-if="error"
          :title="error"
          type="error"
          :closable="false"
          style="margin-bottom: 16px"
        />

        <el-button
          type="primary"
          size="large"
          style="width: 100%"
          :loading="loading"
          @click="handleLogin"
        >
          登 录
        </el-button>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import type { FormInstance, FormRules } from 'element-plus'

const router = useRouter()
const auth = useAuthStore()

const formRef = ref<FormInstance>()
const loading = ref(false)
const error = ref('')

const form = reactive({
  username: '',
  password: '',
})

const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

async function handleLogin() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    loading.value = true
    error.value = ''
    try {
      await auth.login(form.username, form.password)
      router.push('/dashboard')
    } catch (e: unknown) {
      const err = e as { response?: { data?: { message?: string } } }
      error.value = err?.response?.data?.message ?? '登录失败，请检查用户名和密码'
    } finally {
      loading.value = false
    }
  })
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  background: linear-gradient(135deg, #1a1d2b 0%, #2d3561 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.login-card {
  background: #fff;
  border-radius: 12px;
  padding: 40px;
  width: 100%;
  max-width: 420px;
  box-shadow: 0 20px 60px rgba(0,0,0,0.3);
}

.login-logo {
  text-align: center;
  margin-bottom: 32px;
}

.logo-icon {
  font-size: 48px;
  color: #409eff;
  display: block;
  margin: 0 auto 12px;
}

.login-logo h1 {
  font-size: 24px;
  font-weight: 700;
  color: #303133;
  margin-bottom: 4px;
}

.login-subtitle {
  font-size: 13px;
  color: #909399;
}
</style>
