<template>
  <div>
    <PageHeader title="系统设置" />

    <el-tabs v-model="activeTab" type="border-card">
      <!-- 基本设置 -->
      <el-tab-pane label="基本设置" name="basic">
        <el-form :model="basicForm" label-width="120px" style="max-width:520px">
          <el-form-item label="通知天数">
            <el-select
              v-model="basicForm.notify_days"
              multiple
              placeholder="选择提前通知天数"
              style="width:100%"
            >
              <el-option label="提前 1 天" :value="1" />
              <el-option label="提前 3 天" :value="3" />
              <el-option label="提前 7 天" :value="7" />
              <el-option label="提前 14 天" :value="14" />
              <el-option label="提前 30 天" :value="30" />
            </el-select>
          </el-form-item>
          <el-form-item label="Ping 间隔(秒)">
            <el-input-number v-model="basicForm.ping_interval" :min="10" :max="3600" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="basicSaving" @click="saveBasic">保存设置</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <!-- Gmail OAuth2 -->
      <el-tab-pane label="Gmail OAuth2 配置" name="gmail">
        <div class="gmail-panel">
          <!-- Status indicator -->
          <div class="gmail-status">
            <span class="status-label">当前状态：</span>
            <el-tag :type="gmailConfigured ? 'success' : 'danger'">
              {{ gmailConfigured ? '✓ 已配置' : '✗ 未配置' }}
            </el-tag>
          </div>

          <!-- Step guide -->
          <el-card class="steps-card">
            <template #header><span>配置步骤</span></template>
            <el-steps direction="vertical" :active="4" size="small">
              <el-step title="前往 Google Cloud Console">
                <template #description>
                  访问
                  <a href="https://console.cloud.google.com" target="_blank" rel="noopener">
                    console.cloud.google.com
                  </a>
                  并登录您的 Google 账号
                </template>
              </el-step>
              <el-step title="创建项目并启用 Gmail API">
                <template #description>创建新项目 → 进入「API 和服务」→ 搜索并启用 Gmail API</template>
              </el-step>
              <el-step title="创建 OAuth2 凭据">
                <template #description>进入「凭据」→ 创建 OAuth2 客户端 ID → 应用类型选择「桌面应用」</template>
              </el-step>
              <el-step title="复制 Client ID 和 Client Secret">
                <template #description>创建完成后复制凭据填入下方</template>
              </el-step>
            </el-steps>
          </el-card>

          <!-- Credentials form -->
          <el-form :model="gmailForm" label-width="130px" style="max-width:560px;margin-top:20px">
            <el-form-item label="Client ID">
              <el-input v-model="gmailForm.client_id" placeholder="输入 OAuth2 Client ID" />
            </el-form-item>
            <el-form-item label="Client Secret">
              <el-input v-model="gmailForm.client_secret" type="password" placeholder="输入 OAuth2 Client Secret" show-password />
            </el-form-item>
            <el-form-item label="管理员邮箱">
              <el-input v-model="gmailForm.admin_email" placeholder="接收通知的 Gmail 邮箱" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :loading="gmailSaving" @click="saveGmailConfig">保存配置</el-button>
              <el-button @click="getAuthUrl" style="margin-left:10px">获取授权</el-button>
            </el-form-item>
          </el-form>

          <el-divider />

          <!-- Callback code -->
          <div style="max-width:560px">
            <p style="margin-bottom:12px;color:#606266;font-size:14px">
              点击「获取授权」后会在新标签页打开 Google 授权页面，授权完成后将页面中的授权码粘贴至此：
            </p>
            <el-form label-width="80px">
              <el-form-item label="授权码">
                <el-input v-model="callbackCode" placeholder="粘贴 Google 返回的授权码" />
              </el-form-item>
              <el-form-item>
                <el-button type="success" :loading="submitting" @click="submitCallback">提交授权码</el-button>
              </el-form-item>
            </el-form>
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { settingsApi } from '@/api'
import PageHeader from '@/components/PageHeader.vue'

const activeTab = ref('basic')
const basicSaving = ref(false)
const gmailSaving = ref(false)
const submitting = ref(false)
const gmailConfigured = ref(false)
const callbackCode = ref('')

const basicForm = reactive({
  notify_days: [1, 3, 7] as number[],
  ping_interval: 60,
})

const gmailForm = reactive({
  client_id: '',
  client_secret: '',
  admin_email: '',
})

async function loadSettings() {
  try {
    const res = await settingsApi.get()
    const d = res.data as Record<string, unknown>
    if (Array.isArray(d.notify_days)) basicForm.notify_days = d.notify_days as number[]
    if (d.ping_interval) basicForm.ping_interval = d.ping_interval as number
    if (d.gmail_client_id) gmailForm.client_id = d.gmail_client_id as string
    if (d.gmail_admin_email) gmailForm.admin_email = d.gmail_admin_email as string
    gmailConfigured.value = !!(d.gmail_configured || d.gmail_client_id)
  } catch { /* ignore */ }
}

async function saveBasic() {
  basicSaving.value = true
  try {
    await settingsApi.update({
      notify_days: basicForm.notify_days,
      ping_interval: basicForm.ping_interval,
    })
    ElMessage.success('保存成功')
  } catch {
    ElMessage.error('保存失败')
  } finally {
    basicSaving.value = false
  }
}

async function saveGmailConfig() {
  gmailSaving.value = true
  try {
    await settingsApi.update({
      gmail_client_id: gmailForm.client_id,
      gmail_client_secret: gmailForm.client_secret,
      gmail_admin_email: gmailForm.admin_email,
    })
    ElMessage.success('Gmail 配置保存成功')
    gmailConfigured.value = !!gmailForm.client_id
  } catch {
    ElMessage.error('保存失败')
  } finally {
    gmailSaving.value = false
  }
}

async function getAuthUrl() {
  try {
    const res = await settingsApi.getGmailAuthUrl()
    const d = res.data as { url?: string }
    if (d.url) {
      window.open(d.url, '_blank', 'noopener')
    } else {
      ElMessage.warning('未能获取授权链接，请先保存 Client ID 和 Secret')
    }
  } catch {
    ElMessage.error('获取授权链接失败')
  }
}

async function submitCallback() {
  if (!callbackCode.value.trim()) {
    ElMessage.warning('请输入授权码')
    return
  }
  submitting.value = true
  try {
    await settingsApi.submitGmailCallback(callbackCode.value.trim())
    ElMessage.success('授权成功！Gmail 已配置完毕')
    callbackCode.value = ''
    gmailConfigured.value = true
  } catch {
    ElMessage.error('授权失败，请检查授权码是否正确')
  } finally {
    submitting.value = false
  }
}

onMounted(loadSettings)
</script>

<style scoped>
.gmail-panel {
  padding: 4px 0;
}
.gmail-status {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 20px;
}
.status-label { font-size: 14px; color: #606266; }
.steps-card { max-width: 560px; }
.steps-card :deep(.el-step__description) {
  font-size: 13px;
  color: #606266;
  line-height: 1.6;
}
.steps-card a { color: #409eff; }
</style>
