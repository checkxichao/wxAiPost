<!-- src/components/Login.vue -->
<template>
  <el-row type="flex" justify="center" align="middle" class="login-row">
    <el-col :span="8">
      <el-card class="login-card">
        <h2 style="text-align: center;">登录</h2>
        <el-form :model="loginForm" :rules="rules" ref="loginFormRef" label-width="80px">
          <el-form-item label="用户名" prop="username">
            <el-input v-model="loginForm.username" placeholder="请输入用户名" autocomplete="off" />
          </el-form-item>

          <el-form-item label="密码" prop="password">
            <el-input v-model="loginForm.password" type="password" placeholder="请输入密码" autocomplete="off" />
          </el-form-item>

          <el-form-item>
            <el-button type="primary" @click="onSubmit">登录</el-button>
            <el-button @click="resetForm">重置</el-button>
          </el-form-item>
        </el-form>
        <el-alert v-if="error" :title="error" type="error" show-icon />
      </el-card>
    </el-col>
  </el-row>
</template>

<script lang="ts">
import { defineComponent, ref } from 'vue';
import { ElForm, ElMessage } from 'element-plus';
import http from '../services/http';
import { useRouter } from 'vue-router';
import { useUserStore } from '@/store/user'; // 确保路径正确

export default defineComponent({
  name: 'UserLogin',
  setup() {
    const loginFormRef = ref<InstanceType<typeof ElForm> | null>(null);
    const loginForm = ref({
      username: '',
      password: '',
    });
    const error = ref<string | null>(null);
    const router = useRouter();
    const userStore = useUserStore();

    const rules = {
      username: [
        { required: true, message: '请输入用户名', trigger: 'blur' },
      ],
      password: [
        { required: true, message: '请输入密码', trigger: 'blur' },
      ],
    };

    const onSubmit = async () => {
      if (loginFormRef.value) {
        const valid = await loginFormRef.value.validate();
        if (valid) {
          try {
            const response = await http.post('/auth/login', {
              username: loginForm.value.username,
              password: loginForm.value.password,
            });

            console.log('Response:', response.data);  // Check the response structure

            if (response.data.access_token) {
              const token = response.data.access_token;
              const username = response.data.user?.username || loginForm.value.username;

              // Assuming 'role' is returned in the response or use a default
              const role = response.data.user?.role || 'user';  // Default to 'user' if not provided

              // Set the user data including token, username, and role
              userStore.setUser({ token, username, role });

              ElMessage.success('登录成功');
              await router.push('/dashboard'); // Redirect to the dashboard
            } else {
              throw new Error('登录失败，未获取到 Token');
            }
          } catch (err: any) {
            console.error('Login failed:', err);
            error.value = err.response?.data?.error || '登录失败';
          }
        } else {
          console.log('表单验证失败');
        }
      }
    };


    const resetForm = () => {
      if (loginFormRef.value) {
        loginFormRef.value.resetFields();
        error.value = null;
      }
    };

    return {
      loginForm,
      error,
      onSubmit,
      resetForm,
      rules,
      loginFormRef,
    };
  },
});
</script>

<style scoped>
.login-row {
  height: 100vh;
}
.login-card {
  padding: 20px;
}
</style>
