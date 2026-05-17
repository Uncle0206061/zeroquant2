<template>
  <div class="login-container">
    <el-card class="login-card">
      <h2 class="login-title">ZeroQuant 2.0</h2>
      <p class="login-subtitle">量化交易系统</p>
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="0"
        @submit.prevent="handleLogin"
      >
        <el-form-item prop="username">
          <el-input
            v-model="form.username"
            placeholder="用户名 / 手机号"
            :prefix-icon="User"
            size="large"
          />
        </el-form-item>
        <el-form-item prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="密码"
            :prefix-icon="Lock"
            size="large"
            show-password
          />
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            native-type="submit"
            :loading="loading"
            size="large"
            style="width: 100%"
          >
            登录
          </el-button>
        </el-form-item>
        <div class="form-footer">
          <el-checkbox v-model="rememberMe">记住账号</el-checkbox>
          <router-link to="/register">注册账号</router-link>
        </div>
      </el-form>
      <div class="footer-link">
        <!-- 备用链接位 -->
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'

const router = useRouter()
const auth = useAuthStore()
const formRef = ref<FormInstance>()
const loading = ref(false)
const rememberMe = ref(localStorage.getItem('zq_saved_username') !== null)

const form = reactive({
  username: localStorage.getItem('zq_saved_username') || '',
  password: '',
})

const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

async function handleLogin() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    await auth.login(form.username, form.password)
    if (rememberMe.value) {
      localStorage.setItem('zq_saved_username', form.username)
    } else {
      localStorage.removeItem('zq_saved_username')
    }
    ElMessage.success('登录成功')
    router.push('/')
  } catch (e: any) {
    ElMessage.error(e.message || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #001529 0%, #003a70 100%);
}
.login-card {
  width: 400px;
  border-radius: 8px;
}
.login-title {
  text-align: center;
  margin: 0 0 4px;
  font-size: 24px;
  color: #001529;
}
.login-subtitle {
  text-align: center;
  margin: 0 0 24px;
  color: #999;
  font-size: 14px;
}
.remember-row {
  display: flex;
  justify-content: flex-start;
  margin-top: -12px;
}
.form-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 8px;
}
.form-footer a {
  color: #409eff;
  text-decoration: none;
  font-size: 14px;
}
.footer-link {
  text-align: center;
  min-height: 20px;
}
</style>
