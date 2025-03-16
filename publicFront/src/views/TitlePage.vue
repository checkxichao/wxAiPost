<template>
  <div class="title-page">
    <h1>标题管理</h1>

    <!-- 标题列表 -->
    <section>
      <h2>标题列表</h2>
      <el-table v-if="titles && titles.length > 0" :data="titles" style="width: 100%" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="30" :selectable="selectable"></el-table-column>
        <el-table-column prop="id" label="ID" width="80"></el-table-column>
        <el-table-column prop="title" label="标题" width="450">
          <template #default="scope">
            <div v-if="scope.row.isEditing">
              <el-input v-model="scope.row.title" size="small"></el-input>
            </div>
            <div v-else>
              {{ scope.row.title }}
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="updatedAt" label="更新时间" width="120">
          <template #default="scope">
            {{ formatDate(scope.row.updatedAt) }}
          </template>
        </el-table-column>
        <!-- 操作列 -->
        <el-table-column label="操作" width="120">
          <template #default="scope">
            <div class="action-buttons">
              <div v-if="scope.row.isEditing">
                <el-button class="action-button" type="success" @click="saveEdit(scope.row)">保存</el-button>
                <el-button class="action-button" type="info" @click="cancelEdit(scope.row)">取消</el-button>
              </div>
              <div v-else>
                <el-button class="action-button" type="primary" @click="enableEdit(scope.row)">修改</el-button>
                <el-button class="action-button" type="danger"
                           @click="confirmDeleteTitle(scope.row.id, scope.row.title)">删除
                </el-button>
              </div>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <div v-else>
        <el-empty description="暂无数据"></el-empty>
      </div>
      <div v-if="selectedTitles.length > 0" class="bulk-delete-wrapper">
        <el-button type="danger" @click="bulkDelete">批量删除</el-button>
      </div>
      <!-- 分页组件 -->
      <div v-if="totalTitles > pageSize" class="pagination-wrapper">
        <el-pagination
            @size-change="handleSizeChange"
            @current-change="handleCurrentChange"
            :current-page="currentPage"
            :page-sizes="[10, 20, 50, 100]"
            :page-size="pageSize"
            layout="total, sizes, prev, pager, next, jumper"
            :total="totalTitles">
        </el-pagination>
      </div>
    </section>

    <!-- 提交单个标题 -->
    <section>
      <h2>提交单个标题</h2>
      <el-form ref="formRef" :rules="rules" :model="form" label-width="0px">
        <div class="input-group">
          <el-input
              v-model="form.singleTitle"
              placeholder="输入标题"
              clearable
              name="singleTitle"
          ></el-input>
          <el-button type="primary" @click="submitSingleTitle">提交</el-button>
        </div>
      </el-form>
    </section>

    <!-- 提交多个标题 -->
    <section>
      <h2>提交多个标题</h2>
      <el-form ref="batchFormRef" :rules="batchRules" :model="form" label-width="0px">
        <div class="input-group">
          <el-input
              type="textarea"
              v-model="form.multipleTitles"
              placeholder="每行输入一个标题"
              rows="5"
              clearable
              name="multipleTitles"
          ></el-input>
          <el-button type="primary" @click="submitMultipleTitles">批量提交</el-button>
        </div>
        <!-- 新增后缀输入框 -->
        <div class="input-group">
          <el-input
              v-model="form.suffix"
              placeholder="输入后缀"
              clearable
              name="suffix"
          ></el-input>
        </div>
      </el-form>
    </section>
  </div>
</template>

<script lang="ts">
import { Options, Vue } from 'vue-class-component';
import { ElMessage, FormInstance, ElMessageBox } from 'element-plus';
import axios from 'axios';
import http from '../services/http';

interface BackendTitleItem {
  id: number;
  title: string;
  createdAt: string;
  updatedAt: string;
}

interface TitleItem {
  id: number;
  title: string;
  createdAt: string;
  updatedAt: string;
  isEditing: boolean; // 编辑状态
  originalTitle: string; // 原始标题，用于取消编辑
}

interface TitleListResponse {
  data: BackendTitleItem[];
  total: number;
}

interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

@Options({
  name: 'TitlePage',
})
export default class TitlePage extends Vue {
  // 数据列表
  titles: TitleItem[] = [];
  selectedTitles: TitleItem[] = [];  // 选中的标题

  // 表单数据
  form = {
    singleTitle: '',
    multipleTitles: '',
    suffix: '', // 新增后缀字段
  };

  // 表单验证规则
  rules = {
    singleTitle: [
      { required: true, message: '标题不能为空', trigger: 'blur' },
      { min: 1, message: '标题长度至少为1个字符', trigger: 'blur' },
    ],
  };

  batchRules = {
    multipleTitles: [
      { required: true, message: '至少输入一个标题', trigger: 'blur' },
      {
        validator: (_: any, value: string, callback: any) => {
          const titles = value.split('\n').map(title => title.trim()).filter(title => title.length > 0);
          if (titles.length === 0) {
            callback(new Error('至少输入一个标题'));
          } else {
            callback();
          }
        },
        trigger: 'blur',
      },
    ],
    suffix: [ // 新增后缀的验证规则（可选）
      { required: false, message: '后缀不能为空', trigger: 'blur' }, // 改为非必填
      { max: 10, message: '后缀长度不能超过10个字符', trigger: 'blur' },
    ],
  };

