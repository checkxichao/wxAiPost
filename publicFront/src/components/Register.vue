<!-- src/components/Register.vue -->
<template>
  <el-row type="flex" justify="center" align="middle" class="register-row">
    <el-col :span="8">
      <el-card class="register-card">
        <h2 style="text-align: center;">注册</h2>
        <el-form :model="registerForm" :rules="rules" ref="registerFormRef" label-width="80px">
          <el-form-item label="用户名" prop="username">
            <el-input v-model="registerForm.username" placeholder="请输入用户名" autocomplete="off" />
          </el-form-item>

          <el-form-item label="密码" prop="password">
            <el-input v-model="registerForm.password" type="password" placeholder="请输入密码" autocomplete="off" />
          </el-form-item>

          <el-form-item>
            <el-button type="primary" @click="onSubmit">注册</el-button>
            <el-button @click="resetForm">重置</el-button>
          </el-form-item>
        </el-form>
        <el-alert v-if="error" :title="error" type="error" show-icon />
        <el-alert v-if="success" title="用户注册成功" type="success" show-icon />
      </el-card>
    </el-col>
  </el-row>
</template>

<script lang="ts">
import { defineComponent, ref } from 'vue';
import { ElForm, ElMessage } from 'element-plus';
import http from '../services/http';
import { useRouter } from 'vue-router';
import { useUserStore } from '@/store/user';

export default defineComponent({
  name: 'UserRegister', // 改为多字名称，如 'UserRegisterView'
  setup() {
    const registerFormRef = ref<InstanceType<typeof ElForm> | null>(null);
    const registerForm = ref({
      username: '',
      password: '',
    });
    const success = ref<boolean>(false);
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
      if (registerFormRef.value) {
        const valid = await registerFormRef.value.validate();
        if (valid) {
          try {
            const response = await http.post('/auth/register', {
              username: registerForm.value.username,
              password: registerForm.value.password,
            });
            ElMessage.success('用户注册成功');
            success.value = true;
            error.value = null;
            // 保存用户名到 Store（假设后端返回用户名）
            userStore.setUser(response.data.username);
            // 跳转到登录页面
            router.push('/login');
          } catch (err: any) {
            error.value = err.response?.data?.error || '注册失败';
            success.value = false;
          }
        } else {
          console.log('表单验证失败');
        }
      }
    };

    const resetForm = () => {
      if (registerFormRef.value) {
        registerFormRef.value.resetFields();
        success.value = false;
        error.value = null;
      }
    };

    return {
      registerForm,
      success,
      error,
      onSubmit,
      resetForm,
      rules,
      registerFormRef,
    };
  },
});
</script>

<style scoped>
.register-row {
  height: 100vh;
}
.register-card {
  padding: 20px;
}
</style>
