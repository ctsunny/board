<template>
  <el-tag
    :type="tagType"
    size="small"
    class="status-badge"
  >
    {{ label }}
  </el-tag>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  status: string
}>()

const map: Record<string, { label: string; type: 'success' | 'danger' | 'warning' | 'info' | '' }> = {
  active:    { label: '活跃',   type: 'success' },
  expired:   { label: '已过期', type: 'danger' },
  suspended: { label: '已停用', type: 'warning' },
  online:    { label: '在线',   type: 'success' },
  offline:   { label: '离线',   type: 'danger' },
  unknown:   { label: '未知',   type: 'info' },
  enabled:   { label: '启用',   type: 'success' },
  disabled:  { label: '禁用',   type: 'danger' },
}

const entry = computed(() => map[props.status] ?? { label: props.status, type: '' as const })
const label = computed(() => entry.value.label)
const tagType = computed(() => entry.value.type)
</script>
