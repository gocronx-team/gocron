<template>
  <el-main>
    <el-form ref="form" :model="form" :rules="formRules" label-width="auto">
      <el-row>
        <el-col :span="12">
          <el-form-item :label="t('template.name')" prop="name">
            <el-input v-model.trim="form.name" :placeholder="t('template.templateNamePlaceholder')"></el-input>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item :label="t('template.category')" prop="category">
            <el-select
              v-model="form.category"
              filterable
              allow-create
              default-first-option
              :placeholder="t('template.selectCategory')"
              style="width: 100%">
              <el-option value="backup" :label="t('template.category_backup')"></el-option>
              <el-option value="cleanup" :label="t('template.category_cleanup')"></el-option>
              <el-option value="monitor" :label="t('template.category_monitor')"></el-option>
              <el-option value="deploy" :label="t('template.category_deploy')"></el-option>
              <el-option value="api" :label="t('template.category_api')"></el-option>
              <el-option value="custom" :label="t('template.category_custom')"></el-option>
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>
      <el-row>
        <el-col :span="24">
          <el-form-item :label="t('template.description')">
            <el-input v-model="form.description" :placeholder="t('template.templateDescPlaceholder')"></el-input>
          </el-form-item>
        </el-col>
      </el-row>
      <el-row>
        <el-col :span="8">
          <el-form-item :label="t('template.protocol')">
            <el-select v-model.trim="form.protocol">
              <el-option :value="1" label="HTTP"></el-option>
              <el-option :value="2" label="Shell"></el-option>
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="8" v-if="form.protocol === 1">
          <el-form-item :label="t('task.httpMethod')">
            <el-select v-model.trim="form.http_method">
              <el-option :value="1" label="GET"></el-option>
              <el-option :value="2" label="POST"></el-option>
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item :label="t('template.timeout')">
            <el-input-number v-model="form.timeout" :min="0" :max="86400"></el-input-number>
          </el-form-item>
        </el-col>
      </el-row>
      <el-row>
        <el-col :span="20">
          <el-form-item :label="t('template.command')" prop="command">
            <div style="width: 100%;">
              <MonacoEditor
                v-model="form.command"
                :language="editorLanguage"
                height="250px"
              />
              <div style="color: #909399; font-size: 12px; margin-top: 4px;">
                {{ t('template.templateVarTip') }}
              </div>
            </div>
          </el-form-item>
        </el-col>
      </el-row>
      <el-row v-if="form.protocol === 1 && form.http_method === 2">
        <el-col :span="16">
          <el-form-item :label="t('task.httpBody')">
            <el-input type="textarea" :rows="4" v-model="form.http_body"></el-input>
          </el-form-item>
        </el-col>
      </el-row>
      <el-row v-if="form.protocol === 1">
        <el-col :span="16">
          <el-form-item :label="t('task.httpHeaders')">
            <el-input type="textarea" :rows="3" v-model="form.http_headers"></el-input>
          </el-form-item>
        </el-col>
      </el-row>
      <el-form-item>
        <el-button type="primary" @click="submit">{{ t('common.save') }}</el-button>
        <el-button @click="cancel">{{ t('common.cancel') }}</el-button>
      </el-form-item>
    </el-form>
  </el-main>
</template>

<script>
import { useI18n } from 'vue-i18n'
import templateService from '../../api/template'
import MonacoEditor from '../../components/common/MonacoEditor.vue'

export default {
  name: 'template-edit',
  components: { MonacoEditor },
  setup() {
    const { t } = useI18n()
    return { t }
  },
  computed: {
    editorLanguage() {
      return this.form.protocol === 1 ? 'plaintext' : 'shell'
    }
  },
  data() {
    return {
      form: {
        id: '',
        name: '',
        description: '',
        category: 'custom',
        protocol: 2,
        command: '',
        http_method: 1,
        http_body: '',
        http_headers: '',
        timeout: 300
      },
      formRules: {
        name: [
          { required: true, message: '', trigger: 'blur' }
        ],
        category: [
          { required: true, message: '', trigger: 'blur' }
        ],
        command: [
          { required: true, message: '', trigger: 'blur' }
        ]
      }
    }
  },
  created() {
    this.formRules.name[0].message = this.t('template.templateNamePlaceholder')
    this.formRules.category[0].message = this.t('template.selectCategory')
    this.formRules.command[0].message = this.t('message.pleaseEnterCommand')

    const id = this.$route.params.id
    if (id) {
      templateService.detail(id, (data) => {
        if (data) {
          this.form = {
            id: data.id,
            name: data.name,
            description: data.description || '',
            category: data.category,
            protocol: data.protocol,
            command: data.command,
            http_method: data.http_method || 1,
            http_body: data.http_body || '',
            http_headers: data.http_headers || '',
            timeout: data.timeout || 300
          }
        }
      })
    }
  },
  methods: {
    submit() {
      this.$refs.form.validate().then((valid) => {
        if (!valid) return false
        templateService.store(this.form, () => {
          this.$router.push('/template')
        })
      }).catch(() => {})
    },
    cancel() {
      this.$router.push('/template')
    }
  }
}
</script>
