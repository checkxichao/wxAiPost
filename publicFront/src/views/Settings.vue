<!-- src/views/Settings.vue -->
<template>
  <div>
    <h2>设置</h2>
    <el-card>
      <el-form :model="settingsForm" :rules="rules" ref="settingsFormRef" label-width="120px">
        <el-form-item label="当前密码" prop="currentPassword">
          <el-input v-model="settingsForm.currentPassword" type="password" placeholder="请输入当前密码" autocomplete="off" />
        </el-form-item>

        <el-form-item label="新密码" prop="newPassword">
          <el-input v-model="settingsForm.newPassword" type="password" placeholder="请输入新密码" autocomplete="off" />
        </el-form-item>

        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input v-model="settingsForm.confirmPassword" type="password" placeholder="请确认新密码" autocomplete="off" />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="updatePassword">更新密码</el-button>
          <el-button @click="resetForm">重置</el-button>
        </el-form-item>
      </el-form>
      <el-alert v-if="success" title="密码更新成功" type="success" show-icon />
      <el-alert v-if="error" :title="error" type="error" show-icon />
    </el-card>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref } from 'vue';
import { ElForm, ElMessage } from 'element-plus';
import http from '../services/http';

export default defineComponent({
  name: 'SettingsView', // 改为多字名称
  setup() {
    const settingsFormRef = ref<InstanceType<typeof ElForm> | null>(null);
    const settingsForm = ref({
      currentPassword: '',
      newPassword: '',
      confirmPassword: '',
    });
    const success = ref<boolean>(false);
    const error = ref<string | null>(null);

    const rules = {
      currentPassword: [
        { required: true, message: '请输入当前密码', trigger: 'blur' },
      ],
      newPassword: [
        { required: true, message: '请输入新密码', trigger: 'blur' },
      ],
      confirmPassword: [
        { required: true, message: '请确认新密码', trigger: 'blur' },
        {
          validator: (_rule: any, value: string, callback: any) => {
            if (value !== settingsForm.value.newPassword) {
              callback(new Error('两次输入密码不一致'));
            } else {
              callback();
            }
          },
          trigger: 'blur',
        },
      ],
    };

    const updatePassword = async () => {
      if (settingsFormRef.value) {
        const valid = await settingsFormRef.value.validate();
        if (valid) {
          try {
            const response = await http.post('/update-password', {
              currentPassword: settingsForm.value.currentPassword,
              newPassword: settingsForm.value.newPassword,
            });
            ElMessage.success('密码更新成功');
            success.value = true;
            error.value = null;
            // 重置表单
            settingsFormRef.value.resetFields();
          } catch (err: any) {
            error.value = err.response?.data?.error || '密码更新失败';
            success.value = false;
          }
        } else {
          console.log('表单验证失败');
        }
      }
    };

    const resetForm = () => {
      if (settingsFormRef.value) {
        settingsFormRef.value.resetFields();
        success.value = false;
        error.value = null;
      }
    };

    return {
      settingsForm,
      success,
      error,
      updatePassword,
      resetForm,
      rules,
      settingsFormRef,
    };
  },
});
</script>

<style scoped>
h2 {
  margin-bottom: 20px;
}
</style>
