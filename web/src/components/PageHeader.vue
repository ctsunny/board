<template>
  <div class="page-header">
    <div class="page-header__title">
      <h2>{{ title }}</h2>
      <p v-if="subtitle" class="page-header__subtitle">{{ subtitle }}</p>
    </div>
    <div class="page-header__actions">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  subtitle?: string
}>()
</script>

<style scoped>
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 20px;
}
.page-header__title h2 {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
}
.page-header__subtitle {
  font-size: 13px;
  color: #909399;
  margin-top: 4px;
}
.page-header__actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
</style>
