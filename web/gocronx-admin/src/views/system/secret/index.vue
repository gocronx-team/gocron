<!-- 机密管理：类 GitHub Secrets，值写入后不可读取，仅任务执行时按需注入 -->
<template>
  <div class="secret-page art-full-height">
    <ElCard class="art-table-card" shadow="never">
      <ElAlert :closable="false" type="info" show-icon style="margin-bottom: 16px">
        <template #title>{{ t('secret.introTitle') }}</template>
        <div class="intro-body">
          <p>{{ t('secret.introDesc') }}</p>
          <p class="warn">{{ t('secret.keyHint') }}</p>
        </div>
      </ElAlert>

      <div class="toolbar">
        <span class="text-base font-medium">{{ t('menus.system.secret') }}</span>
        <div>
          <ElButton :loading="loading" @click="loadList">{{ t('secret.refresh') }}</ElButton>
          <ElButton type="primary" @click="openCreate">{{ t('secret.create') }}</ElButton>
        </div>
      </div>

      <ElTable v-loading="loading" :data="list" border style="width: 100%">
        <ElTableColumn type="index" label="#" width="60" align="center" />
        <ElTableColumn prop="name" :label="t('secret.name')" align="center" />
        <ElTableColumn
          prop="remark"
          :label="t('secret.remark')"
          align="center"
          show-overflow-tooltip
        />
        <ElTableColumn :label="t('secret.updatedAt')" width="190" align="center">
          <template #default="{ row }">{{ formatDateTime(row.updated_at) }}</template>
        </ElTableColumn>
        <ElTableColumn :label="t('secret.operation')" width="170" align="center">
          <template #default="{ row }">
            <ElButton type="primary" size="small" @click="openEdit(row)">
              {{ t('secret.edit') }}
            </ElButton>
            <ElButton type="danger" size="small" @click="handleRemove(row)">
              {{ t('secret.delete') }}
            </ElButton>
          </template>
        </ElTableColumn>
      </ElTable>
    </ElCard>

    <!-- Create / Edit dialog -->
    <ElDialog v-model="dialogVisible" :title="dialogTitle" width="480px" align-center>
      <ElForm ref="formRef" :model="form" :rules="rules" label-width="90px" @submit.prevent>
        <ElFormItem :label="t('secret.name')" prop="name">
          <ElInput
            v-model.trim="form.name"
            :placeholder="t('secret.namePlaceholder')"
            maxlength="64"
            clearable
          />
        </ElFormItem>
        <ElFormItem :label="t('secret.value')" prop="value">
          <ElInput
            v-model="form.value"
            type="password"
            :placeholder="isEdit ? t('secret.valueEditPlaceholder') : t('secret.valuePlaceholder')"
            show-password
            clearable
          />
        </ElFormItem>
        <ElFormItem :label="t('secret.remark')" prop="remark">
          <ElInput
            v-model.trim="form.remark"
            :placeholder="t('secret.remarkPlaceholder')"
            maxlength="255"
            clearable
          />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="dialogVisible = false">{{ t('secret.cancel') }}</ElButton>
        <ElButton type="primary" :loading="submitting" @click="submit">
          {{ t('secret.confirm') }}
        </ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, computed, onMounted } from 'vue'
  import { useI18n } from 'vue-i18n'
  import {
    ElButton,
    ElCard,
    ElTable,
    ElTableColumn,
    ElDialog,
    ElForm,
    ElFormItem,
    ElInput,
    ElAlert,
    ElMessage,
    ElMessageBox,
    type FormInstance,
    type FormRules
  } from 'element-plus'
  import { fetchSecretList, storeSecret, removeSecret, type SecretItem } from '@/api/secret'
  import { formatDateTime } from '@/utils/date'

  defineOptions({ name: 'Secret' })

  const { t } = useI18n()

  const list = ref<SecretItem[]>([])
  const loading = ref(false)

  const dialogVisible = ref(false)
  const submitting = ref(false)
  const formRef = ref<FormInstance>()
  const form = reactive({ id: 0, name: '', value: '', remark: '' })

  const isEdit = computed(() => form.id > 0)
  const dialogTitle = computed(() => (isEdit.value ? t('secret.editTitle') : t('secret.create')))

  // 合法环境变量名：字母/下划线开头，后接字母数字下划线（与后端校验一致）
  const namePattern = /^[A-Za-z_][A-Za-z0-9_]*$/

  const rules = computed<FormRules>(() => ({
    name: [
      { required: true, message: t('secret.nameRequired'), trigger: 'blur' },
      {
        validator: (_r: unknown, value: string, cb: (e?: Error) => void) => {
          if (value && !namePattern.test(value)) {
            cb(new Error(t('secret.nameInvalid')))
          } else {
            cb()
          }
        },
        trigger: 'blur'
      }
    ],
    // 创建时值必填；编辑时留空表示保持原值
    value: isEdit.value
      ? []
      : [{ required: true, message: t('secret.valueRequired'), trigger: 'blur' }]
  }))

  async function loadList() {
    loading.value = true
    try {
      const res = await fetchSecretList()
      list.value = res?.data || []
    } catch {
      // error toast handled by http util
    } finally {
      loading.value = false
    }
  }

  function resetForm() {
    form.id = 0
    form.name = ''
    form.value = ''
    form.remark = ''
    formRef.value?.clearValidate()
  }

  function openCreate() {
    resetForm()
    dialogVisible.value = true
  }

  function openEdit(row: SecretItem) {
    resetForm()
    form.id = row.id
    form.name = row.name
    form.remark = row.remark
    dialogVisible.value = true
  }

  async function submit() {
    if (!formRef.value) return
    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    submitting.value = true
    try {
      await storeSecret({
        id: form.id || undefined,
        name: form.name,
        value: form.value || undefined,
        remark: form.remark
      })
      ElMessage.success(isEdit.value ? t('secret.updateSuccess') : t('secret.createSuccess'))
      dialogVisible.value = false
      loadList()
    } catch {
      // error toast handled by http util
    } finally {
      submitting.value = false
    }
  }

  async function handleRemove(row: SecretItem) {
    try {
      await ElMessageBox.confirm(
        t('secret.confirmDelete', { name: row.name }),
        t('secret.delete'),
        {
          confirmButtonText: t('secret.confirm'),
          cancelButtonText: t('secret.cancel'),
          type: 'warning',
          center: true
        }
      )
    } catch {
      return
    }
    try {
      await removeSecret(row.id)
      ElMessage.success(t('secret.deleteSuccess'))
      loadList()
    } catch {
      // error toast handled by http util
    }
  }

  onMounted(loadList)
</script>

<style scoped>
  .secret-page {
    display: flex;
    flex-direction: column;
  }

  .intro-body {
    font-size: 13px;
    line-height: 1.7;
  }

  .warn {
    margin: 4px 0 0;
    color: var(--el-color-warning);
  }

  .toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 14px;
  }
</style>
