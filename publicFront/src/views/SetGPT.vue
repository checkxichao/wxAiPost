<!-- src/views/SetGPT.vue -->
<template>
  <div>
    <el-form :model="gptForm" label-width="120px">
      <el-form-item label="GPT API Key">
        <el-input v-model="gptForm.key" placeholder="请输入 GPT API Key"></el-input>
      </el-form-item>
      <el-form-item label="GPT 模式">
        <el-select v-model="gptForm.model" placeholder="请选择 GPT 模式">
          <el-option label="3.5turbo" value="mode1"></el-option>
          <el-option label="4o" value="mode2"></el-option>
        </el-select>
      </el-form-item>
    </el-form>
    <p>请在使用前确保您的账户有足够的余额。</p>
    <el-button type="primary" @click="saveGPTSettings">保存</el-button>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, onMounted } from 'vue';
import { ElMessage } from 'element-plus';
import http from '../services/http';

export default defineComponent({
  name: 'SetGPT',
  setup() {
    const gptForm = ref<{ key: string; model: string }>({
      key: '',
      model: '',
    });

    const fetchGPTSettings = async () => {
      try {
        const response = await http.get('/GPT/getGptInfo');
        // 检查后端响应数据的结构
        if (response.data && response.data.data) {
          gptForm.value = response.data.data;
        } else {
          gptForm.value = { key: '', model: '' };
          ElMessage.warning('没有 GPT 设置');
        }
      } catch (error) {
        ElMessage.error('获取 GPT 设置失败');
      }
    };



    const saveGPTSettings = async () => {
      if (!gptForm.value.key.trim() || !gptForm.value.model.trim()) {
        ElMessage.warning('请填写完整信息');
        return;
      }

      // 转换 mode1 和 mode2 为实际的 model 值
      const modelMapping: { [key: string]: string } = {
        mode1: 'gpt-3.5-turbo',
        mode2: 'gpt-4o'
      };

      // 使用类型断言确保 model 为合法值
      gptForm.value.model = modelMapping[gptForm.value.model as keyof typeof modelMapping] || gptForm.value.model;

      try {
        await http.post('/GPT/SetGptInfo', gptForm.value);
        ElMessage.success('保存成功');
      } catch (error) {
        ElMessage.error('保存失败');
      }
    };




    onMounted(() => {
      fetchGPTSettings();
    });

    return {
      gptForm,
      saveGPTSettings,
    };
  },
});
</script>

<style scoped>
p {
  margin-top: 20px;
  color: #f56c6c;
}
.el-button {
  margin-top: 20px;
}
</style>
