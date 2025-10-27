<template>
  <el-dialog
    title="用户登录"
    v-model="dialogVisible"
    :close-on-click-modal="false"
    :show-close="false"
    :close-on-press-escape="false"
    width="40%">
    <el-form ref="formRef" :model="form" label-width="80px" :rules="formRules">
      <el-form-item label="用户名" prop="username">
        <el-col :span="16">
          <el-input v-model.trim="form.username" placeholder="请输入用户名或邮箱" />
        </el-col>
      </el-form-item>
      <el-form-item label="密码" prop="password">
        <el-col :span="16">
          <el-input v-model.trim="form.password" type="password" placeholder="请输入密码" @keyup.enter="submit" />
        </el-col>
      </el-form-item>
      <el-form-item label="验证码" v-if="require2FA">
        <el-col :span="16">
          <el-input v-model.trim="form.twoFactorCode" placeholder="请输入6位验证码" maxlength="6" @keyup.enter="submit" />
        </el-col>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="submit" :loading="loading">登录</el-button>
      </el-form-item>
    </el-form>
  </el-dialog>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '../../stores/user'
import { useLoading } from '../../composables/useLoading'
import userService from '../../api/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const { loading, withLoading } = useLoading()

const dialogVisible = ref(true)
const require2FA = ref(false)
const formRef = ref()

const form = reactive({
  username: '',
  password: '',
  twoFactorCode: ''
})

const formRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

const submit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    
    await withLoading(async () => {
      const params = {
        username: form.username,
        password: form.password,
        two_factor_code: form.twoFactorCode || undefined
      }
      
      userService.login(params.username, params.password, params.two_factor_code, (data) => {
        if (data.require_2fa) {
          require2FA.value = true
          return
        }
        
        userStore.setUser({
          token: data.token,
          uid: data.uid,
          username: data.username,
          isAdmin: data.is_admin
        })
        
        router.push(route.query.redirect || '/')
      })
    })
  })
}
</script>
