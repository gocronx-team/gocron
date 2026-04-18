<template>
  <el-main>
    <el-form
      :inline="true"
      label-width="auto"
    >
      <el-form-item :label="t('task.id')">
        <el-input
          v-model.trim="searchParams.id"
          style="width: 180px;"
        />
      </el-form-item>
      <el-form-item :label="t('task.name')">
        <el-input
          v-model.trim="searchParams.name"
          style="width: 180px;"
        />
      </el-form-item>
      <el-form-item :label="t('task.tag')">
        <el-select
          v-model="searchParams.selectedTags"
          multiple
          filterable
          allow-create
          default-first-option
          collapse-tags
          collapse-tags-tooltip
          :placeholder="t('task.tagPlaceholder')"
          style="width: 180px;"
        >
          <el-option
            v-for="tag in tagOptions"
            :key="tag"
            :label="tag"
            :value="tag"
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
      <br>
      <el-form-item :label="t('task.protocol')">
        <el-select
          v-model.trim="searchParams.protocol"
          style="width: 180px;"
        >
          <el-option
            :label="t('select')"
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
      <el-form-item :label="t('task.taskNode')">
        <el-select
          v-model.trim="searchParams.host_id"
          style="width: 180px;"
        >
          <el-option
            :label="t('select')"
            value=""
          />
          <el-option
            v-for="item in hosts"
            :key="item.id"
            :label="item.alias + ' - ' + item.name + ':' + item.port "
            :value="item.id"
          />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('common.status')">
        <el-select
          v-model.trim="searchParams.status"
          style="width: 180px;"
        >
          <el-option
            :label="t('select')"
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
    </el-form>
    <el-row
      type="flex"
      justify="end"
      style="margin-bottom: 10px;"
    >
      <el-col
        :span="24"
        style="text-align: right;"
      >
        <span
          v-if="isAdmin && selectedTasks.length > 0"
          style="margin-right: 10px; color: #909399;"
        >{{ t('message.selected') }} {{ selectedTasks.length }} {{ t('message.tasks') }}</span>
        <el-button
          v-if="isAdmin"
          type="success"
          size="default"
          :disabled="selectedTasks.length === 0"
          @click="batchEnable"
        >
          {{ t('message.batchEnable') }}
        </el-button>
        <el-button
          v-if="isAdmin"
          type="warning"
          size="default"
          :disabled="selectedTasks.length === 0"
          @click="batchDisable"
        >
          {{ t('message.batchDisable') }}
        </el-button>
        <el-button
          v-if="isAdmin"
          type="danger"
          size="default"
          :disabled="selectedTasks.length === 0"
          @click="batchRemove"
        >
          {{ t('message.batchDelete') }}
        </el-button>
        <el-button
          v-if="isAdmin"
          type="primary"
          @click="toEdit(null)"
        >
          {{ t('common.add') }}
        </el-button>
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
      :total="taskTotal"
      @size-change="changePageSize"
      @current-change="changePage"
    />
    <el-table
      :data="tasks"
      tooltip-effect="dark"
      border
      style="width: 100%"
      @selection-change="handleSelectionChange"
    >
      <el-table-column
        v-if="isAdmin"
        type="selection"
        width="55"
      />
      <el-table-column type="expand">
        <template #default="scope">
          <el-form
            label-position="left"
            inline
            class="demo-table-expand"
            label-width="auto"
          >
            <el-form-item :label="t('message.taskCreatedTime') + ':'">
              {{ $filters.formatTime(scope.row.created) }} <br>
            </el-form-item>
            <el-form-item :label="t('message.taskType') + ':'">
              {{ formatLevel(scope.row.level) }} <br>
            </el-form-item>
            <el-form-item :label="t('message.singleInstanceRun') + ':'">
              {{ formatMulti(scope.row.multi) }} <br>
            </el-form-item>
            <el-form-item :label="t('message.timeoutTime') + ':'">
              {{ formatTimeout(scope.row.timeout) }} <br>
            </el-form-item>
            <el-form-item :label="t('message.retryCount') + ':'">
              {{ scope.row.retry_times }} <br>
            </el-form-item>
            <el-form-item :label="t('message.retryIntervalTime') + ':'">
              {{ formatRetryTimesInterval(scope.row.retry_interval) }}
            </el-form-item> <br>
            <el-form-item :label="t('message.taskNodeLabel')">
              <div
                v-for="item in scope.row.hosts"
                :key="item.host_id"
              >
                {{ item.alias }} - {{ item.name }}:{{ item.port }} <br>
              </div>
            </el-form-item> <br>
            <el-form-item
              :label="t('message.commandLabel') + ':'"
              style="width: 100%"
            >
              {{ scope.row.command }}
            </el-form-item> <br>
            <el-form-item
              :label="t('message.remarkLabel')"
              style="width: 100%"
            >
              {{ scope.row.remark }}
            </el-form-item>
          </el-form>
        </template>
      </el-table-column>
      <el-table-column
        prop="id"
        :label="t('task.id')"
      />
      <el-table-column
        prop="name"
        :label="t('task.name')"
        width="150"
      />
      <el-table-column
        :label="t('task.tag')"
      >
        <template #default="scope">
          <template v-if="scope.row.tag">
            <el-tag
              v-for="tag in scope.row.tag.split(',').filter(Boolean)"
              :key="tag"
              size="small"
              style="margin-right: 4px; margin-bottom: 2px;"
            >
              {{ tag }}
            </el-tag>
          </template>
        </template>
      </el-table-column>
      <el-table-column
        :label="t('task.cronExpression')"
        min-width="150"
        class-name="no-wrap-header"
      >
        <template #default="scope">
          <span>{{ parseCronSpec(scope.row.spec).expr }}</span>
          <div
            v-if="parseCronSpec(scope.row.spec).tz"
            style="color: #909399; font-size: 12px; line-height: 1.4;"
          >
            {{ parseCronSpec(scope.row.spec).tz }}
          </div>
        </template>
      </el-table-column>
      <el-table-column
        :label="t('task.nextRunTime')"
        width="180"
        class-name="no-wrap-header"
      >
        <template #default="scope">
          {{ $filters.formatTime(scope.row.next_run_time) }}
        </template>
      </el-table-column>
      <el-table-column
        prop="protocol"
        :formatter="formatProtocol"
        :label="t('task.protocol')"
        width="140"
        class-name="no-wrap-header"
      />
      <el-table-column
        v-if="isAdmin"
        :label="t('common.status')"
      >
        <template #default="scope">
          <el-switch
            v-if="scope.row.level === 1"
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
        v-else
        :label="t('common.status')"
      >
        <template #default="scope">
          <el-switch
            v-if="scope.row.level === 1"
            v-model="scope.row.status"
            :active-value="1"
            :inactive-value="0"
            active-color="#13ce66"
            :disabled="true"
            inactive-color="#ff4949"
          />
        </template>
      </el-table-column>
      <el-table-column
        v-if="isAdmin"
        :label="t('common.operation')"
        :width="locale === 'zh-CN' ? 240 : 280"
      >
        <template #default="scope">
          <div style="display: flex; flex-direction: column; gap: 4px;">
            <div style="display: flex; gap: 4px;">
              <el-button
                type="primary"
                size="small"
                style="flex: 1;"
                @click="toEdit(scope.row)"
              >
                {{ t('common.edit') }}
              </el-button>
              <el-button
                type="success"
                size="small"
                style="flex: 1;"
                @click="runTask(scope.row)"
              >
                {{ t('task.manualRun') }}
              </el-button>
            </div>
            <div style="display: flex; gap: 4px;">
              <el-button
                type="info"
                size="small"
                style="flex: 1;"
                @click="jumpToLog(scope.row)"
              >
                {{ t('task.viewLog') }}
              </el-button>
              <el-button
                type="danger"
                size="small"
                style="flex: 1;"
                @click="remove(scope.row)"
              >
                {{ t('common.delete') }}
              </el-button>
            </div>
          </div>
        </template>
      </el-table-column>
    </el-table>
  </el-main>