  private isMountedFlag: boolean = false;

  // 分页相关状态
  currentPage: number = 1; // 当前页码
  pageSize: number = 10; // 每页显示数量
  totalTitles: number = 0; // 总记录数

  // 生命周期钩子，组件挂载时获取标题列表
  mounted() {
    this.isMountedFlag = true;
    this.fetchTitles();
  }

  beforeUnmount() {
    this.isMountedFlag = false;
  }

  handleSelectionChange(selectedItems: TitleItem[]) {
    this.selectedTitles = selectedItems;
  }

  async bulkDelete() {
    const selectedIds = this.selectedTitles.map(title => title.id);
    try {
      const response = await http.post('/send/deleteTitles', { ids: selectedIds });
      ElMessage.success('批量删除成功');
      this.fetchTitles();
      this.selectedTitles = []; // 清空选中的标题
    } catch (error) {
      ElMessage.error('批量删除失败');
    }
  }

  // 获取标题列表
  async fetchTitles() {
    try {
      console.log('当前请求的分页参数:', this.currentPage, this.pageSize); // 调试日志

      const response = await http.get<ApiResponse<TitleListResponse>>('/send/getTitleList', {
        params: {
          page: this.currentPage,
          pageSize: this.pageSize,
        },
      });

      console.log('获取到的标题列表:', response.data); // 调试日志

      // 解构 response.data.data
      const { data, total } = response.data.data;

      if (Array.isArray(data)) {
        // 映射字段名称并添加 isEditing 和 originalTitle 属性
        const mappedTitles = data.map((item: BackendTitleItem) => ({
          id: item.id,
          title: item.title,
          createdAt: item.createdAt,
          updatedAt: item.updatedAt,
          isEditing: false, // 添加编辑状态
          originalTitle: item.title, // 用于取消编辑时恢复原始标题
        }));

        if (this.isMountedFlag) {
          this.titles = mappedTitles; // 更新标题数据
          this.totalTitles = total; // 更新总记录数
        }
      } else {
        console.error('响应数据格式错误，预期为数组:', response.data);
        ElMessage.error('获取标题列表失败，数据格式错误');
        if (this.isMountedFlag) {
          this.titles = [];
          this.totalTitles = 0;
        }
      }
    } catch (error) {
      console.error('获取标题列表失败:', error);
      ElMessage.error('获取标题列表失败');
      if (this.isMountedFlag) {
        this.titles = [];
        this.totalTitles = 0;
      }
    }
  }

  // 提交单个标题
  async submitSingleTitle() {
    try {
      // 验证表单
      const form = this.$refs.formRef as FormInstance;
      await form.validate();
    } catch (error) {
      // 验证失败
      return;
    }

    const trimmedTitle = this.form.singleTitle.trim();
    console.log('提交的标题:', trimmedTitle); // 调试日志

    try {
      const response = await http.post('/send/setTitle', {
        title: trimmedTitle,
      });
      ElMessage.success('标题提交成功');
      this.form.singleTitle = '';
      this.fetchTitles(); // 刷新列表
    } catch (error: unknown) {
      console.error('提交单个标题失败:', error);
      if (axios.isAxiosError(error)) {
        if (error.response?.data?.error) {
          ElMessage.error(`提交失败: ${error.response.data.error}`);
        } else {
          ElMessage.error('提交标题失败');
        }
      } else {
        ElMessage.error('提交标题失败');
      }
    }
  }

  selectable(row: TitleItem) {
    return true;  // 你可以自定义某些条件来禁用复选框
  }

  // 提交多个标题
  async submitMultipleTitles() {
    try {
      // 验证表单
      console.log("batchFormRef:", this.$refs.batchFormRef); // 调试日志
      const batchForm = this.$refs.batchFormRef as FormInstance;
      await batchForm.validate();
    } catch (error) {
      // 验证失败
      console.log(error);
      return;
    }

    const suffix = this.form.suffix.trim(); // 获取并修剪后缀
    const titlesArray = this.form.multipleTitles
        .split('\n')
        .map(title => title.trim())
        .filter(title => title.length > 0)
        .map(title => {
          // 添加后缀并包裹在《》中
          let newTitle = suffix ? `《${title}》${suffix}` : `《${title}》`;

          // 检查长度是否超过64
          if (newTitle.length > 64) {
            const wrapLength = 2; // 《和》共2个字符
            const suffixLength = suffix.length;
            const maxTitleLength = 64 - wrapLength - suffixLength;
            let trimmedTitle = title.slice(0, maxTitleLength);
            newTitle = suffix ? `《${trimmedTitle}》${suffix}` : `《${trimmedTitle}》`;
            ElMessage.warning(`标题 "${title}" 已被截断为 "${newTitle}" 以符合最大长度限制。`);
          }

          return newTitle;
        });

    console.log('提交的多个标题:', titlesArray); // 调试日志

    try {
      // 发送批量标题提交请求
      const response = await http.post('/send/setTitleBatch', {
        titles: titlesArray,
      });
      ElMessage.success('多个标题提交成功');
      this.form.multipleTitles = '';
      this.form.suffix = ''; // 清空后缀
      this.fetchTitles(); // 刷新列表
    } catch (error: unknown) {
      console.error('提交多个标题失败:', error);
      if (axios.isAxiosError(error)) {
        if (error.response?.data?.error) {
          ElMessage.error(`提交失败: ${error.response.data.error}`);
        } else {
          ElMessage.error('提交多个标题失败');
        }
      } else {
        ElMessage.error('提交多个标题失败');
      }
    }
  }

