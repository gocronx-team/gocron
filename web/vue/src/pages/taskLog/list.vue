<template>
  <el-main>
    <el-form :inline="true">
      <el-form-item :label="t('task.id')">
        <el-input v-model.trim="searchParams.task_id" />
      </el-form-item>
      <el-form-item :label="t('task.protocol')">
        <el-select
          v-model.trim="searchParams.protocol"
          :placeholder="t('task.protocol')"
          style="width: 180px"
        >
          <el-option
            :label="t('message.all')"
            value=""
          />
          <el-option
            v-for="item in protocolList"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('common.status')">
        <el-select
          v-model.trim="searchParams.status"
          style="width: 180px"
        >
          <el-option
            :label="t('message.all')"
            value=""
          />
          <el-option
            v-for="item in statusList"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button
          type="primary"
          @click="search()"
        >
          {{ t('common.search') }}
        </el-button>
      </el-form-item>
    </el-form>
    <el-row
      type="flex"
      justify="end"
    >
      <el-col
        v-if="isAdmin && searchParams.task_id"
        :span="4"
      >
        <el-button
          type="warning"
          @click="clearTaskLog"
        >
          {{ t('task.clearTaskLog') }}
        </el-button>
      </el-col>
      <el-col :span="3">
        <el-button
          v-if="isAdmin"
          type="danger"
          @click="clearLog"
        >
          {{
            t('message.clearLog')
          }}
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
      :total="logTotal"
      @size-change="changePageSize"
      @current-change="changePage"
    />
    <el-table
      ref="table"
      :data="logs"
      border
      style="width: 100%"
    >
      <el-table-column type="expand">
        <template #default="scope">
          <el-form label-position="left">
            <el-form-item>
              {{ t('message.retryCount') }}: {{ scope.row.retry_times }} <br>
              {{ t('task.cronExpression') }}: {{ scope.row.spec }} <br>
              {{ t('task.command') }}: {{ scope.row.command }}
            </el-form-item>
          </el-form>
        </template>
      </el-table-column>
      <el-table-column
        prop="id"
        label="ID"
      />
      <el-table-column
        prop="task_id"
        :label="t('task.id')"
      />
      <el-table-column
        prop="name"
        :label="t('task.name')"
        width="180"
      />
      <el-table-column
        prop="protocol"
        :label="t('task.protocol')"
        :formatter="formatProtocol"
      />
      <el-table-column
        :label="t('task.taskNode')"
        width="150"
      >
        <template #default="scope">
          <div v-html="scope.row.hostname" />
        </template>
      </el-table-column>
      <el-table-column
        :label="t('taskLog.duration')"
        width="250"
      >
        <template #default="scope">
          {{ t('taskLog.duration') }}: {{ scope.row.total_time > 0 ? scope.row.total_time : 1
          }}{{ t('message.seconds') }}<br>
          {{ t('taskLog.startTime') }}: {{ $filters.formatTime(scope.row.start_time) }}<br>
          <span v-if="scope.row.status !== 1">{{ t('taskLog.endTime') }}: {{ $filters.formatTime(scope.row.end_time) }}</span>
        </template>
      </el-table-column>
      <el-table-column :label="t('common.status')">
        <template #default="scope">
          <span
            v-if="scope.row.status === 0"
            style="color: red"
          >{{ t('taskLog.failed') }}</span>
          <span
            v-else-if="scope.row.status === 1"
            style="color: green"
          >{{
            t('message.running')
          }}</span>
          <span v-else-if="scope.row.status === 2">{{ t('taskLog.success') }}</span>
          <span
            v-else-if="scope.row.status === 3"
            style="color: #4499ee"
          >{{
            t('message.cancelled')
          }}</span>
        </template>
      </el-table-column>
      <el-table-column
        v-if="isAdmin"
        :label="t('taskLog.result')"
        :width="locale === availableLanguages.zhCN.value ? 120 : 140"
      >
        <template #default="scope">
          <el-button
            v-if="scope.row.status === 2"
            type="success"
            size="small"
            @click="showTaskResult(scope.row)"
          >
            {{ t('taskLog.viewOutput') }}
          </el-button>
          <el-button
            v-if="scope.row.status === 0"
            type="warning"
            size="small"
            @click="showTaskResult(scope.row)"
          >
            {{ t('taskLog.viewOutput') }}
          </el-button>
          <el-button
            v-if="scope.row.status === 3"
            type="info"
            size="small"
            @click="showTaskResult(scope.row)"
          >
            {{ t('taskLog.viewOutput') }}
          </el-button>
          <el-button
            v-if="scope.row.status === 1 && scope.row.protocol === 2"
            type="danger"
            size="small"
            @click="stopTask(scope.row)"
          >
            {{ t('message.stopTask') }}
          </el-button>
        </template>
      </el-table-column>
      <el-table-column
        v-else
        :label="t('taskLog.result')"
        :width="locale === availableLanguages.zhCN.value ? 120 : 140"
      >
        <template #default="scope">
          <el-button
            v-if="scope.row.status === 2"
            type="success"
            size="small"
            @click="showTaskResult(scope.row)"
          >
            {{ t('taskLog.viewOutput') }}
          </el-button>
          <el-button
            v-if="scope.row.status === 0"
            type="warning"
            size="small"
            @click="showTaskResult(scope.row)"
          >
            {{ t('taskLog.viewOutput') }}
          </el-button>
          <el-button
            v-if="scope.row.status === 3"
            type="info"
            size="small"
            @click="showTaskResult(scope.row)"
          >
            {{ t('taskLog.viewOutput') }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-dialog
      v-model="dialogVisible"
      :title="t('message.taskExecutionResult')"
      width="60%"
    >
      <div v-if="currentTaskResult.hostname">
        <strong>{{ t('taskLog.host') }}:</strong>
        <pre v-html="currentTaskResult.hostname" />
      </div>
      <div>
        <strong>{{ t('task.command') }}:</strong>
        <pre>{{ currentTaskResult.command }}</pre>
      </div>
      <div>
        <strong>{{ t('taskLog.output') }}:</strong>
        <pre>{{ currentTaskResult.result }}</pre>
      </div>
    </el-dialog>
  </el-main>
</template>

<script>
import { useI18n } from 'vue-i18n'
import { useNotify } from '@/composables/useNotify'
import taskLogService from '../../api/taskLog'
import { useUserStore } from '../../stores/user'
import { availableLanguages } from '@/const/index'

export default {
  name: 'TaskLog',
  setup() {
    const { t, locale } = useI18n()
    return { t, locale, availableLanguages }
  },
  data() {
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
        hostname: '',
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
        { value: '3', label: this.t('taskLog.success') },
        { value: '1', label: this.t('taskLog.failed') },
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
    },
    '$route.query.task_id': {
      handler(newTaskId) {
        if (newTaskId !== undefined) {
          this.searchParams.task_id = newTaskId
          this.searchParams.page = 1
          this.search()
        }
      }
    }
  },
  created() {
    this.updateTaskIdFromRoute()
    this.search()
  },
  activated() {
    this.updateTaskIdFromRoute()
    this.search()
  },
  methods: {
    formatProtocol(row, col) {
      if (row[col.property] === 1) {
        return 'http'
      }
      return 'shell'
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
      taskLogService.list(this.searchParams, data => {
        this.logs = data.data
        this.logTotal = data.total

        if (callback) {
          callback()
        }
      })
    },
    async clearTaskLog() {
      const taskId = this.searchParams.task_id
      const notify = useNotify()
      if (await notify.confirm(this.t('task.confirmClearTaskLog', { taskId }), this.t('common.tip'))) {
        taskLogService.clearByTaskId(taskId, () => {
          this.searchParams.page = 1
          this.search()
        })
      }
    },
    async clearLog() {
      const notify = useNotify()
      if (await notify.confirm(this.t('message.confirmClearLog'), this.t('common.tip'))) {
        taskLogService.clear(() => {
          this.searchParams.page = 1
          this.search()
        })
      }
    },
    stopTask(item) {
      taskLogService.stop(item.id, item.task_id, () => {
        this.search()
      })
    },
    showTaskResult(item) {
      this.dialogVisible = true
      // 清理命令中的 HTML 实体编码
      let cleanedCommand = item.command
      if (cleanedCommand) {
        cleanedCommand = cleanedCommand
          .replace(/&quot;/g, '"')
          .replace(/&apos;/g, "'")
          .replace(/&#39;/g, "'")
          .replace(/&lt;/g, '<')
          .replace(/&gt;/g, '>')
          .replace(/&amp;/g, '&')
      }
      this.currentTaskResult.hostname = item.hostname || ''
      this.currentTaskResult.command = cleanedCommand
      this.currentTaskResult.result = item.result
    },
    refresh() {
      this.search(() => {
        this.$message.success(this.t('message.refreshSuccess'))
      })
    },
    updateTaskIdFromRoute() {
      if (this.$route.query.task_id) {
        this.searchParams.task_id = this.$route.query.task_id
        this.searchParams.page = 1
      }
    }
  }
}
</script>
<style scoped>
pre {
  white-space: pre-wrap;
  word-wrap: break-word;
  padding: 10px;
  background-color: #4c4c4c;
  color: white;
}
</style>
