<template>
  <div class="live-resources">
    <PageHeader title="直播资源管理" subtitle="将地区、线路、服务器、节点整合到一个统一界面中维护。">
      <el-button type="primary" @click="openCreate(activeTab)">新增{{ tabLabels[activeTab] }}</el-button>
    </PageHeader>

    <div class="overview-grid">
      <button
        v-for="tab in tabs"
        :key="tab"
        type="button"
        class="overview-card"
        :class="{ 'overview-card--active': activeTab === tab }"
        @click="activeTab = tab"
      >
        <span class="overview-card__label">{{ tabLabels[tab] }}</span>
        <strong class="overview-card__value">{{ counts[tab] }}</strong>
      </button>
    </div>

    <el-card class="content-card">
      <el-tabs v-model="activeTab" stretch>
        <el-tab-pane label="直播地区" name="regions">
          <section class="resource-section">
            <div class="section-toolbar">
              <div>
                <h3>直播地区</h3>
                <p>维护直播地区基础信息，供线路选择关联。</p>
              </div>
              <el-button type="primary" @click="openCreate('regions')">新增地区</el-button>
            </div>

            <el-table v-loading="loading.regions" :data="regions" style="width:100%">
              <el-table-column prop="id" label="ID" width="70" />
              <el-table-column prop="name" label="名称" min-width="140" />
              <el-table-column prop="code" label="代码" width="120" />
              <el-table-column prop="remark" label="备注" show-overflow-tooltip />
              <el-table-column label="操作" width="130" fixed="right">
                <template #default="{ row }">
                  <el-button size="small" type="primary" text @click="openEditRegion(row)">编辑</el-button>
                  <el-button size="small" type="danger" text @click="handleDeleteRegion(row)">删除</el-button>
                </template>
              </el-table-column>
            </el-table>
          </section>
        </el-tab-pane>

        <el-tab-pane label="直播线路" name="routes">
          <section class="resource-section">
            <div class="section-toolbar">
              <div>
                <h3>直播线路</h3>
                <p>按地区、服务器统一配置直播线路。</p>
              </div>
              <el-button type="primary" @click="openCreate('routes')">新增线路</el-button>
            </div>

            <el-card class="filter-card" shadow="never">
              <el-row :gutter="12">
                <el-col :xs="24" :sm="8">
                  <el-select v-model="routeFilters.region_id" placeholder="按地区筛选" clearable style="width:100%" @change="loadRoutes">
                    <el-option v-for="item in regions" :key="item.id" :label="item.name" :value="item.id" />
                  </el-select>
                </el-col>
                <el-col :xs="24" :sm="8">
                  <el-select v-model="routeFilters.server_id" placeholder="按服务器筛选" clearable style="width:100%" @change="loadRoutes">
                    <el-option v-for="item in servers" :key="item.id" :label="item.name" :value="item.id" />
                  </el-select>
                </el-col>
                <el-col :xs="24" :sm="8">
                  <el-select v-model="routeFilters.status" placeholder="按状态筛选" clearable style="width:100%" @change="loadRoutes">
                    <el-option label="启用" value="enabled" />
                    <el-option label="禁用" value="disabled" />
                    <el-option label="活跃" value="active" />
                  </el-select>
                </el-col>
              </el-row>
            </el-card>

            <el-table v-loading="loading.routes" :data="routes" style="width:100%">
              <el-table-column prop="id" label="ID" width="70" />
              <el-table-column prop="name" label="名称" min-width="140" />
              <el-table-column prop="url" label="URL" min-width="200" show-overflow-tooltip />
              <el-table-column label="协议" width="90">
                <template #default="{ row }">
                  <el-tag :type="protocolType(row.protocol)" size="small">{{ row.protocol }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="region_name" label="地区" width="120" />
              <el-table-column prop="server_name" label="服务器" width="130" />
              <el-table-column label="状态" width="90">
                <template #default="{ row }">
                  <StatusBadge :status="row.status || 'unknown'" />
                </template>
              </el-table-column>
              <el-table-column label="操作" width="130" fixed="right">
                <template #default="{ row }">
                  <el-button size="small" type="primary" text @click="openEditRoute(row)">编辑</el-button>
                  <el-button size="small" type="danger" text @click="handleDeleteRoute(row)">删除</el-button>
                </template>
              </el-table-column>
            </el-table>
          </section>
        </el-tab-pane>

        <el-tab-pane label="服务器" name="servers">
          <section class="resource-section">
            <div class="section-toolbar">
              <div>
                <h3>服务器</h3>
                <p>集中维护服务器资源并支持即时 Ping 检测。</p>
              </div>
              <el-button type="primary" @click="openCreate('servers')">新增服务器</el-button>
            </div>

            <el-table v-loading="loading.servers" :data="servers" style="width:100%">
              <el-table-column prop="id" label="ID" width="70" />
              <el-table-column prop="name" label="名称" min-width="140" />
              <el-table-column prop="ip" label="IP地址" width="150" />
              <el-table-column prop="location" label="位置" width="130" />
              <el-table-column label="状态" width="95">
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
              <el-table-column label="最后Ping" width="170">
                <template #default="{ row }">{{ formatDate(row.last_ping_at) }}</template>
              </el-table-column>
              <el-table-column label="操作" width="200" fixed="right">
                <template #default="{ row }">
                  <el-button size="small" :loading="pingingId === row.id" @click="handlePing(row)">Ping</el-button>
                  <el-button size="small" type="primary" text @click="openEditServer(row)">编辑</el-button>
                  <el-button size="small" type="danger" text @click="handleDeleteServer(row)">删除</el-button>
                </template>
              </el-table-column>
            </el-table>
          </section>
        </el-tab-pane>

        <el-tab-pane label="节点" name="nodes">
          <section class="resource-section">
            <div class="section-toolbar">
              <div>
                <h3>节点</h3>
                <p>统一管理线路下的节点与服务器分配关系。</p>
              </div>
              <el-button type="primary" @click="openCreate('nodes')">新增节点</el-button>
            </div>

            <el-card class="filter-card" shadow="never">
              <el-row :gutter="12">
                <el-col :xs="24" :sm="12">
                  <el-select v-model="nodeFilters.route_id" placeholder="按线路筛选" clearable style="width:100%" @change="loadNodes">
                    <el-option v-for="item in routes" :key="item.id" :label="item.name" :value="item.id" />
                  </el-select>
                </el-col>
                <el-col :xs="24" :sm="12">
                  <el-select v-model="nodeFilters.server_id" placeholder="按服务器筛选" clearable style="width:100%" @change="loadNodes">
                    <el-option v-for="item in servers" :key="item.id" :label="item.name" :value="item.id" />
                  </el-select>
                </el-col>
              </el-row>
            </el-card>

            <el-table v-loading="loading.nodes" :data="nodes" style="width:100%">
              <el-table-column prop="id" label="ID" width="70" />
              <el-table-column prop="name" label="名称" min-width="140" />
              <el-table-column prop="route_name" label="线路" min-width="130" />
              <el-table-column prop="server_name" label="服务器" min-width="130" />
              <el-table-column prop="address" label="地址" min-width="170" show-overflow-tooltip />
              <el-table-column prop="port" label="端口" width="80" />
              <el-table-column label="协议" width="90">
                <template #default="{ row }">
                  <el-tag size="small">{{ row.protocol }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column label="状态" width="90">
                <template #default="{ row }">
                  <StatusBadge :status="row.status || 'unknown'" />
                </template>
              </el-table-column>
              <el-table-column label="操作" width="130" fixed="right">
                <template #default="{ row }">
                  <el-button size="small" type="primary" text @click="openEditNode(row)">编辑</el-button>
                  <el-button size="small" type="danger" text @click="handleDeleteNode(row)">删除</el-button>
                </template>
              </el-table-column>
            </el-table>
          </section>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <el-dialog v-model="dialogs.regions" :title="editing.regions ? '编辑地区' : '新增地区'" width="420px" destroy-on-close>
      <el-form ref="regionFormRef" :model="regionForm" :rules="regionRules" label-width="70px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="regionForm.name" placeholder="地区名称" />
        </el-form-item>
        <el-form-item label="代码" prop="code">
          <el-input v-model="regionForm.code" placeholder="如 CN-BJ" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="regionForm.remark" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogs.regions = false">取消</el-button>
        <el-button type="primary" :loading="saving.regions" @click="handleSaveRegion">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="dialogs.routes" :title="editing.routes ? '编辑线路' : '新增线路'" width="500px" destroy-on-close>
      <el-form ref="routeFormRef" :model="routeForm" :rules="routeRules" label-width="80px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="routeForm.name" />
        </el-form-item>
        <el-form-item label="URL" prop="url">
          <el-input v-model="routeForm.url" placeholder="rtmp://..." />
        </el-form-item>
        <el-form-item label="协议" prop="protocol">
          <el-select v-model="routeForm.protocol" style="width:100%">
            <el-option label="RTMP" value="RTMP" />
            <el-option label="HLS" value="HLS" />
            <el-option label="RTSP" value="RTSP" />
            <el-option label="SRT" value="SRT" />
          </el-select>
        </el-form-item>
        <el-form-item label="地区">
          <el-select v-model="routeForm.region_id" clearable style="width:100%">
            <el-option v-for="item in regions" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="服务器">
          <el-select v-model="routeForm.server_id" clearable style="width:100%">
            <el-option v-for="item in servers" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="routeForm.status" style="width:100%">
            <el-option label="启用" value="enabled" />
            <el-option label="禁用" value="disabled" />
            <el-option label="活跃" value="active" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="routeForm.remark" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogs.routes = false">取消</el-button>
        <el-button type="primary" :loading="saving.routes" @click="handleSaveRoute">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="dialogs.servers" :title="editing.servers ? '编辑服务器' : '新增服务器'" width="460px" destroy-on-close>
      <el-form ref="serverFormRef" :model="serverForm" :rules="serverRules" label-width="70px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="serverForm.name" />
        </el-form-item>
        <el-form-item label="IP" prop="ip">
          <el-input v-model="serverForm.ip" placeholder="192.168.1.1" />
        </el-form-item>
        <el-form-item label="位置">
          <el-input v-model="serverForm.location" placeholder="北京/上海/香港" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="serverForm.remark" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogs.servers = false">取消</el-button>
        <el-button type="primary" :loading="saving.servers" @click="handleSaveServer">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="dialogs.nodes" :title="editing.nodes ? '编辑节点' : '新增节点'" width="500px" destroy-on-close>
      <el-form ref="nodeFormRef" :model="nodeForm" :rules="nodeRules" label-width="80px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="nodeForm.name" />
        </el-form-item>
        <el-form-item label="线路">
          <el-select v-model="nodeForm.route_id" clearable style="width:100%">
            <el-option v-for="item in routes" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="服务器">
          <el-select v-model="nodeForm.server_id" clearable style="width:100%">
            <el-option v-for="item in servers" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="地址" prop="address">
          <el-input v-model="nodeForm.address" placeholder="IP 或域名" />
        </el-form-item>
        <el-form-item label="端口" prop="port">
          <el-input-number v-model="nodeForm.port" :min="1" :max="65535" style="width:100%" />
        </el-form-item>
        <el-form-item label="协议">
          <el-select v-model="nodeForm.protocol" style="width:100%">
            <el-option label="RTMP" value="RTMP" />
            <el-option label="HLS" value="HLS" />
            <el-option label="RTSP" value="RTSP" />
            <el-option label="SRT" value="SRT" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="nodeForm.status" style="width:100%">
            <el-option label="启用" value="enabled" />
            <el-option label="禁用" value="disabled" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogs.nodes = false">取消</el-button>
        <el-button type="primary" :loading="saving.nodes" @click="handleSaveNode">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { nodesApi, regionsApi, routesApi, serversApi } from '@/api'
import { formatDate } from '@/utils'
import PageHeader from '@/components/PageHeader.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import type { FormInstance, FormRules } from 'element-plus'

type TabKey = 'regions' | 'routes' | 'servers' | 'nodes'
type ProtocolType = '' | 'success' | 'warning' | 'info' | 'danger'
type ListData<T> = T[] | { list?: T[]; items?: T[] }

interface Region {
  id: number
  name: string
  code?: string
  remark?: string
}

interface Server {
  id: number
  name: string
  ip?: string
  location?: string
  status?: string
  latency?: number | null
  last_ping_at?: string
  remark?: string
}

interface RouteItem {
  id: number
  name: string
  url?: string
  protocol?: string
  region_id?: number
  server_id?: number
  region_name?: string
  server_name?: string
  status?: string
  remark?: string
}

interface NodeItem {
  id: number
  name: string
  route_id?: number
  server_id?: number
  route_name?: string
  server_name?: string
  address?: string
  port?: number
  protocol?: string
  status?: string
}

const tabs: TabKey[] = ['regions', 'routes', 'servers', 'nodes']
const tabLabels: Record<TabKey, string> = {
  regions: '地区',
  routes: '线路',
  servers: '服务器',
  nodes: '节点',
}

const route = useRoute()
const router = useRouter()

const activeTab = ref<TabKey>('regions')
const pingingId = ref<number | null>(null)

const loading = reactive<Record<TabKey, boolean>>({
  regions: false,
  routes: false,
  servers: false,
  nodes: false,
})

const saving = reactive<Record<TabKey, boolean>>({
  regions: false,
  routes: false,
  servers: false,
  nodes: false,
})

const dialogs = reactive<Record<TabKey, boolean>>({
  regions: false,
  routes: false,
  servers: false,
  nodes: false,
})

const editing = reactive<{
  regions: Region | null
  routes: RouteItem | null
  servers: Server | null
  nodes: NodeItem | null
}>({
  regions: null,
  routes: null,
  servers: null,
  nodes: null,
})

const regions = ref<Region[]>([])
const routes = ref<RouteItem[]>([])
const servers = ref<Server[]>([])
const nodes = ref<NodeItem[]>([])

const counts = computed(() => ({
  regions: regions.value.length,
  routes: routes.value.length,
  servers: servers.value.length,
  nodes: nodes.value.length,
}))

const regionFormRef = ref<FormInstance>()
const routeFormRef = ref<FormInstance>()
const serverFormRef = ref<FormInstance>()
const nodeFormRef = ref<FormInstance>()

const regionForm = reactive({
  name: '',
  code: '',
  remark: '',
})

const routeFilters = reactive({
  region_id: undefined as number | undefined,
  server_id: undefined as number | undefined,
  status: '',
})

const routeForm = reactive({
  name: '',
  url: '',
  protocol: 'RTMP',
  region_id: undefined as number | undefined,
  server_id: undefined as number | undefined,
  status: 'enabled',
  remark: '',
})

const serverForm = reactive({
  name: '',
  ip: '',
  location: '',
  remark: '',
})

const nodeFilters = reactive({
  route_id: undefined as number | undefined,
  server_id: undefined as number | undefined,
})

const nodeForm = reactive({
  name: '',
  route_id: undefined as number | undefined,
  server_id: undefined as number | undefined,
  address: '',
  port: 1935,
  protocol: 'RTMP',
  status: 'enabled',
})

const regionRules: FormRules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
}

const routeRules: FormRules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  url: [{ required: true, message: '请输入URL', trigger: 'blur' }],
  protocol: [{ required: true, message: '请选择协议', trigger: 'change' }],
}

