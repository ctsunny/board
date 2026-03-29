<template>
  <div>
    <PageHeader title="直播地区">
      <el-button type="primary" :icon="Plus" @click="openCreate">新增地区</el-button>
    </PageHeader>

    <el-card>
      <el-table v-loading="loading" :data="list" style="width:100%">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="name" label="名称" min-width="120" />
        <el-table-column prop="code" label="代码" width="100" />
        <el-table-column prop="remark" label="备注" show-overflow-tooltip />
        <el-table-column label="操作" width="130" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" text @click="openEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" text @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editRow ? '编辑地区' : '新增地区'" width="420px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-width="70px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="地区名称" />
        </el-form-item>
        <el-form-item label="代码" prop="code">
          <el-input v-model="form.code" placeholder="如 CN-BJ" />
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
import { regionsApi } from '@/api'
import PageHeader from '@/components/PageHeader.vue'
import type { FormInstance, FormRules } from 'element-plus'

const loading = ref(false)
const saving = ref(false)
const list = ref<Record<string, unknown>[]>([])
const dialogVisible = ref(false)
const editRow = ref<Record<string, unknown> | null>(null)
const formRef = ref<FormInstance>()

const form = reactive({ name: '', code: '', remark: '' })
const rules: FormRules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
}

async function loadData() {
  loading.value = true
  try {
    const res = await regionsApi.list()
    const d = res.data as { list?: unknown[]; items?: unknown[] } | unknown[]
    list.value = (Array.isArray(d) ? d : ((d as { list?: unknown[] }).list ?? [])) as Record<string, unknown>[]
  } catch {
    list.value = []
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editRow.value = null
  Object.assign(form, { name: '', code: '', remark: '' })
  dialogVisible.value = true
}

function openEdit(row: Record<string, unknown>) {
  editRow.value = row
  Object.assign(form, { name: row.name ?? '', code: row.code ?? '', remark: row.remark ?? '' })
  dialogVisible.value = true
}

async function handleSave() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    saving.value = true
    try {
      if (editRow.value) {
        await regionsApi.update(editRow.value.id as number, { ...form })
      } else {
        await regionsApi.create({ ...form })
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
  await ElMessageBox.confirm(`确认删除地区「${row.name}」？`, '提示', { type: 'warning' })
  try {
    await regionsApi.delete(row.id as number)
    ElMessage.success('已删除')
    loadData()
  } catch {
    ElMessage.error('删除失败')
  }
}

onMounted(loadData)
</script>
