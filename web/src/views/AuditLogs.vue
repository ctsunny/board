<template>
  <div>
    <PageHeader title="操作日志" subtitle="系统操作记录（只读）" />

    <!-- Filter -->
    <el-card class="filter-card">
      <el-row :gutter="12">
        <el-col :xs="12" :sm="6">
          <el-input v-model="filters.action" placeholder="操作类型" clearable @change="loadData" />
        </el-col>
        <el-col :xs="12" :sm="6">
          <el-input v-model="filters.resource" placeholder="资源" clearable @change="loadData" />
        </el-col>
        <el-col :xs="12" :sm="8">
          <el-date-picker
            v-model="filters.date_range"
            type="daterange"
            range-separator="-"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            style="width:100%"
            @change="loadData"
          />
        </el-col>
      </el-row>
    </el-card>

    <el-card style="margin-top:16px">
      <el-table v-loading="loading" :data="list" style="width:100%">
        <el-table-column label="时间" width="160">
          <template #default="{ row }">{{ formatDate(row.created_at as string) }}</template>
        </el-table-column>
        <el-table-column prop="action" label="操作" width="120" />
        <el-table-column prop="resource" label="资源" width="120" />
        <el-table-column prop="detail" label="详情" show-overflow-tooltip />
        <el-table-column prop="ip" label="IP" width="130" />
        <el-table-column prop="operator" label="操作人" width="100" />
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
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { auditLogsApi } from '@/api'
import { formatDate, getListData } from '@/utils'
import PageHeader from '@/components/PageHeader.vue'

const loading = ref(false)
const list = ref<Record<string, unknown>[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const filters = reactive({
  action: '',
  resource: '',
  date_range: [] as string[],
})

async function loadData() {
  loading.value = true
  try {
    const params: Record<string, unknown> = {
      page: page.value,
      per_page: pageSize.value,
      action: filters.action,
      resource: filters.resource,
    }
    if (filters.date_range?.length === 2) {
      params.start_date = filters.date_range[0]
      params.end_date = filters.date_range[1]
    }
    const res = await auditLogsApi.list(params)
    const d = res.data as { total?: number }
    list.value = getListData(d)
    total.value = d.total ?? list.value.length
  } catch {
    list.value = []
  } finally {
    loading.value = false
  }
}

onMounted(loadData)
</script>

<style scoped>
.filter-card :deep(.el-card__body) { padding: 14px 16px; }
.pagination-wrap { margin-top: 16px; display: flex; justify-content: flex-end; }
</style>
