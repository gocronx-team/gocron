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
        <el-col :span="3">
          <el-button type="success" v-if="isAdmin" @click="showRegisterDialog">{{ t('host.agentRegister') }}</el-button>
        </el-col>
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

      <!-- Agent 注册对话框 -->
      <el-dialog
        v-model="registerDialogVisible"
        :title="t('host.agentRegister')"
        width="700px">
        <el-form label-width="120px">
          <el-form-item :label="t('host.registerToken')">
            <el-input v-model="registerToken" readonly>
              <template #append>
                <el-button @click="generateToken" :loading="generating">{{ t('host.generateToken') }}</el-button>
              </template>
            </el-input>
          </el-form-item>
          <el-form-item :label="t('host.serverUrl')">
            <el-input v-model="serverUrl" placeholder="http://your-server:5920"></el-input>
          </el-form-item>
          <el-form-item :label="t('host.oneClickInstall')">
            <el-tabs>
              <el-tab-pane :label="t('host.linuxMac')">
                <el-input
                  v-model="installCommand"
                  type="textarea"
                  :rows="2"
                  readonly>
                </el-input>
                <el-button type="primary" @click="copyInstallCommand" style="margin-top: 10px">
                  {{ t('common.copy') }}
                </el-button>
                <div style="margin-top: 10px; color: #909399; font-size: 12px;">
                  {{ t('host.installTip') }}
                </div>
              </el-tab-pane>
              <el-tab-pane :label="t('host.windows')">
                <el-input
                  v-model="installCommandWindows"
                  type="textarea"
                  :rows="2"
                  readonly>
                </el-input>
                <el-button type="primary" @click="copyInstallCommandWindows" style="margin-top: 10px">
                  {{ t('common.copy') }}
                </el-button>
                <div style="margin-top: 10px; color: #909399; font-size: 12px;">
                  {{ t('host.installTipWindows') }}
                </div>
              </el-tab-pane>
              <el-tab-pane :label="t('host.manualInstall')">
                <el-input
                  v-model="registerCommand"
                  type="textarea"
                  :rows="4"
                  readonly>
                </el-input>
                <el-button type="primary" @click="copyCommand" style="margin-top: 10px">
                  {{ t('common.copy') }}
                </el-button>
              </el-tab-pane>
              <el-tab-pane :label="t('host.manualDownload')">
                <div style="padding: 10px 0;">
                  <el-row :gutter="10" style="margin-bottom: 15px;">
                    <el-col :span="16">
                      <el-select v-model="selectedPlatform" style="width: 100%" size="large">
                        <el-option label="Linux AMD64" value="linux-amd64"></el-option>
                        <el-option label="Linux ARM64" value="linux-arm64"></el-option>
                        <el-option label="macOS AMD64 (Intel)" value="darwin-amd64"></el-option>
                        <el-option label="macOS ARM64 (Apple Silicon)" value="darwin-arm64"></el-option>
                        <el-option label="Windows AMD64" value="windows-amd64"></el-option>
                        <el-option label="Windows ARM64" value="windows-arm64"></el-option>
                      </el-select>
                    </el-col>
                    <el-col :span="8">
                      <el-button type="primary" @click="downloadAgent" :disabled="!registerToken" size="large" style="width: 100%">
                        {{ t('host.downloadAgent') }}
                      </el-button>
                    </el-col>
                  </el-row>
                  <el-alert type="info" :closable="false" show-icon>
                    <template #title>
                      <span style="font-size: 13px;">{{ t('host.installTip') }}</span>
                    </template>
                  </el-alert>
                </div>
              </el-tab-pane>
            </el-tabs>
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="registerDialogVisible = false">{{ t('common.close') }}</el-button>
        </template>
      </el-dialog>
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
    // 开发环境使用后端端口，生产环境使用当前地址
    const serverUrl = window.location.port === '8080' ? 'http://localhost:5920' : window.location.origin
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
      registerDialogVisible: false,
      registerToken: '',
      serverUrl: serverUrl,
      generating: false,
      selectedPlatform: 'linux-amd64'
    }
  },
  computed: {
    registerCommand() {
      if (!this.registerToken) {
        return this.t('host.generateTokenFirst')
      }
      return `./gocron-node \\
  -s 0.0.0.0:5921 \\
  -enable-register \\
  -gocron-url ${this.serverUrl} \\
  -register-token ${this.registerToken}`
    },
    installCommand() {
      if (!this.registerToken) {
        return this.t('host.generateTokenFirst')
      }
      return `curl -fsSL "${this.serverUrl}/install/agent?token=${this.registerToken}" | bash`
    },
    installCommandWindows() {
      if (!this.registerToken) {
        return this.t('host.generateTokenFirst')
      }
      return `powershell -Command "Invoke-WebRequest -Uri '${this.serverUrl}/install/agent?token=${this.registerToken}&os=windows' -OutFile install.bat; .\\install.bat"`
    }
  },
  created () {
    this.search()
  },
  watch: {
    '$route'(to, from) {
      if (to.path === '/host' && (from.path === '/host/create' || from.path.startsWith('/host/edit/'))) {
        this.search()
      }
    }
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
    showRegisterDialog() {
      this.registerDialogVisible = true
      this.loadRegisterToken()
    },
    loadRegisterToken() {
      hostService.getRegisterToken((data) => {
        this.registerToken = data.token || ''
      })
    },
    generateToken() {
      this.generating = true
      hostService.generateRegisterToken((data) => {
        this.registerToken = data.token
        this.$message.success(this.t('message.generateSuccess'))
        this.generating = false
      })
    },
    copyCommand() {
      if (!this.registerToken) {
        this.$message.warning(this.t('host.generateTokenFirst'))
        return
      }
      navigator.clipboard.writeText(this.registerCommand).then(() => {
        this.$message.success(this.t('message.copySuccess'))
      }).catch(() => {
        this.$message.error(this.t('message.copyFailed'))
      })
    },
    copyInstallCommand() {
      if (!this.registerToken) {
        this.$message.warning(this.t('host.generateTokenFirst'))
        return
      }
      navigator.clipboard.writeText(this.installCommand).then(() => {
        this.$message.success(this.t('message.copySuccess'))
      }).catch(() => {
        this.$message.error(this.t('message.copyFailed'))
      })
    },
    copyInstallCommandWindows() {
      if (!this.registerToken) {
        this.$message.warning(this.t('host.generateTokenFirst'))
        return
      }
      navigator.clipboard.writeText(this.installCommandWindows).then(() => {
        this.$message.success(this.t('message.copySuccess'))
      }).catch(() => {
        this.$message.error(this.t('message.copyFailed'))
      })
    },
    downloadAgent() {
      if (!this.registerToken) {
        this.$message.warning(this.t('host.generateTokenFirst'))
        return
      }
      const [os, arch] = this.selectedPlatform.split('-')
      const url = `${this.serverUrl}/download/agent?token=${this.registerToken}&os=${os}&arch=${arch}`
      window.open(url, '_blank')
    }
  }
}
</script>
