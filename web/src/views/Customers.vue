<template>
  <div>
    <PageHeader title="客户管理">
      <el-button type="primary" :icon="Plus" @click="openCreate">新增客户</el-button>
      <el-button
        type="danger"
        :icon="Delete"
        :disabled="!selection.length"
        @click="handleBatchDelete"
      >批量删除</el-button>
      <el-button
        type="warning"
        :icon="Refresh"
        :disabled="!selection.length"
        @click="batchRenewVisible = true"
      >批量续期</el-button>
      <el-button :icon="Download" @click="handleExport">导出CSV</el-button>
    </PageHeader>

    <!-- Filter bar -->
    <el-card class="filter-card">
      <el-row :gutter="12">
        <el-col :xs="24" :sm="8" :md="5">
          <el-input v-model="filters.name" placeholder="搜索姓名/联系方式" clearable @change="loadData" />
        </el-col>
        <el-col :xs="12" :sm="6" :md="4">
          <el-select v-model="filters.status" placeholder="状态" clearable @change="loadData" style="width:100%">
            <el-option label="活跃" value="active" />
            <el-option label="已过期" value="expired" />
            <el-option label="已停用" value="suspended" />
          </el-select>
        </el-col>
        <el-col :xs="12" :sm="6" :md="4">
          <el-select v-model="filters.billing_type" placeholder="计费类型" clearable @change="loadData" style="width:100%">
            <el-option label="月付" value="monthly" />
            <el-option label="季付" value="quarterly" />
            <el-option label="年付" value="yearly" />
            <el-option label="一次性" value="once" />
          </el-select>
        </el-col>
        <el-col :xs="12" :sm="6" :md="4">
          <el-input
            v-model.number="filters.expiring_days"
            placeholder="N天内到期"
            type="number"
            clearable
            @change="loadData"
          />
        </el-col>
        <el-col :xs="12" :sm="6" :md="4">
          <el-input v-model="filters.tag" placeholder="标签筛选" clearable @change="loadData" />
        </el-col>
      </el-row>
    </el-card>

    <!-- Table -->
    <el-card style="margin-top:16px">
      <el-table
        v-loading="loading"
        :data="list"
        @selection-change="selection = $event"
        row-key="id"
      >
        <el-table-column type="selection" width="42" />
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="name" label="姓名" min-width="110" show-overflow-tooltip />
        <el-table-column prop="contact" label="联系方式" min-width="120" show-overflow-tooltip />
        <el-table-column prop="region_name" label="地区" min-width="110" show-overflow-tooltip />
        <el-table-column prop="route_name" label="线路" min-width="110" show-overflow-tooltip />
        <el-table-column prop="server_name" label="服务器" min-width="110" show-overflow-tooltip />
        <el-table-column prop="node_name" label="节点" min-width="110" show-overflow-tooltip />
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <StatusBadge :status="row.status || 'unknown'" />
          </template>
        </el-table-column>
        <el-table-column prop="billing_type" label="计费类型" width="90" />
        <el-table-column label="流量" width="100">
          <template #default="{ row }">
            {{ row.traffic_used != null ? formatBytes(row.traffic_used) : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="到期日" width="130">
          <template #default="{ row }">
            <span :class="expiryClass(row.expires_at)">{{ formatDate(row.expires_at, 'YYYY-MM-DD') }}</span>
          </template>
        </el-table-column>
        <el-table-column label="金额" width="90">
          <template #default="{ row }">{{ formatMoney(row.amount) }}</template>
        </el-table-column>
        <el-table-column label="标签" min-width="120">
          <template #default="{ row }">
            <el-tag
              v-for="tag in parseTags(row.tags)"
              :key="tag"
              size="small"
              style="margin-right:4px"
            >{{ tag }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="130" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" text @click="openEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" text @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @change="loadData"
        />
      </div>
    </el-card>

    <!-- Create/Edit dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="editRow ? '编辑客户' : '新增客户'"
      width="600px"
      destroy-on-close
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="姓名" prop="name">
              <el-input v-model="form.name" placeholder="客户姓名" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="联系方式" prop="contact">
              <el-input v-model="form.contact" placeholder="电话/微信/邮箱" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="状态" prop="status">
              <el-select v-model="form.status" style="width:100%">
                <el-option label="活跃" value="active" />
                <el-option label="已过期" value="expired" />
                <el-option label="已停用" value="suspended" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="计费类型" prop="billing_type">
              <el-select v-model="form.billing_type" style="width:100%">
                <el-option label="月付" value="monthly" />
                <el-option label="季付" value="quarterly" />
                <el-option label="年付" value="yearly" />
                <el-option label="一次性" value="once" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="金额" prop="amount">
              <el-input-number v-model="form.amount" :min="0" :precision="2" style="width:100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="到期日" prop="expires_at">
              <el-date-picker
                v-model="form.expires_at"
                type="date"
                value-format="YYYY-MM-DD"
                style="width:100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="地区">
              <el-select
                v-model="form.region_name"
                clearable
                filterable
                allow-create
                default-first-option
                placeholder="可选下拉或自定义"
                style="width:100%"
              >
                <el-option
                  v-for="region in regions"
                  :key="region.id as number"
                  :label="region.name as string"
                  :value="region.name as string"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="线路">
              <el-select
                v-model="form.route_name"
                clearable
                filterable
                allow-create
                default-first-option
                placeholder="可选下拉或自定义"
                style="width:100%"
              >
                <el-option
                  v-for="route in routes"
                  :key="route.id as number"
                  :label="route.name as string"
                  :value="route.name as string"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="服务器">
              <el-select
                v-model="form.server_name"
                clearable
                filterable
                allow-create
                default-first-option
                placeholder="可选下拉或自定义"
                style="width:100%"
              >
                <el-option
                  v-for="server in servers"
                  :key="server.id as number"
                  :label="server.name as string"
                  :value="server.name as string"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="节点">
              <el-select
                v-model="form.node_name"
                clearable
                filterable
                allow-create
                default-first-option
                placeholder="可选下拉或自定义"
                style="width:100%"
              >
                <el-option
                  v-for="node in nodes"
                  :key="node.id as number"
                  :label="node.name as string"
                  :value="node.name as string"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="备注">
              <el-input v-model="form.remark" type="textarea" :rows="2" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="标签">
              <el-input v-model="form.tags" placeholder="多个标签用逗号分隔" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <!-- Batch renew dialog -->
    <el-dialog v-model="batchRenewVisible" title="批量续期" width="340px">
      <el-form label-width="80px">
        <el-form-item label="续期天数">
          <el-input-number v-model="renewDays" :min="1" :max="3650" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="batchRenewVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleBatchRenew">确认续期</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, Refresh, Download } from '@element-plus/icons-vue'
import dayjs from 'dayjs'
import { customersApi, regionsApi, routesApi, serversApi, nodesApi } from '@/api'
import { formatDate, formatBytes, formatMoney, isExpired, isExpiringSoon, downloadBlob, getListData } from '@/utils'
import PageHeader from '@/components/PageHeader.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import type { FormInstance, FormRules } from 'element-plus'

const loading = ref(false)
const saving = ref(false)
const list = ref<Record<string, unknown>[]>([])
const regions = ref<Record<string, unknown>[]>([])
const routes = ref<Record<string, unknown>[]>([])
const servers = ref<Record<string, unknown>[]>([])
const nodes = ref<Record<string, unknown>[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const selection = ref<Record<string, unknown>[]>([])
const dialogVisible = ref(false)
const batchRenewVisible = ref(false)
const editRow = ref<Record<string, unknown> | null>(null)
const renewDays = ref(30)
const formRef = ref<FormInstance>()

const filters = reactive({
  name: '',
  status: '',
  billing_type: '',
  expiring_days: undefined as number | undefined,
  tag: '',
})

const form = reactive({
  name: '',
  contact: '',
  status: 'active',
  billing_type: 'monthly',
  amount: 0,
  expires_at: '',
  region_name: '',
  route_name: '',
  server_name: '',
  node_name: '',
  remark: '',
  tags: '',
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
  status: [{ required: true, message: '请选择状态', trigger: 'change' }],
}

function parseTags(tags: unknown): string[] {
  if (!tags) return []
  if (Array.isArray(tags)) return tags as string[]
  return String(tags).split(',').filter(Boolean)
}

function expiryClass(date: unknown) {
  const d = date as string | null
  if (isExpired(d)) return 'text-danger'
  if (isExpiringSoon(d)) return 'text-warning'
  return ''
}

function formatDateForForm(date: unknown): string {
  if (!date) return ''
  return formatDate(date as string, 'YYYY-MM-DD')
}

function trimText(value: unknown): string {
  return String(value ?? '').trim()
}

function buildPayload() {
  const payload: Record<string, unknown> = {
    name: trimText(form.name),
    contact: trimText(form.contact),
    status: form.status,
    billing_type: form.billing_type,
    amount: form.amount,
    region_name: trimText(form.region_name),
    route_name: trimText(form.route_name),
    server_name: trimText(form.server_name),
    node_name: trimText(form.node_name),
    remark: trimText(form.remark),
    tags: trimText(form.tags),
  }
  if (form.expires_at) {
    payload.expires_at = dayjs(form.expires_at).startOf('day').format('YYYY-MM-DDTHH:mm:ssZ')
  }
  return payload
}

async function loadPagedOptions(apiList: (params?: Record<string, unknown>) => Promise<{ data: unknown }>) {
  const perPage = 200
  const firstRes = await apiList({ page: 1, per_page: perPage })
  const firstData = firstRes.data
  const firstList = getListData(firstData)
  if (!firstData || Array.isArray(firstData) || typeof firstData !== 'object') {
    return firstList
  }

  const total = Number((firstData as { total?: number }).total ?? firstList.length)
  if (!Number.isFinite(total) || total <= firstList.length) {
    return firstList
  }

  const pages = Math.ceil(total / perPage)
  const rest = await Promise.allSettled(
    Array.from({ length: pages - 1 }, (_, index) => apiList({ page: index + 2, per_page: perPage }))
  )
  return firstList.concat(
    rest.flatMap((result) => (result.status === 'fulfilled' ? getListData(result.value.data) : []))
  )
}

async function loadData() {
  loading.value = true
  try {
    const params: Record<string, unknown> = {
      page: page.value,
      page_size: pageSize.value,
      ...filters,
    }
    const res = await customersApi.list(params)
    const d = res.data as { total?: number }
    list.value = getListData(d)
    total.value = d.total ?? list.value.length
  } catch {
    list.value = []
  } finally {
    loading.value = false
  }
}

async function loadSelects() {
  try {
    const [regionRes, routeList, serverRes, nodeList] = await Promise.all([
      regionsApi.list(),
      loadPagedOptions(routesApi.list),
      serversApi.list(),
      loadPagedOptions(nodesApi.list),
    ])
    regions.value = getListData(regionRes.data)
    routes.value = routeList
    servers.value = getListData(serverRes.data)
    nodes.value = nodeList
  } catch {
    regions.value = []
    routes.value = []
    servers.value = []
    nodes.value = []
  }
}

function openCreate() {
  editRow.value = null
  Object.assign(form, {
    name: '',
    contact: '',
    status: 'active',
    billing_type: 'monthly',
    amount: 0,
    expires_at: '',
    region_name: '',
    route_name: '',
    server_name: '',
    node_name: '',
    remark: '',
    tags: '',
  })
  dialogVisible.value = true
}

function openEdit(row: Record<string, unknown>) {
  editRow.value = row
  Object.assign(form, {
    name: row.name ?? '',
    contact: row.contact ?? '',
    status: row.status ?? 'active',
    billing_type: row.billing_type ?? 'monthly',
    amount: row.amount ?? 0,
    expires_at: formatDateForForm(row.expires_at),
    region_name: row.region_name ?? '',
    route_name: row.route_name ?? '',
    server_name: row.server_name ?? '',
    node_name: row.node_name ?? '',
    remark: row.remark ?? '',
    tags: Array.isArray(row.tags) ? (row.tags as string[]).join(',') : row.tags ?? '',
  })
  dialogVisible.value = true
}

async function handleSave() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    saving.value = true
    try {
      const payload = buildPayload()
      if (editRow.value) {
        await customersApi.update(editRow.value.id as number, payload)
      } else {
        await customersApi.create(payload)
      }
      ElMessage.success('保存成功')
      dialogVisible.value = false
      loadData()
    } catch (e: unknown) {
      const err = e as { response?: { data?: { message?: string; error?: string } } }
      ElMessage.error(err?.response?.data?.message ?? err?.response?.data?.error ?? '保存失败')
    } finally {
      saving.value = false
    }
  })
}

async function handleDelete(row: Record<string, unknown>) {
  await ElMessageBox.confirm(`确认删除客户「${row.name}」？`, '提示', { type: 'warning' })
  try {
    await customersApi.delete(row.id as number)
    ElMessage.success('已删除')
    loadData()
  } catch {
    ElMessage.error('删除失败')
  }
}

async function handleBatchDelete() {
  if (!selection.value.length) return
  await ElMessageBox.confirm(`确认删除选中的 ${selection.value.length} 条记录？`, '提示', { type: 'warning' })
  try {
    await customersApi.batchDelete(selection.value.map((r) => r.id as number))
    ElMessage.success('批量删除成功')
    loadData()
  } catch {
    ElMessage.error('批量删除失败')
  }
}

async function handleBatchRenew() {
  if (!selection.value.length) return
  saving.value = true
  try {
    await customersApi.batchRenew(selection.value.map((r) => r.id as number), renewDays.value)
    ElMessage.success('批量续期成功')
    batchRenewVisible.value = false
    loadData()
  } catch {
    ElMessage.error('批量续期失败')
  } finally {
    saving.value = false
  }
}

async function handleExport() {
  try {
    const res = await customersApi.exportCSV({ ...filters })
    downloadBlob(res.data as Blob, 'customers.csv')
  } catch {
    ElMessage.error('导出失败')
  }
}

onMounted(() => {
  loadData()
  loadSelects()
})
</script>

<style scoped>
.filter-card { margin-bottom: 0; }
.filter-card :deep(.el-card__body) { padding: 14px 16px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
.text-danger { color: #f56c6c; font-weight: 600; }
.text-warning { color: #e6a23c; font-weight: 600; }
</style>
