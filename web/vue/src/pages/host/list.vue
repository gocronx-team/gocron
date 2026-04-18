<template>
  <el-main>
    <el-form :inline="true">
      <el-row>
        <el-form-item label="ID">
          <el-input v-model.trim="searchParams.id" />
        </el-form-item>
        <el-form-item :label="t('host.name')">
          <el-input v-model.trim="searchParams.name" />
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            @click="search()"
          >
            {{ t('common.search') }}
          </el-button>
        </el-form-item>
      </el-row>
    </el-form>
    <el-row
      type="flex"
      justify="end"
      style="gap: 10px; margin-bottom: 15px;"
    >
      <el-button
        v-if="isAdmin"
        type="success"
        icon="Download"
        @click="showAgentInstall"
      >
        {{ t('host.autoRegister') }}
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
        icon="Refresh"
        @click="refresh"
      >
        {{ t('common.refresh') }}
      </el-button>
    </el-row>
    <el-pagination
      v-model:current-page="searchParams.page"
      v-model:page-size="searchParams.page_size"
      background
      layout="prev, pager, next, sizes, total"
      :total="hostTotal"
      @size-change="changePageSize"
      @current-change="changePage"
    />
    <el-table
      :data="hosts"
      tooltip-effect="dark"
      border
      style="width: 100%"
    >
      <el-table-column
        prop="id"
        label="ID"
      />
      <el-table-column
        prop="alias"
        :label="t('host.alias')"
      />
      <el-table-column
        prop="name"
        :label="t('host.name')"
      />
      <el-table-column
        prop="port"
        :label="t('host.port')"
      />
      <el-table-column :label="t('task.viewLog')">
        <template #default="scope">
          <el-button
            type="success"
            @click="toTasks(scope.row)"
          >
            {{ t('task.list') }}
          </el-button>
        </template>
      </el-table-column>
      <el-table-column
        prop="remark"
        :label="t('host.remark')"
      />
      <el-table-column
        v-if="isAdmin"
        :label="t('common.operation')"
        :width="locale === 'zh-CN' ? 260 : 300"
      >
        <template #default="scope">
          <el-button
            type="primary"
            size="small"
            @click="toEdit(scope.row)"
          >
            {{ t('common.edit') }}
          </el-button>
          <el-button
            type="info"
            size="small"
            @click="ping(scope.row)"
          >
            {{ t('system.testSend') }}
          </el-button>
          <el-button
            type="danger"
            size="small"
            @click="remove(scope.row)"
          >
            {{ t('common.delete') }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog
      v-model="agentDialogVisible"
      :title="t('host.agentInstall')"
      width="750px"
    >
      <div v-if="installCommand">
        <el-alert
          :title="t('host.installTip')"
          type="info"
          :closable="false"
          style="margin-bottom: 20px"
          show-icon
        />
        
        <el-tabs
          v-model="activeTab"
          type="card"
        >
          <el-tab-pane
            label="Linux / macOS"
            name="linux"
          >
            <div style="padding: 15px; background: #f5f7fa; border-radius: 4px;">
              <div style="margin-bottom: 10px; color: #606266; font-size: 14px;">
                <el-icon style="vertical-align: middle;">
                  <Monitor />
                </el-icon>
                {{ t('host.bashCommand') }}
              </div>
              <el-input
                v-model="installCommand"
                type="textarea"
                :rows="3"
                readonly
                style="font-family: monospace; font-size: 13px;"
              />
              <div style="margin-top: 10px; text-align: right;">
                <el-button
                  type="primary"
                  icon="DocumentCopy"
                  @click="copyCommand('linux')"
                >
                  Copy
                </el-button>
              </div>
            </div>
          </el-tab-pane>
          
          <el-tab-pane
            label="Windows"
            name="windows"
          >
            <div style="padding: 15px;">
              <el-alert
                type="warning"
                :closable="false"
                style="margin-bottom: 15px;"
              >
                <template #title>
                  <strong>{{ t('host.windowsManualInstall') }}</strong>
                </template>
                {{ t('host.windowsManualInstallTip') }}
              </el-alert>
              
              <el-steps
                direction="vertical"
                :active="3"
              >
                <el-step
                  :title="t('host.windowsStep1')"
                  :description="t('host.windowsStep1Desc')"
                />
                <el-step
                  :title="t('host.windowsStep2')"
                  :description="t('host.windowsStep2Desc')"
                />
                <el-step
                  :title="t('host.windowsStep3')"
                  :description="t('host.windowsStep3Desc')"
                />
              </el-steps>
            </div>
          </el-tab-pane>
        </el-tabs>
        
        <el-divider />
        
        <div style="padding: 10px 0;">
          <el-descriptions
            :column="1"
            border
          >
            <el-descriptions-item :label="t('host.tokenExpires')">
              <el-tag
                type="warning"
                effect="plain"
              >
                {{ expiresAt }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item :label="t('host.tokenUsage')">
              <span style="color: #67c23a;">{{ t('host.tokenReusable') }}</span>
            </el-descriptions-item>
          </el-descriptions>
        </div>
      </div>
      <div
        v-else
        style="text-align: center; padding: 20px"
      >
        <el-icon
          class="is-loading"
          :size="30"
        >
          <Loading />
        </el-icon>
        <p>{{ t('common.loading') }}</p>
      </div>
    </el-dialog>
  </el-main>
</template>

<script>
import { useI18n } from 'vue-i18n'
import { useNotify } from '@/composables/useNotify'
import { Loading } from '@element-plus/icons-vue'
import hostService from '../../api/host'
import agentService from '../../api/agent'
import { useUserStore } from '../../stores/user'

export default {
  name: 'HostList',
  components: {
    Loading
  },
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
      isAdmin: userStore.isAdmin,
      agentDialogVisible: false,
      installCommand: '',
      expiresAt: '',
      activeTab: 'linux',
      cachedToken: null,
      cachedTokenExpires: null
    }
  },
  watch: {
    '$route'(to, from) {
      if (to.path === '/host' && (from.path === '/host/create' || from.path.startsWith('/host/edit/'))) {
        this.search()
      }
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
    async remove (item) {
      const notify = useNotify()
      if (await notify.confirm(this.t('message.confirmDeleteNode'), this.t('common.tip'))) {
        hostService.remove(item.id, () => this.refresh())
      }
    },
    ping (item) {
      if (!item.id || item.id <= 0) {
        this.$message.error(this.t('message.dataNotFound'))
        return
      }
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
    },
    showAgentInstall () {
      this.agentDialogVisible = true
      
      // 检查是否有缓存的token且未过期
      const now = new Date()
      if (this.cachedToken && this.cachedTokenExpires && now < this.cachedTokenExpires) {
        // 使用缓存的token
        this.installCommand = this.cachedToken.install_cmd
        this.expiresAt = this.cachedTokenExpires.toLocaleString()
        return
      }
      
      // 生成新token
      this.installCommand = ''
      this.expiresAt = ''
      agentService.generateToken((data) => {
        this.installCommand = data.install_cmd
        const expiresDate = new Date(data.expires_at)
        this.expiresAt = expiresDate.toLocaleString()
        
        // 缓存token信息
        this.cachedToken = data
        this.cachedTokenExpires = expiresDate
      })
    },
    copyCommand (type) {
      const cmd = type === 'windows' ? this.installCommandWindows : this.installCommand
      navigator.clipboard.writeText(cmd).then(() => {
        this.$message.success(this.t('message.copySuccess'))
      }).catch(() => {
        this.$message.error(this.t('message.copyFailed'))
      })
    }
  }
}
</script>