</template>

<script>
import { useI18n } from 'vue-i18n'
import taskService from '../../api/task'
import { useUserStore } from '../../stores/user'
import { useNotify } from '@/composables/useNotify'

export default {
  name: 'TaskList',
  setup() {
    const { t, locale } = useI18n()
    return { t, locale }
  },
  data () {
    const userStore = useUserStore()
    return {
      tasks: [],
      hosts: [],
      taskTotal: 0,
      isFirstActivate: true,
      selectedTasks: [],
      searchParams: {
        page_size: 20,
        page: 1,
        id: '',
        protocol: '',
        name: '',
        tag: '',
        selectedTags: [],
        host_id: '',
        status: ''
      },
      tagOptions: [],
      isAdmin: userStore.isAdmin,
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
        {
          value: '2',
          label: this.t('message.activated')
        },
        {
          value: '1',
          label: this.t('message.stopped')
        }
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
  created () {
    const hostId = this.$route.query.host_id
    if (hostId) {
      this.searchParams.host_id = hostId
    }

    this.loadTagOptions()
    this.search()
  },
  activated () {
    if (this.isFirstActivate) {
      this.isFirstActivate = false
      return
    }
    this.search()
  },
  methods: {
    formatLevel (value) {
      return value === 1 ? this.t('task.mainTask') : this.t('task.childTask')
    },
    formatTimeout (value) {
      return value > 0 ? value + this.t('message.seconds') : this.t('message.noLimit')
    },
    formatRetryTimesInterval (value) {
      return value > 0 ? value + this.t('message.seconds') : this.t('message.systemDefault')
    },
    formatMulti (value) {
      return value > 0 ? this.t('common.no') : this.t('common.yes')
    },
    changeStatus (item) {
      if (item.status) {
        taskService.enable(item.id, () => {
          this.search()
        })
      } else {
        taskService.disable(item.id, () => {
          this.search()
        })
      }
    },
    parseCronSpec (spec) {
      if (!spec) return { expr: '', tz: '' }
      const match = spec.match(/^(?:CRON_TZ|TZ)=(\S+)\s+(.+)$/)
      if (match) return { tz: match[1], expr: match[2] }
      return { expr: spec, tz: '' }
    },
    formatProtocol (row, col) {
      if (row[col.property] === 2) {
        return 'shell'
      }
      if (row.http_method === 1) {
        return 'http-get'
      }
      return 'http-post'
    },
    changePage (page) {
      this.searchParams.page = page
      this.search()
    },
    changePageSize (pageSize) {
      this.searchParams.page_size = pageSize
      this.search()
    },
    loadTagOptions () {
      taskService.allTags((tags) => {
        this.tagOptions = tags || []
      })
    },
    search (callback = null) {
      this.searchParams.tag = (this.searchParams.selectedTags || []).join(',')
      taskService.list(this.searchParams, (tasks, hosts) => {
        this.tasks = tasks.data
        this.taskTotal = tasks.total
        this.hosts = hosts
        if (callback) {
          callback()
        }
      })
    },
    async runTask (item) {
      const notify = useNotify()
      if (await notify.confirm(this.t('message.confirmRunTask', { name: item.name }), this.t('message.manualRunTask'))) {
        taskService.run(item.id, () => {
          this.$message.success(this.t('message.taskStarted'))
        })
      }
    },
    async remove (item) {
      const notify = useNotify()
      if (await notify.confirm(this.t('message.confirmDeleteTask', { name: item.name }), this.t('message.confirmDeleteTitle'))) {
        taskService.remove(item.id, () => {
          this.refresh()
        })
      }
    },
    jumpToLog (item) {
      this.$router.push(`/task/log?task_id=${item.id}`)
    },
    refresh () {
      this.search(() => {
        this.$message.success(this.t('message.refreshSuccess'))
      })
    },
    toEdit (item) {
      let path = ''
      if (item === null) {
        path = '/task/create'
      } else {
        path = `/task/edit/${item.id}`
      }
      this.$router.push(path)
    },
    handleSelectionChange (selection) {
      this.selectedTasks = selection.filter(task => task.level === 1)
    },
    async batchEnable () {
      if (this.selectedTasks.length === 0) {
        this.$message.warning(this.t('message.pleaseSelectTask', { action: this.t('task.enable') }))
        return
      }
      const notify = useNotify()
      if (await notify.confirm(this.t('message.confirmBatchEnable', { count: this.selectedTasks.length }), this.t('message.batchEnable'))) {
        const ids = this.selectedTasks.map(task => task.id)
        taskService.batchEnable(ids, () => {
          this.$message.success(this.t('message.batchEnableSuccess'))
          this.selectedTasks = []
          this.search()
        })
      }
    },
    async batchDisable () {
      if (this.selectedTasks.length === 0) {
        this.$message.warning(this.t('message.pleaseSelectTask', { action: this.t('task.disable') }))
        return
      }
      const notify = useNotify()
      if (await notify.confirm(this.t('message.confirmBatchDisable', { count: this.selectedTasks.length }), this.t('message.batchDisable'))) {
        const ids = this.selectedTasks.map(task => task.id)
        taskService.batchDisable(ids, () => {
          this.$message.success(this.t('message.batchDisableSuccess'))
          this.selectedTasks = []
          this.search()
        })
      }
    },
    async batchRemove () {
      if (this.selectedTasks.length === 0) {
        this.$message.warning(this.t('message.pleaseSelectTask', { action: this.t('common.delete') }))
        return
      }
      const notify = useNotify()
      if (await notify.confirm(this.t('message.confirmBatchDelete', { count: this.selectedTasks.length }), this.t('message.batchDelete'))) {
        const ids = this.selectedTasks.map(task => task.id)
        taskService.batchRemove(ids, () => {
          this.$message.success(this.t('message.batchDeleteSuccess'))
          this.selectedTasks = []
          this.search()
        })
      }
    }
  }
}
</script>
<style scoped>
  .demo-table-expand {
    font-size: 0;
  }
  .demo-table-expand label {
    color: #99a9bf;
  }
  .demo-table-expand .el-form-item {
    margin-right: 0;
    margin-bottom: 0;
    width: 50%;
  }

  /* 防止表头文字换行 */
  :deep(.no-wrap-header .cell) {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  /* 表头文字居中对齐 */
  :deep(.el-table th .cell) {
    text-align: center;
  }

  /* 表格内容居中对齐 */
  :deep(.el-table td .cell) {
    text-align: center;
  }
</style>