const serverRules: FormRules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  ip: [{ required: true, message: '请输入IP地址', trigger: 'blur' }],
}

const nodeRules: FormRules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  address: [{ required: true, message: '请输入地址', trigger: 'blur' }],
}

function protocolType(protocol?: string): ProtocolType {
  const map: Record<string, ProtocolType> = { RTMP: '', HLS: 'success', RTSP: 'warning', SRT: 'info' }
  return protocol ? (map[protocol] ?? '') : ''
}

function normalizeList<T>(data: ListData<T>): T[] {
  return Array.isArray(data) ? data : data.list ?? data.items ?? []
}

function getQueryTab(value: unknown): TabKey | undefined {
  const raw = Array.isArray(value) ? value[0] : value
  return typeof raw === 'string' && tabs.includes(raw as TabKey) ? (raw as TabKey) : undefined
}

async function validateForm(formRef: FormInstance | undefined) {
  if (!formRef) return false
  try {
    await formRef.validate()
    return true
  } catch {
    return false
  }
}

async function loadRegions() {
  loading.regions = true
  try {
    const res = await regionsApi.list()
    regions.value = normalizeList(res.data as ListData<Region>)
  } catch {
    regions.value = []
  } finally {
    loading.regions = false
  }
}

async function loadRoutes() {
  loading.routes = true
  try {
    const res = await routesApi.list({ ...routeFilters })
    routes.value = normalizeList(res.data as ListData<RouteItem>)
  } catch {
    routes.value = []
  } finally {
    loading.routes = false
  }
}

