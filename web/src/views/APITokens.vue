<template>
  <div>
    <PageHeader title="API Token">
      <el-button type="primary" :icon="Plus" @click="openCreate">创建 Token</el-button>
    </PageHeader>

    <el-card>
      <el-table v-loading="loading" :data="list" style="width:100%">
        <el-table-column prop="name" label="名称" min-width="140" />
        <el-table-column label="Token" min-width="200">
          <template #default="{ row }">
            <div class="token-cell">
              <span class="token-text">{{ revealedId === row.id ? (row.token as string) : maskToken(row.token as string) }}</span>
              <el-button
                size="small"
                text
                @click="revealedId = revealedId === row.id ? null : row.id as number"
              >
                <el-icon><View /></el-icon>
              </el-button>
              <el-button size="small" text @click="copyText(row.token as string)">
                <el-icon><CopyDocument /></el-icon>
              </el-button>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="最后使用" width="160">
          <template #default="{ row }">{{ formatDate(row.last_used_at as string) }}</template>
        </el-table-column>
        <el-table-column label="创建时间" width="160">
          <template #default="{ row }">{{ formatDate(row.created_at as string) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="80" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="danger" text @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Create dialog -->
    <el-dialog v-model="dialogVisible" title="创建 API Token" width="420px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="70px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="Token 用途说明" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleCreate">创建</el-button>
      </template>
    </el-dialog>

    <!-- New token display (only shown once) -->
    <el-dialog v-model="newTokenVisible" title="Token 创建成功" width="460px" :close-on-click-modal="false">
      <el-alert type="warning" :closable="false" style="margin-bottom:16px">
        <template #title>请立即复制保存！此 Token 不会再次显示。</template>
      </el-alert>
      <div class="new-token-box">
        <code>{{ newToken }}</code>
        <el-button :icon="CopyDocument" @click="copyText(newToken)">复制</el-button>
      </div>
      <template #footer>
        <el-button type="primary" @click="newTokenVisible = false">我已保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, View, CopyDocument } from '@element-plus/icons-vue'
import { tokensApi } from '@/api'
import { formatDate } from '@/utils'
import PageHeader from '@/components/PageHeader.vue'
import type { FormInstance, FormRules } from 'element-plus'

const loading = ref(false)
const saving = ref(false)
const list = ref<Record<string, unknown>[]>([])
const dialogVisible = ref(false)
const newTokenVisible = ref(false)
const newToken = ref('')
const revealedId = ref<number | null>(null)
const formRef = ref<FormInstance>()

const form = reactive({ name: '' })
const rules: FormRules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
}

function maskToken(token: string): string {
  if (!token) return '••••••••'
  return token.slice(0, 8) + '••••••••••••••••' + token.slice(-4)
}

async function copyText(text: string) {
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.warning('复制失败，请手动复制')
  }
}

async function loadData() {
  loading.value = true
  try {
    const res = await tokensApi.list()
    const d = res.data as { list?: unknown[]; items?: unknown[] } | unknown[]
    list.value = (Array.isArray(d) ? d : ((d as { list?: unknown[] }).list ?? [])) as Record<string, unknown>[]
  } catch { list.value = [] } finally { loading.value = false }
}

function openCreate() {
  Object.assign(form, { name: '' })
  dialogVisible.value = true
}

async function handleCreate() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    saving.value = true
    try {
      const res = await tokensApi.create({ name: form.name })
      const d = res.data as { token?: string }
      newToken.value = d.token ?? ''
      dialogVisible.value = false
      newTokenVisible.value = true
      loadData()
    } catch {
      ElMessage.error('创建失败')
    } finally {
      saving.value = false
    }
  })
}

async function handleDelete(row: Record<string, unknown>) {
  await ElMessageBox.confirm(`确认删除 Token「${row.name}」？`, '提示', { type: 'warning' })
  try {
    await tokensApi.delete(row.id as number)
    ElMessage.success('已删除')
    loadData()
  } catch {
    ElMessage.error('删除失败')
  }
}

onMounted(loadData)
</script>

<style scoped>
.token-cell {
  display: flex;
  align-items: center;
  gap: 4px;
  font-family: monospace;
}
.token-text {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px;
}
.new-token-box {
  display: flex;
  align-items: center;
  gap: 10px;
  background: #f5f7fa;
  border-radius: 6px;
  padding: 12px 14px;
}
.new-token-box code {
  flex: 1;
  word-break: break-all;
  font-size: 13px;
  color: #409eff;
}
</style>
