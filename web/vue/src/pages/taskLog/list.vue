<template>
  <el-container>
    <task-sidebar></task-sidebar>
    <el-main>
      <el-form :inline="true" >
        <el-form-item :label="t('task.id')">
          <el-input v-model.trim="searchParams.task_id"></el-input>
        </el-form-item>
        <el-form-item :label="t('task.protocol')">
          <el-select v-model.trim="searchParams.protocol" :placeholder="t('task.protocol')" style="width: 180px;">
            <el-option :label="t('message.all')" value=""></el-option>
            <el-option
            v-for="item in protocolList"
            :key="item.value"
            :label="item.label"
            :value="item.value">
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item :label="t('common.status')">
          <el-select v-model.trim="searchParams.status" style="width: 180px;">
            <el-option :label="t('message.all')" value=""></el-option>
            <el-option
              v-for="item in statusList"
              :key="item.value"
              :label="item.label"
              :value="item.value">
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="search()">{{ t('common.search') }}</el-button>
        </el-form-item>
      </el-form>
      <el-row type="flex" justify="end">
        <el-col :span="3">
          <el-button type="danger" v-if="isAdmin" @click="clearLog">{{ t('message.clearLog') }}</el-button>
        </el-col>
        <el-col :span="2">
          <el-button type="info" @click="refresh">{{ t('common.refresh') }}</el-button>
        </el-col>
      </el-row>
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
        <el-table-column type="expand">
          <template #default="scope">
            <el-form label-position="left">
              <el-form-item>
                  {{ t('message.retryCount') }}: {{scope.row.retry_times}} <br>
                  {{ t('task.cronExpression') }}: {{scope.row.spec}} <br>
                  {{ t('task.command') }}: {{scope.row.command}}
              </el-form-item>
            </el-form>
          </template>
        </el-table-column>
        <el-table-column
          prop="id"
          label="ID">
        </el-table-column>
        <el-table-column
          prop="task_id"
          :label="t('task.id')">
        </el-table-column>
        <el-table-column
          prop="name"
          :label="t('task.name')"
        width="180">
        </el-table-column>
        <el-table-column
          prop="protocol"
          :label="t('task.protocol')"
          :formatter="formatProtocol">
        </el-table-column>
        <el-table-column
          :label="t('task.taskNode')"
          width="150">
          <template #default="scope">
            <div v-html="scope.row.hostname"></div>
          </template>
        </el-table-column>
        <el-table-column
          :label="t('taskLog.duration')"
          width="250">
          <template #default="scope">
            {{ t('taskLog.duration') }}: {{scope.row.total_time > 0 ? scope.row.total_time : 1}}{{ t('message.seconds') }}<br>
            {{ t('taskLog.startTime') }}: {{$filters.formatTime(scope.row.start_time)}}<br>
            <span v-if="scope.row.status !== 1">{{ t('taskLog.endTime') }}: {{$filters.formatTime(scope.row.end_time)}}</span>
          </template>
        </el-table-column>
        <el-table-column
          :label="t('common.status')">
          <template #default="scope">
            <span style="color:red" v-if="scope.row.status === 0">{{ t('taskLog.failed') }}</span>
            <span style="color:green" v-else-if="scope.row.status === 1">{{ t('message.running') }}</span>
            <span v-else-if="scope.row.status === 2">{{ t('taskLog.success') }}</span>
            <span style="color:#4499EE" v-else-if="scope.row.status === 3">{{ t('message.cancelled') }}</span>
          </template>
        </el-table-column>
        <el-table-column
          :label="t('taskLog.result')"
          :width="locale === 'zh-CN' ? 120 : 140" v-if="isAdmin">
          <template #default="scope">
            <el-button type="success"
                       size="small"
                       v-if="scope.row.status === 2"
                       @click="showTaskResult(scope.row)">{{ t('taskLog.viewOutput') }}</el-button>
            <el-button type="warning"
                       size="small"
                       v-if="scope.row.status === 0"
                       @click="showTaskResult(scope.row)" >{{ t('taskLog.viewOutput') }}</el-button>
            <el-button type="danger"
                       size="small"
                       v-if="scope.row.status === 1 && scope.row.protocol === 2"
                       @click="stopTask(scope.row)">{{ t('message.stopTask') }}
            </el-button>
          </template>
        </el-table-column>
        <el-table-column
          :label="t('taskLog.result')"
          :width="locale === 'zh-CN' ? 120 : 140" v-else>
          <template #default="scope">
            <el-button type="success"
                       size="small"
                       v-if="scope.row.status === 2"
                       @click="showTaskResult(scope.row)">{{ t('taskLog.viewOutput') }}</el-button>
            <el-button type="warning"
                       size="small"
                       v-if="scope.row.status === 0"
                       @click="showTaskResult(scope.row)" >{{ t('taskLog.viewOutput') }}</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-dialog :title="t('message.taskExecutionResult')" v-model="dialogVisible" width="60%">
        <div>
          <pre>{{currentTaskResult.command}}</pre>
        </div>
        <div>
          <pre>{{currentTaskResult.result}}</pre>
        </div>
      </el-dialog>
    </el-main>
  </el-container>