async function loadServers() {
  loading.servers = true
  try {
    const res = await serversApi.list()
    servers.value = normalizeList(res.data as ListData<Server>)
  } catch {
    servers.value = []
  } finally {
    loading.servers = false
  }
}

async function loadNodes() {
  loading.nodes = true
  try {
    const res = await nodesApi.list({ ...nodeFilters })
    nodes.value = normalizeList(res.data as ListData<NodeItem>)
  } catch {
    nodes.value = []
  } finally {
    loading.nodes = false
  }
}

function openCreate(tab: TabKey) {
  activeTab.value = tab
  if (tab === 'regions') {
    editing.regions = null
    Object.assign(regionForm, { name: '', code: '', remark: '' })
  }
  if (tab === 'routes') {
    editing.routes = null
    Object.assign(routeForm, { name: '', url: '', protocol: 'RTMP', region_id: undefined, server_id: undefined, status: 'enabled', remark: '' })
  }
  if (tab === 'servers') {
    editing.servers = null
    Object.assign(serverForm, { name: '', ip: '', location: '', remark: '' })
  }
  if (tab === 'nodes') {
    editing.nodes = null
    Object.assign(nodeForm, { name: '', route_id: undefined, server_id: undefined, address: '', port: 1935, protocol: 'RTMP', status: 'enabled' })
  }
  dialogs[tab] = true
}

