<template>
  <el-main>
    <el-row
      type="flex"
      justify="end"
    >
      <el-col :span="2">
        <el-button
          type="primary"
          @click="toEdit(null)"
        >
          {{ t('common.add') }}
        </el-button>
      </el-col>
      <el-col :span="2">
        <el-button
          type="info"
          @click="refresh"
        >
          {{ t('common.refresh') }}
        </el-button>
      </el-col>
    </el-row>
    <el-pagination
      v-model:current-page="searchParams.page"
      v-model:page-size="searchParams.page_size"
      background
      layout="prev, pager, next, sizes, total"
      :total="userTotal"
      @size-change="changePageSize"
      @current-change="changePage"
    />
    <el-table
      :data="users"
      tooltip-effect="dark"
      border
      style="width: 100%"
    >
      <el-table-column
        prop="id"
        label="ID"
      />
      <el-table-column
        prop="name"
        :label="t('user.username')"
      />
      <el-table-column
        prop="email"
        :label="t('user.email')"
      />
      <el-table-column
        prop="is_admin"
        :formatter="formatRole"
        :label="t('user.role')"
      />
      <el-table-column :label="t('common.status')">
        <template #default="scope">
          <el-switch
            v-model="scope.row.status"
            :active-value="1"
            :inactive-value="0"
            active-color="#13ce66"
            inactive-color="#ff4949"
            @change="changeStatus(scope.row)"
          />
        </template>
      </el-table-column>
      <el-table-column
        v-if="isAdmin"
        :label="t('common.operation')"
        :width="locale === availableLanguages.zhCN.value ? 280 : 340"
      >
        <template #default="scope">
          <el-button
            type="primary"
            size="small"
            @click="toEdit(scope.row)"
          >
            {{
              t('common.edit')
            }}
          </el-button>
          <el-button
            type="success"
            size="small"
            @click="editPassword(scope.row)"
          >
            {{
              t('user.changePassword')
            }}
          </el-button>
          <el-button
            type="danger"
            size="small"
            @click="remove(scope.row)"
          >
            {{
              t('common.delete')
            }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-main>
</template>

<script>
import { useI18n } from 'vue-i18n'
import { useNotify } from '@/composables/useNotify'
import userService from '../../api/user'
import { useUserStore } from '../../stores/user'
import { availableLanguages } from '@/const/lang'

export default {
  name: 'UserList',
  setup() {
    const { t, locale } = useI18n()
    return { t, locale, availableLanguages }
  },
  data() {
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
  mounted() {
    this.search()
  },
  methods: {
    changeStatus(item) {
      if (item.status) {
        userService.enable(item.id)
      } else {
        userService.disable(item.id)
      }
    },
    formatRole(row, col) {
      if (row[col.property] === 1) {
        return this.t('user.admin')
      }
      return this.t('user.normalUser')
    },
    changePage(page) {
      this.searchParams.page = page
      this.search()
    },
    changePageSize(pageSize) {
      this.searchParams.page_size = pageSize
      this.search()
    },
    search(callback = null) {
      userService.list(this.searchParams, data => {
        this.users = data.data
        this.userTotal = data.total
        if (callback) {
          callback()
        }
      })
    },
    async remove(item) {
      const notify = useNotify()
      if (await notify.confirm(this.t('message.confirmDeleteUser'), this.t('common.tip'))) {
        userService.remove(item.id, () => {
          this.refresh()
        })
      }
    },
    toEdit(item) {
      let path = ''
      if (item === null) {
        path = '/user/create'
      } else {
        path = `/user/edit/${item.id}`
      }
      this.$router.push(path)
    },
    refresh() {
      this.search(() => {
        this.$message.success(this.t('message.refreshSuccess'))
      })
    },
    editPassword(item) {
      this.$router.push(`/user/edit-password/${item.id}`)
    }
  }
}
</script>
