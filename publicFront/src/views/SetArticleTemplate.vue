<template>
  <div>
    <!-- 选择模板的下拉框 -->
    <el-form :model="templateForm" label-width="120px">
      <el-form-item label="选择模板">
        <el-select v-model="templateForm.selectedTemplate" placeholder="请选择模板" @change="fetchTemplate">
          <el-option
              v-for="template in templates"
              :key="template"
              :label="template"
              :value="template">
          </el-option>
        </el-select>
      </el-form-item>

      <!-- 输入模板内容 -->
      <el-form-item label="文章模板">
        <el-input
            type="textarea"
            v-model="templateForm.content"
            :rows="20"
            placeholder="请输入文章模板"
        ></el-input>
      </el-form-item>
    </el-form>

    <!-- 保存按钮 -->
    <el-button type="primary" @click="saveTemplate">保存</el-button>

    <!-- 预览按钮 -->
    <el-button type="info" @click="previewTemplate">预览</el-button>

    <!-- 新增模板输入框 -->
    <el-form :model="newTemplate" label-width="120px" style="margin-top: 20px;">
      <el-form-item label="新增模板名称">
        <el-input v-model="newTemplate.templateId" placeholder="请输入新增模板名称"></el-input>
      </el-form-item>
      <el-form-item label="新增模板内容">
        <el-input
            type="textarea"
            v-model="newTemplate.content"
            :rows="10"
            placeholder="请输入新增模板内容"
        ></el-input>
      </el-form-item>
      <el-button type="success" @click="addTemplate">新增模板</el-button>
    </el-form>
  </div>
</template>

<script lang="ts">
import {defineComponent, ref, onMounted} from 'vue';
import {ElMessage} from 'element-plus';
import http from '../services/http';

export default defineComponent({
  name: 'SetArticleTemplate',
  setup() {
    const templateForm = ref<{ content: string; selectedTemplate: string | null }>({
      content: '',
      selectedTemplate: null,
    });

    const templates = ref<string[]>([]);  // 存储模板文件名
    const newTemplate = ref<{ templateId: string; content: string }>({
      templateId: '',
      content: '',
    });

    const fetchTemplates = async () => {
      try {
        const response = await http.get('/send/getTemplateList');
        console.log("Full response:", response); // 打印整个响应对象

        if (response.data && response.data.data && response.data.data.templates) {
          console.log("Received templates:", response.data.data.templates);  // 访问 response.data.data.templates
          templates.value = response.data.data.templates.filter((template: string) => template.endsWith('.txt'));
        } else {
          console.error("Templates not found in the response");
        }
      } catch (error) {
        ElMessage.error('获取模板列表失败');
        console.error(error);
      }
    };





    const fetchTemplate = async (templateId: string) => {
      if (!templateId) return;

      console.log("Fetching template with ID:", templateId);  // 打印 templateId

      try {
        const response = await http.get(`/send/getTemplate/${templateId}`);
        console.log("Full response:", response);  // 打印完整的响应
        console.log("response.data:", response.data);  // 打印 response.data
        console.log("Fetched template content:", response.data.data ? response.data.data.content : "Content not found");  // 访问正确的路径

        if (response.data && response.data.data && response.data.data.content) {
          templateForm.value.content = response.data.data.content;
        } else {
          console.error("Content not found in response data");
        }
      } catch (error) {
        ElMessage.error('获取文章模板内容失败');
        console.error(error);
      }
    };





    // 保存模板
    const saveTemplate = async () => {
      if (!templateForm.value.selectedTemplate || !templateForm.value.content.trim()) {
        ElMessage.warning('模板名称和内容不能为空');
        return;
      }

      try {
        // 发送模板名称（selectedTemplate）以及内容（content）到后端
        await http.post('/send/setTemplate', {
          templateId: templateForm.value.selectedTemplate,  // 传递模板ID
          content: templateForm.value.content,             // 传递模板内容
        });
        ElMessage.success('模板保存成功');
      } catch (error) {
        ElMessage.error('保存失败');
        console.error(error);
      }
    };

    // 新增模板
    const addTemplate = async () => {
      if (!newTemplate.value.templateId || !newTemplate.value.content.trim()) {
        ElMessage.warning('模板名称和内容不能为空');
        return;
      }

      try {
        await http.post('/send/addTemplate', {
          templateId: newTemplate.value.templateId,
          content: newTemplate.value.content,
        });
        ElMessage.success('模板新增成功');
        fetchTemplates(); // 重新加载模板列表
        newTemplate.value.templateId = ''; // 清空新增模板名称
        newTemplate.value.content = ''; // 清空新增模板内容
      } catch (error) {
        ElMessage.error('新增模板失败');
        console.error(error);
      }
    };

    const previewTemplate = () => {
      const content = templateForm.value.content;
      if (!content.trim()) {
        ElMessage.warning('模板内容为空，无法预览');
        return;
      }

      // 创建一个新的 HTML 文档
      const previewWindow = window.open('', '_blank');

      // 确保 previewWindow 不为 null
      if (previewWindow) {
        previewWindow.document.write(`
      <html>
        <head>
          <title>模板预览</title>
          <style>
            body {
              font-family: Arial, sans-serif;
              padding: 20px;
            }
            h1 {
              font-size: 24px;
              margin-bottom: 10px;
            }
            .template-content {
              white-space: pre-wrap;
              word-wrap: break-word;
              font-size: 16px;
            }
          </style>
        </head>
        <body>
          <h1>模板预览</h1>
          <div class="template-content">${content}</div>
        </body>
      </html>
    `);
        previewWindow.document.close();
      } else {
        ElMessage.error('无法打开预览窗口');
      }
    };


    // 初始化
    onMounted(() => {
      fetchTemplates();
    });

    return {
      templateForm,
      saveTemplate,
      templates,
      fetchTemplate,
      newTemplate,
      addTemplate,
      previewTemplate,
    };
  },
});
</script>

<style scoped>
.el-button {
  margin-top: 20px;
}
</style>