function openEditRegion(row: Region) {
  editing.regions = row
  Object.assign(regionForm, { name: row.name ?? '', code: row.code ?? '', remark: row.remark ?? '' })
  dialogs.regions = true
}

function openEditRoute(row: RouteItem) {
  editing.routes = row
  Object.assign(routeForm, {
    name: row.name ?? '',
    url: row.url ?? '',
    protocol: row.protocol ?? 'RTMP',
    region_id: row.region_id,
    server_id: row.server_id,
    status: row.status ?? 'enabled',
    remark: row.remark ?? '',
  })
  dialogs.routes = true
}

function openEditServer(row: Server) {
  editing.servers = row
  Object.assign(serverForm, { name: row.name ?? '', ip: row.ip ?? '', location: row.location ?? '', remark: row.remark ?? '' })
  dialogs.servers = true
}

function openEditNode(row: NodeItem) {
  editing.nodes = row
  Object.assign(nodeForm, {
    name: row.name ?? '',
    route_id: row.route_id,
    server_id: row.server_id,
    address: row.address ?? '',
    port: row.port ?? 1935,
    protocol: row.protocol ?? 'RTMP',
    status: row.status ?? 'enabled',
  })
  dialogs.nodes = true
}

async function handleSaveRegion() {
  if (!(await validateForm(regionFormRef.value))) return
  saving.regions = true
  try {
    if (editing.regions) {
      await regionsApi.update(editing.regions.id, { ...regionForm })
    } else {
      await regionsApi.create({ ...regionForm })
    }
    ElMessage.success('地区保存成功')
    dialogs.regions = false
    await Promise.all([loadRegions(), loadRoutes()])
  } catch {
    ElMessage.error('地区保存失败')
  } finally {
    saving.regions = false
  }
}