</template>

<script>
import { useI18n } from 'vue-i18n'
import { ElMessageBox } from 'element-plus'
import taskSidebar from '../task/sidebar.vue'
import taskLogService from '../../api/taskLog'
import { useUserStore } from '../../stores/user'

export default {
  name: 'task-log',
  setup() {
    const { t, locale } = useI18n()
    return { t, locale }
  },
  data () {
    const userStore = useUserStore()
    return {
      logs: [],
      logTotal: 0,
      searchParams: {
        page_size: 20,
        page: 1,
        task_id: '',
        protocol: '',
        status: ''
      },
      isAdmin: userStore.isAdmin,
      dialogVisible: false,
      currentTaskResult: {
        command: '',
        result: ''
      },
      protocolList: [
        {
          value: '1',
          label: 'http'
        },
        {
          value: '2',
          label: 'shell'
        }
      ],
      statusList: []
    }
  },
  computed: {
    computedStatusList() {
      return [
        { value: '1', label: this.t('taskLog.failed') },
        { value: '2', label: this.t('message.running') },
        { value: '3', label: this.t('taskLog.success') },
        { value: '4', label: this.t('message.cancelled') }
      ]
    }
  },
  watch: {
    computedStatusList: {
      handler(newVal) {
        this.statusList = newVal
      },
      immediate: true
    }
  },
  components: {taskSidebar},
  created () {
    if (this.$route.query.task_id) {
      this.searchParams.task_id = this.$route.query.task_id
    }
    this.search()
  },
  activated () {
    this.search()
  },
  methods: {
    formatProtocol (row, col) {
      if (row[col.property] === 1) {
        return 'http'
      }
      return 'shell'
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
      taskLogService.list(this.searchParams, (data) => {
        this.logs = data.data
        this.logTotal = data.total

        if (callback) {
          callback()
        }
      })
    },
    clearLog () {
      ElMessageBox.confirm(this.t('message.confirmClearLog'), this.t('common.tip'), {
        confirmButtonText: this.t('common.confirm'),
        cancelButtonText: this.t('common.cancel'),
        type: 'warning',
        center: true
      }).then(() => {
        taskLogService.clear(() => {
          this.searchParams.page = 1
          this.search()
        })
      }).catch(() => {})
    },
    stopTask (item) {
      taskLogService.stop(item.id, item.task_id, () => {
        this.search()
      })
    },
    showTaskResult (item) {
      this.dialogVisible = true
      this.currentTaskResult.command = item.command
      this.currentTaskResult.result = item.result
    },
    refresh () {
      this.search(() => {
        this.$message.success(this.t('message.refreshSuccess'))
      })
    }
  }
}
</script>
<style scoped>
  pre {
    white-space: pre-wrap;
    word-wrap: break-word;
    padding:10px;
    background-color: #4C4C4C;
    color: white;
  }
</style>
