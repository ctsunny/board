<template>
  <div>
    <PageHeader title="服务器管理">
      <el-button type="primary" :icon="Plus" @click="openCreate">新增服务器</el-button>
    </PageHeader>

    <el-card>
      <el-table v-loading="loading" :data="list" style="width:100%">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="name" label="名称" min-width="120" />
        <el-table-column prop="ip" label="IP地址" width="140" />
        <el-table-column prop="location" label="地区" width="120" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <span class="status-dot" :class="row.status || 'unknown'" />
            <StatusBadge :status="row.status || 'unknown'" />
          </template>
        </el-table-column>
        <el-table-column label="延迟" width="90">
          <template #default="{ row }">
            <span v-if="row.latency != null" class="latency-badge">{{ row.latency }}ms</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="最后Ping" width="160">
          <template #default="{ row }">{{ formatDate(row.last_ping_at as string) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handlePing(row)" :loading="pingingId === row.id">Ping</el-button>
            <el-button size="small" type="primary" text @click="openEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" text @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editRow ? '编辑服务器' : '新增服务器'" width="460px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="70px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="IP" prop="ip">
          <el-input v-model="form.ip" placeholder="192.168.1.1" />
        </el-form-item>
        <el-form-item label="地区">
          <el-select v-model="form.location" clearable filterable placeholder="选择地区" style="width:100%">
            <el-option
              v-for="region in regionOptions"
              :key="region"
              :label="region"
              :value="region"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.remark" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { serversApi, regionsApi } from '@/api'
import { formatDate, getListData } from '@/utils'
import PageHeader from '@/components/PageHeader.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import type { FormInstance, FormRules } from 'element-plus'

const loading = ref(false)
const saving = ref(false)
const list = ref<Record<string, unknown>[]>([])
const regions = ref<Record<string, unknown>[]>([])
const dialogVisible = ref(false)
const editRow = ref<Record<string, unknown> | null>(null)
const pingingId = ref<number | null>(null)
const formRef = ref<FormInstance>()

const regionOptions = computed(() => {
  const options = regions.value
    .map((region) => String(region.name ?? '').trim())
    .filter(Boolean)
  const current = String(form.location ?? '').trim()
  return current && !options.includes(current) ? options.concat(current) : options
})

const form = reactive({ name: '', ip: '', location: '', remark: '' })
const rules: FormRules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  ip:   [{ required: true, message: '请输入IP地址', trigger: 'blur' }],
}

async function loadData() {
  loading.value = true
  try {
    const res = await serversApi.list()
    const d = res.data as { list?: unknown[]; items?: unknown[] } | unknown[]
    list.value = (Array.isArray(d) ? d : ((d as { list?: unknown[] }).list ?? [])) as Record<string, unknown>[]
  } catch {
    list.value = []
  } finally {
    loading.value = false
  }
}

async function loadRegions() {
  try {
    const res = await regionsApi.list()
    regions.value = getListData(res.data)
  } catch {
    regions.value = []
  }
}

async function handlePing(row: Record<string, unknown>) {
  pingingId.value = row.id as number
  try {
    const res = await serversApi.ping(row.id as number)
    ElMessage.success(`Ping 成功，延迟 ${(res.data as { latency?: number }).latency ?? '?'}ms`)
    loadData()
  } catch {
    ElMessage.error('Ping 失败')
  } finally {
    pingingId.value = null
  }
}

function openCreate() {
  editRow.value = null
  Object.assign(form, { name: '', ip: '', location: '', remark: '' })
  dialogVisible.value = true
}

function openEdit(row: Record<string, unknown>) {
  editRow.value = row
  Object.assign(form, { name: row.name ?? '', ip: row.ip ?? '', location: row.location ?? '', remark: row.remark ?? '' })
  dialogVisible.value = true
}

async function handleSave() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    saving.value = true
    try {
      if (editRow.value) {
        await serversApi.update(editRow.value.id as number, { ...form })
      } else {
        await serversApi.create({ ...form })
      }
      ElMessage.success('保存成功')
      dialogVisible.value = false
      loadData()
    } catch {
      ElMessage.error('保存失败')
    } finally {
      saving.value = false
    }
  })
}

async function handleDelete(row: Record<string, unknown>) {
  await ElMessageBox.confirm(`确认删除服务器「${row.name}」？`, '提示', { type: 'warning' })
  try {
    await serversApi.delete(row.id as number)
    ElMessage.success('已删除')
    loadData()
  } catch {
    ElMessage.error('删除失败')
  }
}

onMounted(() => {
  loadData()
  loadRegions()
})
</script>

<style scoped>
.status-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 6px;
  vertical-align: middle;
}
.status-dot.online  { background: #67c23a; }
.status-dot.offline { background: #f56c6c; }
.status-dot.unknown { background: #c0c4cc; }
.latency-badge {
  font-size: 12px;
  color: #67c23a;
  background: #f0f9eb;
  padding: 2px 8px;
  border-radius: 10px;
}
</style>
