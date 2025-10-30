<template>
  <el-container>
    <system-sidebar></system-sidebar>
    <el-main>
      <el-pagination
        background
        layout="prev, pager, next, sizes, total"
        :total="logTotal"
        v-model:current-page="searchParams.page"
        v-model:page-size="searchParams.page_size"
        @size-change="changePageSize"
        @current-change="changePage">
      </el-pagination>
      <el-table
        :data="logs"
        border
        ref="table"
        style="width: 100%">
        <el-table-column
          prop="id"
          label="ID">
        </el-table-column>
        <el-table-column
          prop="username"
          :label="t('user.username')">
        </el-table-column>
        <el-table-column
          prop="ip"
          :label="t('system.loginIp')">
        </el-table-column>
        <el-table-column
          :label="t('system.loginTime')"
          width="">
          <template #default="scope">
            {{ $filters.formatTime(scope.row.created) }}
          </template>
        </el-table-column>
      </el-table>
    </el-main>
  </el-container>
</template>

<script>
import { useI18n } from 'vue-i18n'
import systemSidebar from './sidebar.vue'
import systemService from '../../api/system'
export default {
  name: 'login-log',
  setup() {
    const { t } = useI18n()
    return { t }
  },
  data () {
    return {
      logs: [],
      logTotal: 0,
      searchParams: {
        page_size: 20,
        page: 1
      }
    }
  },
  created () {
    this.search()
  },
  components: {systemSidebar},
  methods: {
    changePage (page) {
      this.searchParams.page = page
      this.search()
    },
    changePageSize (pageSize) {
      this.searchParams.page_size = pageSize
      this.search()
    },
    search () {
      systemService.loginLogList(this.searchParams, (data) => {
        this.logs = data.data
        this.logTotal = data.total
      })
    }
  }
}
</script>