async function handleSaveRoute() {
  if (!(await validateForm(routeFormRef.value))) return
  saving.routes = true
  try {
    if (editing.routes) {
      await routesApi.update(editing.routes.id, { ...routeForm })
    } else {
      await routesApi.create({ ...routeForm })
    }
    ElMessage.success('线路保存成功')
    dialogs.routes = false
    await Promise.all([loadRoutes(), loadNodes()])
  } catch {
    ElMessage.error('线路保存失败')
  } finally {
    saving.routes = false
  }
}

async function handleSaveServer() {
  if (!(await validateForm(serverFormRef.value))) return
  saving.servers = true
  try {
    if (editing.servers) {
      await serversApi.update(editing.servers.id, { ...serverForm })
    } else {
      await serversApi.create({ ...serverForm })
    }
    ElMessage.success('服务器保存成功')
    dialogs.servers = false
    await Promise.all([loadServers(), loadRoutes(), loadNodes()])
  } catch {
    ElMessage.error('服务器保存失败')
  } finally {
    saving.servers = false
  }
}

async function handleSaveNode() {
  if (!(await validateForm(nodeFormRef.value))) return
  saving.nodes = true
  try {
    if (editing.nodes) {
      await nodesApi.update(editing.nodes.id, { ...nodeForm })
    } else {
      await nodesApi.create({ ...nodeForm })
    }
    ElMessage.success('节点保存成功')
    dialogs.nodes = false
    await loadNodes()
  } catch {
    ElMessage.error('节点保存失败')
  } finally {
    saving.nodes = false
  }
}

