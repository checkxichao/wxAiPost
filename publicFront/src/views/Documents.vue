<!-- src/views/Documents.vue -->
<template>
  <div>
    <h2>操作手册</h2>
    <el-card>
      <p>要发文的文章设置好服务器IP白名单：</p>
      <p>拿到appid和secrets填入保存区</p>
      <p>设置好模板,勾选要不要带作者,带作者名字不要大于6个字</p>
      <p>设置随机时间间隔</p>
      <p>设置文章是否需要随机文字颜色</p>
      <p>设置gpt密钥,勾选版本</p>
      <!-- 你可以在这里添加更多的 Element Plus 组件来丰富页面内容 -->
    </el-card>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'DocumentsPage', // 确保组件名称为多字
  setup() {
    // 这里可以添加组合式 API 的逻辑
    return {};
  },
});
</script>

<style scoped>
h2 {
  margin-bottom: 20px;
}
</style>
