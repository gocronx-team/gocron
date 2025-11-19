<template>
  <el-container>
    <el-main>
      <el-form ref="form" :model="form" :rules="formRules" :label-width="locale === 'zh-CN' ? '100px' : '150px'" style="width: 500px;">
        <el-form-item>
          <el-input v-model="form.id" type="hidden"></el-input>
        </el-form-item>
        <el-form-item :label="t('host.alias')" prop="alias">
          <el-input v-model="form.alias"></el-input>
        </el-form-item>
        <el-form-item :label="t('host.name')" prop="name">
          <el-input v-model="form.name"></el-input>
        </el-form-item>
        <el-form-item :label="t('host.port')" prop="port">
          <el-input v-model.number="form.port"></el-input>
        </el-form-item>
        <el-form-item :label="t('host.remark')">
          <el-input
            type="textarea"
            :rows="5"
            v-model="form.remark">
          </el-input>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="submit()">{{ t('common.save') }}</el-button>
          <el-button @click="cancel">{{ t('common.cancel') }}</el-button>
        </el-form-item>
      </el-form>
    </el-main>
  </el-container>
</template>

<script>
import { useI18n } from 'vue-i18n'
import hostService from '../../api/host'
export default {
  name: 'edit',
  setup() {
    const { t, locale } = useI18n()
    return { t, locale }
  },
  data: function () {
    return {
      form: {
        id: '',
        name: '',
        port: 5921,
        alias: '',
        remark: ''
      },
      formRules: {}
    }
  },
  computed: {
    computedFormRules() {
      return {
        name: [
          {required: true, message: this.t('host.nameRequired'), trigger: 'blur'}
        ],
        port: [
          {required: true, message: this.t('host.portRequired'), trigger: 'blur'},
          {type: 'number', message: this.t('host.portInvalid')}
        ],
        alias: [
          {required: true, message: this.t('host.aliasRequired'), trigger: 'blur'}
        ]
      }
    }
  },
  watch: {
    computedFormRules: {
      handler(newVal) {
        this.formRules = newVal
      },
      immediate: true
    }
  },
  created () {
    const id = this.$route.params.id
    if (!id) {
      this.resetForm()
      return
    }
    hostService.detail(id, (data) => {
      if (!data) {
        this.$message.error(this.t('message.dataNotFound'))
        this.cancel()
        return
      }
      this.form.id = data.id
      this.form.name = data.name
      this.form.port = data.port
      this.form.alias = data.alias
      this.form.remark = data.remark
    })
  },
  methods: {
    resetForm() {
      this.form = {
        id: '',
        name: '',
        port: 5921,
        alias: '',
        remark: ''
      }
      if (this.$refs.form) {
        this.$refs.form.clearValidate()
      }
    },
    submit () {
      this.$refs['form'].validate((valid) => {
        if (!valid) {
          return false
        }
        this.save()
      })
    },
    save () {
      hostService.update(this.form, () => {
        this.$router.push('/host')
      })
    },
    cancel () {
      this.$router.push('/host')
    }
  }
}
</script>
