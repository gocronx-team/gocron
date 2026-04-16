<template>
  <el-main>
    <el-form :inline="true" label-width="auto">
      <el-form-item :label="t('template.category')">
        <el-select v-model="searchParams.category" style="width: 150px;" @change="search">
          <el-option :label="t('template.category_all')" value=""></el-option>
          <el-option
            v-for="cat in categoryList"
            :key="cat"
            :label="getCategoryLabel(cat)"
            :value="cat">
          </el-option>
        </el-select>
      </el-form-item>
      <el-form-item :label="t('template.name')">
        <el-input v-model.trim="searchParams.name" style="width: 180px;" @keyup.enter="search"></el-input>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="search">{{ t('common.search') }}</el-button>
      </el-form-item>
    </el-form>
    <el-row type="flex" justify="end" style="margin-bottom: 10px;">
      <el-col :span="24" style="text-align: right;">
        <el-button type="primary" @click="toEdit(null)" v-if="isAdmin">{{ t('template.createNew') }}</el-button>
        <el-button type="info" @click="refresh">{{ t('common.refresh') }}</el-button>
      </el-col>
    </el-row>
    <el-pagination
      background
      layout="prev, pager, next, sizes, total"
      :total="total"
      v-model:current-page="searchParams.page"
      v-model:page-size="searchParams.page_size"
      @size-change="changePageSize"
      @current-change="changePage">
    </el-pagination>
    <el-table :data="templates" border style="width: 100%">
      <el-table-column type="expand">
        <template #default="scope">
          <div style="padding: 12px;">
            <strong>{{ t('template.command') }}:</strong>
            <pre style="white-space: pre-wrap; word-break: break-all; background: #f5f7fa; padding: 12px; border-radius: 4px; margin-top: 8px;">{{ scope.row.command }}</pre>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="id" label="ID" width="60" align="center"></el-table-column>
      <el-table-column prop="name" :label="t('template.name')" min-width="150" align="center">
        <template #default="scope">
          {{ scope.row.name }}
          <el-tag v-if="scope.row.is_builtin === 1" size="small" type="info" style="margin-left: 4px;">{{ t('template.builtin') }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="description" :label="t('template.description')" min-width="200" align="center"></el-table-column>
      <el-table-column prop="category" :label="t('template.category')" width="100" align="center">
        <template #default="scope">
          <el-tag size="small">{{ getCategoryLabel(scope.row.category) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="t('template.protocol')" width="80" align="center">
        <template #default="scope">
          {{ scope.row.protocol === 1 ? 'HTTP' : 'Shell' }}
        </template>
      </el-table-column>
      <el-table-column :label="t('common.operation')" width="180" align="center" v-if="isAdmin">
        <template #default="scope">
          <div style="display: flex; flex-direction: column; gap: 4px;">
            <el-button type="primary" size="small" style="width: 100%;" @click="useTemplate(scope.row)">{{ t('template.useTemplate') }}</el-button>
            <div v-if="scope.row.is_builtin !== 1" style="display: flex; gap: 4px;">
              <el-button size="small" style="flex: 1;" @click="toEdit(scope.row)">{{ t('common.edit') }}</el-button>
              <el-button type="danger" size="small" style="flex: 1;" @click="remove(scope.row)">{{ t('common.delete') }}</el-button>
            </div>
          </div>
        </template>
      </el-table-column>
    </el-table>
  </el-main>
</template>

<script>
import { useI18n } from 'vue-i18n'
import templateService from '../../api/template'
import { useUserStore } from '../../stores/user'
import { ElMessageBox } from 'element-plus'

export default {
  name: 'template-list',
  setup() {
    const { t, locale } = useI18n()
    return { t, locale }
  },
  data() {
    const userStore = useUserStore()
    return {
      templates: [],
      total: 0,
      categoryList: [],
      isAdmin: userStore.isAdmin,
      searchParams: {
        page: 1,
        page_size: 20,
        category: '',
        name: ''
      },
      isFirstActivate: true
    }
  },
  created() {
    this.loadCategories()
    this.search()
  },
  activated() {
    if (this.isFirstActivate) {
      this.isFirstActivate = false
      return
    }
    this.loadCategories()
    this.search()
  },
  methods: {
    getCategoryLabel(cat) {
      const key = `template.category_${cat}`
      const label = this.t(key)
      return label === key ? cat : label
    },
    loadCategories() {
      templateService.categories((data) => {
        this.categoryList = data || []
      })
    },
    search() {
      templateService.list(this.searchParams, (data) => {
        this.templates = data.data || []
        this.total = data.total || 0
      })
    },
    changePage(page) {
      this.searchParams.page = page
      this.search()
    },
    changePageSize(size) {
      this.searchParams.page_size = size
      this.search()
    },
    refresh() {
      this.search()
      this.$message.success(this.t('message.refreshSuccess'))
    },
    toEdit(item) {
      if (item === null) {
        this.$router.push('/template/create')
      } else {
        this.$router.push(`/template/edit/${item.id}`)
      }
    },
    useTemplate(row) {
      this.$router.push({
        path: '/task/create',
        query: { template_id: row.id }
      })
    },
    remove(row) {
      ElMessageBox.confirm(
        this.t('template.confirmDelete', { name: row.name }),
        this.t('common.tip'),
        {
          confirmButtonText: this.t('common.confirm'),
          cancelButtonText: this.t('common.cancel'),
          type: 'warning'
        }
      ).then(() => {
        templateService.remove(row.id, () => {
          this.$message.success(this.t('message.deleteSuccess'))
          this.search()
        })
      }).catch(() => {})
    }
  }
}
</script>