  // 启用编辑模式
  enableEdit(row: TitleItem) {
    row.isEditing = true;
  }

  // 取消编辑模式
  cancelEdit(row: TitleItem) {
    row.title = row.originalTitle; // 恢复原始标题
    row.isEditing = false;
  }

  // 保存编辑后的标题
  async saveEdit(row: TitleItem) {
    const trimmedTitle = row.title.trim();
    if (trimmedTitle === '') {
      ElMessage.error('标题不能为空');
      return;
    }

    try {
      const response = await http.post('/send/editTitle', {
        id: row.id,
        title: trimmedTitle,
      });
      ElMessage.success('标题修改成功');
      row.isEditing = false;
      this.fetchTitles(); // 刷新列表
    } catch (error: unknown) {
      console.error('修改标题失败:', error);
      if (axios.isAxiosError(error)) {
        if (error.response?.data?.error) {
          ElMessage.error(`修改失败: ${error.response.data.error}`);
        } else {
          ElMessage.error('修改标题失败');
        }
      } else {
        ElMessage.error('修改标题失败');
      }
    }
  }

  // 删除标题
  async deleteTitle(id: number, title: string) {
    try {
      await http.post('/send/deleteTitle', {
        id: id,
        title: title,
      });
      ElMessage.success('标题删除成功');
      this.fetchTitles(); // 刷新列表
    } catch (error: unknown) {
      console.error('删除标题失败:', error);
      if (axios.isAxiosError(error)) {
        if (error.response?.data?.error) {
          ElMessage.error(`删除失败: ${error.response.data.error}`);
        } else {
          ElMessage.error('删除标题失败');
        }
      } else {
        ElMessage.error('删除标题失败');
      }
    }
  }

  // 确认删除标题
  async confirmDeleteTitle(id: number, title: string) {
    try {
      await ElMessageBox.confirm(
          '确定删除这个标题吗?',
          '提示',
          {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning',
          }
      );
      // 用户点击确定，执行删除操作
      this.deleteTitle(id, title);
    } catch (error) {
      // 用户点击取消或关闭对话框，不执行任何操作
      // 可选：显示取消信息
      // ElMessage.info('取消删除');
    }
  }

  // 格式化日期
  formatDate(dateString: string): string {
    if (!dateString) {
      return '无效日期';
    }
    const date = new Date(dateString);
    if (isNaN(date.getTime())) {
      return '无效日期';
    }
    const options: Intl.DateTimeFormatOptions = {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    };
    return date.toLocaleDateString(undefined, options);
  }

  // 分页处理函数：页面大小变化
  handleSizeChange(newSize: number) {
    console.log('每页显示条数变化为:', newSize); // 调试日志
    this.pageSize = newSize;
    this.currentPage = 1; // 重置到第一页
    this.fetchTitles(); // 重新获取数据
  }

  // 分页处理函数：当前页变化
  handleCurrentChange(newPage: number) {
    console.log('当前页码变化为:', newPage); // 调试日志
    this.currentPage = newPage;
    this.fetchTitles(); // 重新获取数据
  }
}
</script>
<style scoped>
.title-page {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

.title-page h1 {
  text-align: center;
  margin-bottom: 30px;
}

section {
  margin-bottom: 40px;
}

.el-table th,
.el-table td {
  text-align: center;
}

/* 移除全宽样式 */
.el-input,
.el-button {
  /* width: 100%; */
  margin-bottom: 10px;
}

/* 新增输入组样式 */
.input-group {
  display: flex;
  align-items: center;
  gap: 10px; /* 输入框和按钮之间的间距 */
}

.input-group .el-input {
  flex: 1; /* 输入框占据剩余空间 */
  min-width: 200px; /* 设置最小宽度以防止过窄 */
}

.input-group .el-button {
  flex: none; /* 按钮保持固定宽度 */
  width: 100px; /* 可根据需要调整按钮宽度 */
}

/* 操作按钮样式 */
.action-buttons {
  display: flex;
  gap: 5px; /* 按钮之间的间距 */
  justify-content: center;
}

.action-button {
  padding: 2px 6px;
  font-size: 12px;
  line-height: 1;
}

/* 分页组件样式 */
.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}
.bulk-delete-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 10px;
}
</style>
