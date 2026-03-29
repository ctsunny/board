<template>
  <div>
    <PageHeader title="直播线路">
      <el-button type="primary" :icon="Plus" @click="openCreate">新增线路</el-button>
    </PageHeader>

    <!-- Filter -->
    <el-card class="filter-card">
      <el-row :gutter="12">
        <el-col :xs="12" :sm="6">
          <el-select v-model="filters.region_id" placeholder="按地区" clearable @change="loadData" style="width:100%">
            <el-option v-for="r in regions" :key="r.id as number" :label="r.name as string" :value="r.id" />
          </el-select>
        </el-col>
        <el-col :xs="12" :sm="6">
          <el-select v-model="filters.server_id" placeholder="按服务器" clearable @change="loadData" style="width:100%">
            <el-option v-for="s in servers" :key="s.id as number" :label="s.name as string" :value="s.id" />
          </el-select>
        </el-col>
        <el-col :xs="12" :sm="6">
          <el-select v-model="filters.status" placeholder="状态" clearable @change="loadData" style="width:100%">
            <el-option label="启用" value="enabled" />
            <el-option label="禁用" value="disabled" />
          </el-select>
        </el-col>
      </el-row>
    </el-card>

    <el-card style="margin-top:16px">
      <el-table v-loading="loading" :data="list" style="width:100%">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="name" label="名称" min-width="120" />
        <el-table-column prop="url" label="URL" min-width="180" show-overflow-tooltip />
        <el-table-column label="协议" width="90">
          <template #default="{ row }">
            <el-tag :type="protocolType(row.protocol as string)" size="small">{{ row.protocol }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="region_name" label="地区" width="110" />
        <el-table-column prop="server_name" label="服务器" width="120" />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <StatusBadge :status="row.status || 'unknown'" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="130" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" text @click="openEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" text @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editRow ? '编辑线路' : '新增线路'" width="480px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="URL" prop="url">
          <el-input v-model="form.url" placeholder="rtmp://..." />
        </el-form-item>
        <el-form-item label="协议" prop="protocol">
          <el-select v-model="form.protocol" style="width:100%">
            <el-option label="RTMP" value="RTMP" />
            <el-option label="HLS" value="HLS" />
            <el-option label="RTSP" value="RTSP" />
            <el-option label="SRT" value="SRT" />
          </el-select>
        </el-form-item>
        <el-form-item label="地区">
          <el-select v-model="form.region_id" clearable style="width:100%">
            <el-option v-for="r in regions" :key="r.id as number" :label="r.name as string" :value="r.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="服务器">
          <el-select v-model="form.server_id" clearable style="width:100%">
            <el-option v-for="s in servers" :key="s.id as number" :label="s.name as string" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status" style="width:100%">
            <el-option label="启用" value="enabled" />
            <el-option label="禁用" value="disabled" />
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
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { routesApi, regionsApi, serversApi } from '@/api'
import PageHeader from '@/components/PageHeader.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import type { FormInstance, FormRules } from 'element-plus'

type ProtocolType = '' | 'success' | 'warning' | 'info' | 'danger'

function protocolType(p: string): ProtocolType {
  const m: Record<string, ProtocolType> = { RTMP: '', HLS: 'success', RTSP: 'warning', SRT: 'info' }
  return m[p] ?? ''
}

const loading = ref(false)
const saving = ref(false)
const list = ref<Record<string, unknown>[]>([])
const regions = ref<Record<string, unknown>[]>([])
const servers = ref<Record<string, unknown>[]>([])
const dialogVisible = ref(false)
const editRow = ref<Record<string, unknown> | null>(null)
const formRef = ref<FormInstance>()

const filters = reactive({ region_id: undefined as number | undefined, server_id: undefined as number | undefined, status: '' })
const form = reactive({ name: '', url: '', protocol: 'RTMP', region_id: undefined as number | undefined, server_id: undefined as number | undefined, status: 'enabled', remark: '' })

const rules: FormRules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  url: [{ required: true, message: '请输入URL', trigger: 'blur' }],
  protocol: [{ required: true, message: '请选择协议', trigger: 'change' }],
}

async function loadData() {
  loading.value = true
  try {
    const res = await routesApi.list({ ...filters })
    const d = res.data as { list?: unknown[]; items?: unknown[] } | unknown[]
    list.value = (Array.isArray(d) ? d : ((d as { list?: unknown[] }).list ?? [])) as Record<string, unknown>[]
  } catch { list.value = [] } finally { loading.value = false }
}

async function loadSelects() {
  try {
    const [rRes, sRes] = await Promise.all([regionsApi.list(), serversApi.list()])
    const rd = rRes.data as { list?: unknown[] } | unknown[]; regions.value = (Array.isArray(rd) ? rd : ((rd as { list?: unknown[] }).list ?? [])) as Record<string, unknown>[]
    const sd = sRes.data as { list?: unknown[] } | unknown[]; servers.value = (Array.isArray(sd) ? sd : ((sd as { list?: unknown[] }).list ?? [])) as Record<string, unknown>[]
  } catch { /* ignore */ }
}

function openCreate() {
  editRow.value = null
  Object.assign(form, { name: '', url: '', protocol: 'RTMP', region_id: undefined, server_id: undefined, status: 'enabled', remark: '' })
  dialogVisible.value = true
}

function openEdit(row: Record<string, unknown>) {
  editRow.value = row
  Object.assign(form, { name: row.name ?? '', url: row.url ?? '', protocol: row.protocol ?? 'RTMP', region_id: row.region_id, server_id: row.server_id, status: row.status ?? 'enabled', remark: row.remark ?? '' })
  dialogVisible.value = true
}

async function handleSave() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    saving.value = true
    try {
      if (editRow.value) { await routesApi.update(editRow.value.id as number, { ...form }) }
      else { await routesApi.create({ ...form }) }
      ElMessage.success('保存成功'); dialogVisible.value = false; loadData()
    } catch { ElMessage.error('保存失败') } finally { saving.value = false }
  })
}

async function handleDelete(row: Record<string, unknown>) {
  await ElMessageBox.confirm(`确认删除线路「${row.name}」？`, '提示', { type: 'warning' })
  try { await routesApi.delete(row.id as number); ElMessage.success('已删除'); loadData() }
  catch { ElMessage.error('删除失败') }
}

onMounted(() => { loadData(); loadSelects() })
</script>

<style scoped>
.filter-card :deep(.el-card__body) { padding: 14px 16px; }
</style>
