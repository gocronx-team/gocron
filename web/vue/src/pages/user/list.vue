<template>
  <el-container>
    <el-main>
      <el-row type="flex" justify="end">
        <el-col :span="2">
          <el-button type="primary"  @click="toEdit(null)">{{ t('common.add') }}</el-button>
        </el-col>
        <el-col :span="2">
          <el-button type="info" @click="refresh">{{ t('common.refresh') }}</el-button>
        </el-col>
      </el-row>
      <el-pagination
        background
        layout="prev, pager, next, sizes, total"
        :total="userTotal"
        v-model:current-page="searchParams.page"
        v-model:page-size="searchParams.page_size"
        @size-change="changePageSize"
        @current-change="changePage">
      </el-pagination>
      <el-table
        :data="users"
        tooltip-effect="dark"
        border
        style="width: 100%">
        <el-table-column
          prop="id"
          label="ID">
        </el-table-column>
        <el-table-column
          prop="name"
          :label="t('user.username')">
        </el-table-column>
        <el-table-column
          prop="email"
          :label="t('user.email')">
        </el-table-column>
        <el-table-column
          prop="is_admin"
          :formatter="formatRole"
          :label="t('user.role')">
        </el-table-column>
        <el-table-column
          :label="t('common.status')">
          <template #default="scope">
            <el-switch
              v-model="scope.row.status"
              :active-value="1"
              :inactive-value="0"
              active-color="#13ce66"
              @change="changeStatus(scope.row)"
              inactive-color="#ff4949">
            </el-switch>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.operation')" :width="locale === 'zh-CN' ? 280 : 340" v-if="this.isAdmin">
          <template #default="scope">
            <el-button type="primary" size="small" @click="toEdit(scope.row)">{{ t('common.edit') }}</el-button>
            <el-button type="success" size="small" @click="editPassword(scope.row)">{{ t('user.changePassword') }}</el-button>
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
import userService from '../../api/user'
import { useUserStore } from '../../stores/user'

export default {
  name: 'user-list',
  setup() {
    const { t, locale } = useI18n()
    return { t, locale }
  },
  data () {
    const userStore = useUserStore()
    return {
      users: [],
      userTotal: 0,
      searchParams: {
        page_size: 20,
        page: 1
      },
      isAdmin: userStore.isAdmin
    }
  },
  created () {
    this.search()
  },
  methods: {
    changeStatus (item) {
      if (item.status) {
        userService.enable(item.id)
      } else {
        userService.disable(item.id)
      }
    },
    formatRole (row, col) {
      if (row[col.property] === 1) {
        return this.t('user.admin')
      }
      return this.t('user.normalUser')
    },
    changePage (page) {
      this.searchParams.page = page
      this.search()
    },
    changePageSize (pageSize) {
      this.searchParams.page_size = pageSize
      this.search()
    },
    search (callback = null) {
      userService.list(this.searchParams, (data) => {
        this.users = data.data
        this.userTotal = data.total
        if (callback) {
          callback()
        }
      })
    },
    remove (item) {
      ElMessageBox.confirm(this.t('message.confirmDeleteUser'), this.t('common.tip'), {
        confirmButtonText: this.t('common.confirm'),
        cancelButtonText: this.t('common.cancel'),
        type: 'warning',
        center: true
      }).then(() => {
        userService.remove(item.id, () => {
          this.refresh()
        })
      }).catch(() => {})
    },
    toEdit (item) {
      let path = ''
      if (item === null) {
        path = '/user/create'
      } else {
        path = `/user/edit/${item.id}`
      }
      this.$router.push(path)
    },
    refresh () {
      this.search(() => {
        this.$message.success(this.t('message.refreshSuccess'))
      })
    },
    editPassword (item) {
      this.$router.push(`/user/edit-password/${item.id}`)
    }
  }
}
</script>
