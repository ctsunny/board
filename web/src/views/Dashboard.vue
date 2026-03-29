<template>
  <div>
    <PageHeader title="仪表盘" subtitle="系统运行概览" />

    <!-- Stats cards -->
    <el-row :gutter="16" class="stats-row">
      <el-col
        v-for="card in statCards"
        :key="card.key"
        :xs="12" :sm="8" :md="4"
      >
        <div class="stat-card" :style="{ borderTopColor: card.color }">
          <div class="stat-value" :style="{ color: card.color }">{{ stats[card.key] ?? '-' }}</div>
          <div class="stat-label">{{ card.label }}</div>
        </div>
      </el-col>
    </el-row>

    <el-row :gutter="16" style="margin-top:20px">
      <!-- Customer trend chart -->
      <el-col :xs="24" :md="14">
        <el-card>
          <template #header>
            <span>客户数量趋势（近30天）</span>
          </template>
          <v-chart :option="chartOption" style="height:260px" autoresize />
        </el-card>
      </el-col>

      <!-- Server status -->
      <el-col :xs="24" :md="10">
        <el-card>
          <template #header><span>服务器状态</span></template>
          <div v-if="servers.length === 0" class="empty-tip">暂无服务器数据</div>
          <div class="server-grid">
            <div
              v-for="srv in servers"
              :key="srv.id"
              class="server-card"
            >
              <div class="server-dot" :class="srv.status" />
              <div class="server-info">
                <div class="server-ip">{{ srv.ip || srv.name }}</div>
                <div class="server-meta">
                  <StatusBadge :status="srv.status || 'unknown'" />
                  <span v-if="srv.latency != null" class="latency-badge">{{ srv.latency }}ms</span>
                </div>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card style="margin-top:20px">
      <template #header><span>客户面板</span></template>
      <div v-if="recentCustomers.length === 0" class="empty-tip">暂无客户数据</div>
      <el-table v-else :data="recentCustomers" size="small" style="width:100%">
        <el-table-column prop="name" label="客户" min-width="120" show-overflow-tooltip />
        <el-table-column prop="contact" label="联系方式" min-width="140" show-overflow-tooltip />
        <el-table-column label="资源信息" min-width="220" show-overflow-tooltip>
          <template #default="{ row }">
            {{ [row.region_name, row.route_name, row.server_name, row.node_name].filter(Boolean).join(' / ') || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <StatusBadge :status="row.status || 'unknown'" />
          </template>
        </el-table-column>
        <el-table-column label="到期日" width="130">
          <template #default="{ row }">{{ formatDate(row.expires_at, 'YYYY-MM-DD') }}</template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Recent audit logs -->
    <el-card style="margin-top:20px">
      <template #header><span>最近操作日志</span></template>
      <el-table :data="auditLogs" size="small" style="width:100%">
        <el-table-column prop="created_at" label="时间" width="160">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
        <el-table-column prop="action" label="操作" width="120" />
        <el-table-column prop="resource" label="资源" width="120" />
        <el-table-column prop="detail" label="详情" show-overflow-tooltip />
        <el-table-column prop="ip" label="IP" width="130" />
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { dashboardApi, serversApi, auditLogsApi, customersApi } from '@/api'
import { formatDate, getListData } from '@/utils'
import PageHeader from '@/components/PageHeader.vue'
import StatusBadge from '@/components/StatusBadge.vue'
import dayjs from 'dayjs'

interface ServerItem {
  id: number | string
  name: string
  ip?: string
  status?: string
  latency?: number
}

const stats = reactive<Record<string, number>>({})
const servers = ref<ServerItem[]>([])
const auditLogs = ref<Record<string, unknown>[]>([])
const recentCustomers = ref<Record<string, unknown>[]>([])

const statCards = [
  { key: 'total_customers',   label: '总客户数',   color: '#409eff' },
  { key: 'active_customers',  label: '活跃客户',   color: '#67c23a' },
  { key: 'expiring_soon',     label: '即将到期(7天)', color: '#e6a23c' },
  { key: 'online_servers',    label: '在线服务器', color: '#67c23a' },
  { key: 'offline_servers',   label: '离线服务器', color: '#f56c6c' },
  { key: 'monthly_revenue',   label: '本月收入(¥)', color: '#9b59b6' },
]

// Generate mock 30-day trend
function makeTrendData() {
  const days: string[] = []
  const values: number[] = []
  for (let i = 29; i >= 0; i--) {
    days.push(dayjs().subtract(i, 'day').format('MM/DD'))
    values.push(Math.floor(80 + Math.random() * 40))
  }
  return { days, values }
}

const trend = makeTrendData()

const chartOption = ref({
  tooltip: { trigger: 'axis' },
  grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
  xAxis: {
    type: 'category',
    data: trend.days,
    axisLabel: { fontSize: 11 },
  },
  yAxis: { type: 'value', minInterval: 1 },
  series: [
    {
      name: '客户数',
      type: 'line',
      data: trend.values,
      smooth: true,
      areaStyle: { opacity: 0.15 },
      itemStyle: { color: '#409eff' },
      lineStyle: { width: 2 },
    },
  ],
})

async function loadData() {
  try {
    const res = await dashboardApi.getDashboard()
    const d = res.data as Record<string, unknown>
    Object.assign(stats, d)
    if (Array.isArray(d.trend_days) && Array.isArray(d.trend_values)) {
      chartOption.value.xAxis.data = d.trend_days as string[]
      chartOption.value.series[0].data = d.trend_values as number[]
    }
  } catch {
    // use defaults
  }

  try {
    const sRes = await serversApi.list()
    servers.value = getListData(sRes.data) as unknown as ServerItem[]
  } catch {
    servers.value = []
  }

  try {
    const lRes = await auditLogsApi.list({ page: 1, page_size: 10 })
    auditLogs.value = getListData(lRes.data)
  } catch {
    auditLogs.value = []
  }

  try {
    const cRes = await customersApi.list({ page: 1, per_page: 8 })
    recentCustomers.value = getListData(cRes.data)
  } catch {
    recentCustomers.value = []
  }
}

onMounted(loadData)
</script>

<style scoped>
.stats-row { margin-bottom: 8px; }

.stat-card {
  background: #fff;
  border-radius: 8px;
  padding: 20px 16px;
  border-top: 3px solid #409eff;
  box-shadow: 0 2px 8px rgba(0,0,0,0.06);
  margin-bottom: 16px;
  text-align: center;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  line-height: 1.2;
}

.stat-label {
  font-size: 13px;
  color: #909399;
  margin-top: 6px;
}

.server-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 10px;
}

.server-card {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
}

.server-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.server-dot.online  { background: #67c23a; }
.server-dot.offline { background: #f56c6c; }
.server-dot.unknown { background: #c0c4cc; }

.server-ip { font-size: 13px; font-weight: 500; }

.server-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 4px;
}

.latency-badge {
  font-size: 11px;
  color: #67c23a;
  background: #f0f9eb;
  padding: 1px 6px;
  border-radius: 10px;
}

.empty-tip {
  color: #909399;
  text-align: center;
  padding: 20px;
  font-size: 13px;
}
</style>
