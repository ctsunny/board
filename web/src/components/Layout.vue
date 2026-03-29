<template>
  <div class="layout" :class="{ 'sidebar-collapsed': sidebarCollapsed }">
    <!-- Mobile overlay -->
    <div v-if="mobileOpen" class="mobile-overlay" @click="mobileOpen = false" />

    <!-- Sidebar -->
    <aside class="sidebar" :class="{ 'mobile-open': mobileOpen }">
      <div class="sidebar__brand">
        <el-icon class="brand-icon"><Monitor /></el-icon>
        <span class="brand-name">Board</span>
      </div>
      <nav class="sidebar__nav">
        <router-link
          v-for="item in menuItems"
          :key="item.path"
          :to="item.path"
          class="nav-item"
          active-class="nav-item--active"
          @click="mobileOpen = false"
        >
          <el-icon><component :is="item.icon" /></el-icon>
          <span>{{ item.label }}</span>
        </router-link>
      </nav>
    </aside>

    <!-- Main content -->
    <div class="main-wrap">
      <!-- Top bar -->
      <header class="topbar">
        <div class="topbar__left">
          <el-button
            class="hamburger-btn"
            text
            @click="toggleSidebar"
          >
            <el-icon size="20"><Expand /></el-icon>
          </el-button>
          <span class="topbar__title">Board 管理系统</span>
        </div>
        <div class="topbar__right">
          <el-tag type="success" size="small">管理员</el-tag>
          <span class="admin-name">{{ auth.username }}</span>
          <el-button type="danger" size="small" text @click="handleLogout">
            <el-icon><SwitchButton /></el-icon>
            退出
          </el-button>
        </div>
      </header>

      <!-- Page content -->
      <main class="page-content">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'

const router = useRouter()
const auth = useAuthStore()
const appStore = useAppStore()

const mobileOpen = ref(false)
const isMobile = ref(window.innerWidth < 768)
const sidebarCollapsed = computed(() => appStore.sidebarCollapsed)

function toggleSidebar() {
  if (isMobile.value) {
    mobileOpen.value = !mobileOpen.value
  } else {
    appStore.toggleSidebar()
  }
}

function handleLogout() {
  auth.logout()
  router.push('/login')
}

function handleResize() {
  isMobile.value = window.innerWidth < 768
  if (!isMobile.value) mobileOpen.value = false
}

onMounted(() => window.addEventListener('resize', handleResize))
onUnmounted(() => window.removeEventListener('resize', handleResize))

const menuItems = [
  { path: '/dashboard', label: '仪表盘', icon: 'House' },
  { path: '/customers', label: '客户管理', icon: 'User' },
  { path: '/regions', label: '直播地区', icon: 'Location' },
  { path: '/routes', label: '直播线路', icon: 'Connection' },
  { path: '/servers', label: '服务器', icon: 'Monitor' },
  { path: '/nodes', label: '节点', icon: 'Share' },
  { path: '/tokens', label: 'API Token', icon: 'Key' },
  { path: '/audit-logs', label: '操作日志', icon: 'Document' },
  { path: '/settings', label: '系统设置', icon: 'Setting' },
]
</script>

<style scoped>
.layout {
  display: flex;
  height: 100vh;
  overflow: hidden;
}

.sidebar {
  width: 220px;
  min-width: 220px;
  background: #1a1d2b;
  display: flex;
  flex-direction: column;
  transition: transform 0.3s;
  z-index: 100;
}

.sidebar__brand {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 20px 18px;
  color: #fff;
  font-size: 18px;
  font-weight: 700;
  border-bottom: 1px solid rgba(255,255,255,0.08);
}

.brand-icon { font-size: 24px; color: #409eff; }
.brand-name { letter-spacing: 1px; }

.sidebar__nav {
  flex: 1;
  overflow-y: auto;
  padding: 10px 0;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 20px;
  color: rgba(255,255,255,0.65);
  text-decoration: none;
  font-size: 14px;
  transition: all 0.2s;
  border-left: 3px solid transparent;
}

.nav-item:hover {
  color: #fff;
  background: rgba(255,255,255,0.06);
}

.nav-item--active {
  color: #409eff;
  background: rgba(64, 158, 255, 0.12);
  border-left-color: #409eff;
}

.main-wrap {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: #f5f7fa;
}

.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  height: 60px;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
  box-shadow: 0 1px 4px rgba(0,0,0,0.06);
  flex-shrink: 0;
}

.topbar__left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.topbar__title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.topbar__right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.admin-name {
  font-size: 14px;
  color: #606266;
}

.hamburger-btn {
  padding: 6px;
}

.page-content {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}

.layout.sidebar-collapsed .sidebar {
  width: 64px;
  min-width: 64px;
}

.layout.sidebar-collapsed .sidebar .brand-name,
.layout.sidebar-collapsed .sidebar .nav-item span {
  display: none;
}

.layout.sidebar-collapsed .sidebar .sidebar__brand {
  justify-content: center;
  padding: 20px 10px;
}

.layout.sidebar-collapsed .sidebar .nav-item {
  justify-content: center;
  padding: 14px 10px;
}

.mobile-overlay {
  display: none;
}

@media (max-width: 767px) {
  .sidebar {
    position: fixed;
    left: 0;
    top: 0;
    height: 100%;
    transform: translateX(-100%);
  }

  .sidebar.mobile-open {
    transform: translateX(0);
  }

  .mobile-overlay {
    display: block;
    position: fixed;
    inset: 0;
    background: rgba(0,0,0,0.4);
    z-index: 99;
  }

  .layout.sidebar-collapsed .sidebar {
    width: 220px;
    min-width: 220px;
  }

  .layout.sidebar-collapsed .sidebar .brand-name,
  .layout.sidebar-collapsed .sidebar .nav-item span {
    display: inline;
  }

  .layout.sidebar-collapsed .sidebar .sidebar__brand {
    justify-content: flex-start;
    padding: 20px 18px;
  }

  .layout.sidebar-collapsed .sidebar .nav-item {
    justify-content: flex-start;
    padding: 12px 20px;
  }

  .page-content {
    padding: 12px;
  }
}
</style>
