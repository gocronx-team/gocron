<template>
  <el-container>
    <el-main>
      <el-form :inline="true" >
        <el-row>
          <el-form-item label="ID">
            <el-input v-model.trim="searchParams.id"></el-input>
          </el-form-item>
          <el-form-item :label="t('host.name')">
            <el-input v-model.trim="searchParams.name"></el-input>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="search()">{{ t('common.search') }}</el-button>
          </el-form-item>
        </el-row>
      </el-form>
      <el-row type="flex" justify="end">
        <el-col :span="2">
          <el-button type="primary" v-if="isAdmin"  @click="toEdit(null)">{{ t('common.add') }}</el-button>
        </el-col>
        <el-col :span="2">
          <el-button type="info" @click="refresh">{{ t('common.refresh') }}</el-button>
        </el-col>
      </el-row>
      <el-pagination
        background
        layout="prev, pager, next, sizes, total"
        :total="hostTotal"
        v-model:current-page="searchParams.page"
        v-model:page-size="searchParams.page_size"
        @size-change="changePageSize"
        @current-change="changePage">
      </el-pagination>
      <el-table
        :data="hosts"
        tooltip-effect="dark"
        border
        style="width: 100%">
        <el-table-column
          prop="id"
          label="ID">
        </el-table-column>
        <el-table-column
          prop="alias"
          :label="t('host.alias')">
        </el-table-column>
        <el-table-column
          prop="name"
          :label="t('host.name')">
        </el-table-column>
        <el-table-column
          prop="port"
          :label="t('host.port')">
        </el-table-column>
        <el-table-column :label="t('task.viewLog')">
          <template #default="scope">
            <el-button type="success" @click="toTasks(scope.row)">{{ t('task.list') }}</el-button>
          </template>
        </el-table-column>
        <el-table-column
          prop="remark"
          :label="t('host.remark')">
        </el-table-column>
        <el-table-column :label="t('common.operation')" :width="locale === 'zh-CN' ? 260 : 300" v-if="this.isAdmin">
          <template #default="scope">
            <el-button type="primary" size="small" @click="toEdit(scope.row)">{{ t('common.edit') }}</el-button>
            <el-button type="info" size="small" @click="ping(scope.row)">{{ t('system.testSend') }}</el-button>
            <el-button type="danger" size="small" @click="remove(scope.row)">{{ t('common.delete') }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-main>
  </el-container>
</template>

<script>
import { useI18n } from 'vue-i18n'
import { ElMessageBox } from 'element-plus'
import hostService from '../../api/host'
import { useUserStore } from '../../stores/user'

export default {
  name: 'host-list',
  setup() {
    const { t, locale } = useI18n()
    return { t, locale }
  },
  data () {
    const userStore = useUserStore()
    return {
      hosts: [],
      hostTotal: 0,
      searchParams: {
        page_size: 20,
        page: 1,
        id: '',
        name: '',
        alias: ''
      },
      isAdmin: userStore.isAdmin
    }
  },
  created () {
    this.search()
  },
  methods: {
    changePage (page) {
      this.searchParams.page = page
      this.search()
    },
    changePageSize (pageSize) {
      this.searchParams.page_size = pageSize
      this.search()
    },
    search (callback = null) {
      hostService.list(this.searchParams, (data) => {
        this.hosts = data.data
        this.hostTotal = data.total
        if (callback) {
          callback()
        }
      })
    },
    remove (item) {
      ElMessageBox.confirm(this.t('message.confirmDeleteNode'), this.t('common.tip'), {
        confirmButtonText: this.t('common.confirm'),
        cancelButtonText: this.t('common.cancel'),
        type: 'warning',
        center: true
      }).then(() => {
        hostService.remove(item.id, () => this.refresh())
      }).catch(() => {})
    },
    ping (item) {
      hostService.ping(item.id, () => {
        this.$message.success(this.t('message.connectionSuccess'))
      })
    },
    toEdit (item) {
      let path = ''
      if (item === null) {
        path = '/host/create'
      } else {
        path = `/host/edit/${item.id}`
      }
      this.$router.push(path)
    },
    refresh () {
      this.search(() => {
        this.$message.success(this.t('message.refreshSuccess'))
      })
    },
    toTasks (item) {
      this.$router.push(
        {
          path: '/task',
          query: {
            host_id: item.id
          }
        })
    }
  }
}
</script>