async function handleDeleteRegion(row: Region) {
  await ElMessageBox.confirm(`确认删除地区「${row.name}」？`, '提示', { type: 'warning' })
  try {
    await regionsApi.delete(row.id)
    ElMessage.success('地区已删除')
    await Promise.all([loadRegions(), loadRoutes()])
  } catch {
    ElMessage.error('地区删除失败')
  }
}

async function handleDeleteRoute(row: RouteItem) {
  await ElMessageBox.confirm(`确认删除线路「${row.name}」？`, '提示', { type: 'warning' })
  try {
    await routesApi.delete(row.id)
    ElMessage.success('线路已删除')
    await Promise.all([loadRoutes(), loadNodes()])
  } catch {
    ElMessage.error('线路删除失败')
  }
}

async function handleDeleteServer(row: Server) {
  await ElMessageBox.confirm(`确认删除服务器「${row.name}」？`, '提示', { type: 'warning' })
  try {
    await serversApi.delete(row.id)
    ElMessage.success('服务器已删除')
    await Promise.all([loadServers(), loadRoutes(), loadNodes()])
  } catch {
    ElMessage.error('服务器删除失败')
  }
}

async function handleDeleteNode(row: NodeItem) {
  await ElMessageBox.confirm(`确认删除节点「${row.name}」？`, '提示', { type: 'warning' })
  try {
    await nodesApi.delete(row.id)
    ElMessage.success('节点已删除')
    await loadNodes()
  } catch {
    ElMessage.error('节点删除失败')
  }
}

async function handlePing(row: Server) {
  pingingId.value = row.id
  try {
    const res = await serversApi.ping(row.id)
    ElMessage.success(`Ping 成功，延迟 ${(res.data as { latency?: number }).latency ?? '?'}ms`)
    await loadServers()
  } catch {
    ElMessage.error('Ping 失败')
  } finally {
    pingingId.value = null
  }
}

watch(
  () => route.query.tab,
  (tab) => {
    activeTab.value = getQueryTab(tab) ?? 'regions'
  },
  { immediate: true },
)

watch(activeTab, async (tab) => {
  if (getQueryTab(route.query.tab) !== tab) {
    await router.replace({ query: { ...route.query, tab } })
  }
})

onMounted(async () => {
  await Promise.all([loadRegions(), loadRoutes(), loadServers(), loadNodes()])
})
</script>

<style scoped>
.live-resources {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.overview-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.overview-card {
  border: 1px solid #e4e7ed;
  background: #fff;
  border-radius: 12px;
  padding: 16px;
  text-align: left;
  cursor: pointer;
  transition: all 0.2s ease;
}

.overview-card:hover,
.overview-card--active {
  border-color: #409eff;
  box-shadow: 0 8px 20px rgba(64, 158, 255, 0.12);
  transform: translateY(-1px);
}

.overview-card__label {
  display: block;
  font-size: 13px;
  color: #909399;
  margin-bottom: 8px;
}

.overview-card__value {
  font-size: 28px;
  color: #303133;
  line-height: 1;
}

.content-card :deep(.el-card__body) {
  padding-top: 8px;
}

.resource-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.section-toolbar {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
  align-items: center;
}

.section-toolbar h3 {
  font-size: 18px;
  color: #303133;
  margin-bottom: 4px;
}

.section-toolbar p {
  font-size: 13px;
  color: #909399;
}

.filter-card :deep(.el-card__body) {
  padding: 14px 16px;
}

.status-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 6px;
  vertical-align: middle;
}

.status-dot.online { background: #67c23a; }
.status-dot.offline { background: #f56c6c; }
.status-dot.unknown { background: #c0c4cc; }

.latency-badge {
  font-size: 12px;
  color: #67c23a;
  background: #f0f9eb;
  padding: 2px 8px;
  border-radius: 10px;
}

@media (max-width: 767px) {
  .overview-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
