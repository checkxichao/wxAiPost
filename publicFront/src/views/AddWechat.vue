<template>
  <div>
    <!-- 表格区域 -->
    <el-card class="card">
      <el-table :data="paginatedWechatAccounts" style="width: 100%" stripe>
        <el-table-column prop="id" label="ID" width="80"></el-table-column>
        <el-table-column prop="name" label="公众号名" width="150"></el-table-column>
        <el-table-column prop="wxid" label="微信openid" width="180"></el-table-column>
        <el-table-column prop="secret" label="密钥" width="300"></el-table-column>
        <el-table-column prop="isuse" label="是否使用" width="80">
          <template #default="scope">
            <el-tag :type="scope.row.isuse ? 'success' : 'info'">
              {{ scope.row.isuse ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="bindWechat" label="绑定者" width="150"></el-table-column>
        <el-table-column prop="created_at" label="时间" width="200">
          <template #default="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180">
          <template #default="scope">
            <el-button type="danger" size="small" @click="deleteWechatAccount(scope.row.wxid)">删除</el-button>
            <el-button type="primary" size="small" @click="editWechatAccount(scope.row)">编辑</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页组件 -->
      <div v-if="wechatAccounts.length > 0" class="pagination-wrapper">
        <el-pagination
            @size-change="handleSizeChange"
            @current-change="handleCurrentChange"
            :current-page="currentPage"
            :page-sizes="[10, 20, 50, 100]"
            :page-size="pageSize"
            layout="total, sizes, prev, pager, next, jumper"
            :total="wechatAccounts.length">
        </el-pagination>
      </div>
    </el-card>

    <!-- 表单区域 -->
    <div class="form-wrapper">
      <!-- 添加表单 -->
      <el-card class="card form-container" style="width: 50%;">
        <el-form :model="newWechat" label-width="120px" class="vertical-form">
          <el-form-item label="公众号名" :rules="formRules.name">
            <el-input v-model="newWechat.name" placeholder="请输入公众号名"></el-input>
          </el-form-item>
          <el-form-item label="微信openid" :rules="formRules.wxid">
            <el-input v-model="newWechat.wxid" placeholder="请输入微信openid"></el-input>
          </el-form-item>
          <el-form-item label="密钥" :rules="formRules.secret">
            <el-input v-model="newWechat.secret" placeholder="请输入密钥"></el-input>
          </el-form-item>
          <el-form-item label="绑定者" :rules="formRules.bindWechat">
            <el-input v-model="newWechat.bindWechat" placeholder="请输入绑定者"></el-input>
          </el-form-item>
          <el-form-item style="text-align: center;">
            <el-button type="primary" @click="addWechatAccount">提交</el-button>
          </el-form-item>
        </el-form>
      </el-card>

      <!-- 编辑表单 -->
      <el-card class="card form-container" style="width: 50%;">
        <el-form :model="editWechat" label-width="120px" class="vertical-form">
          <el-form-item label="公众号名">
            <el-input v-model="editWechat.name" placeholder="请输入公众号名"></el-input>
          </el-form-item>
          <el-form-item label="微信openid">
            <el-input v-model="editWechat.wxid" placeholder="请输入微信openid"></el-input>
          </el-form-item>
          <el-form-item label="微信secret">
            <el-input v-model="editWechat.secret" placeholder="请输入微信secret"></el-input>
          </el-form-item>
          <el-form-item label="绑定者">
            <el-input v-model="editWechat.bindWechat" placeholder="请输入微信绑定者"></el-input>
          </el-form-item>
          <el-form-item label="是否使用">
            <el-radio-group v-model="editWechat.isuse">
              <el-radio :label="true">是</el-radio>
              <el-radio :label="false">否</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item style="text-align: center;">
            <el-button type="primary" @click="saveEditWechatAccount">保存修改</el-button>
          </el-form-item>
        </el-form>
      </el-card>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, onMounted, computed } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import http from '../services/http'; // 确保你已经配置了 http 请求

interface WechatAccount {
  id: number;
  name: string;
  wxid: string;
  secret: string;
  bindWechat: string;
  isuse: boolean;
  created_at: string;
}

export default defineComponent({
  name: 'AddWechat',
  setup() {
    const wechatAccounts = ref<WechatAccount[]>([]);
    const newWechat = ref<{ name: string; wxid: string; secret: string; bindWechat: string }>({
      name: '',
      wxid: '',
      secret: '',
      bindWechat: '',
    });

    // 编辑表单数据
    const editWechat = ref<{ id: number; name: string; wxid: string; isuse: boolean; bindWechat: string; secret: string }>({
      id: 0,
      name: '',
      wxid: '',
      isuse: true,
      bindWechat: '',
      secret: '',
    });

    // 分页相关状态
    const currentPage = ref<number>(1); // 当前页码
    const pageSize = ref<number>(10); // 每页显示数量

    // 计算当前页需要显示的数据
    const paginatedWechatAccounts = computed(() => {
      const start = (currentPage.value - 1) * pageSize.value;
      const end = currentPage.value * pageSize.value;
      return wechatAccounts.value.slice(start, end);
    });

    // 表单验证规则
    const formRules = {
      name: { required: true, message: '请输入公众号名', trigger: 'blur' },
      wxid: { required: true, message: '请输入微信openid', trigger: 'blur' },
      secret: { required: true, message: '请输入密钥', trigger: 'blur' },
      bindWechat: { required: true, message: '请输入绑定者', trigger: 'blur' },
    };

    const fetchWechatAccounts = async () => {
      try {
        const response = await http.get('/wechat/getWechat');
        wechatAccounts.value = response.data.info;
      } catch (error) {
        ElMessage.error('获取微信信息失败');
      }
    };

    const addWechatAccount = async () => {
      if (!newWechat.value.name || !newWechat.value.wxid || !newWechat.value.secret || !newWechat.value.bindWechat) {
        ElMessage.warning('请填写完整信息');
        return;
      }
      try {
        await http.post('/wechat/addWechat', newWechat.value);
        ElMessage.success('添加成功');
        newWechat.value.name = '';
        newWechat.value.wxid = '';
        newWechat.value.secret = '';
        newWechat.value.bindWechat = '';
        await fetchWechatAccounts();
      } catch (error) {
        ElMessage.error('添加微信号失败');
        await fetchWechatAccounts();
      }
    };

    const deleteWechatAccount = async (wxid: string) => {
      try {
        await ElMessageBox.confirm('确定要删除该微信号吗?', '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning',
        });

        // 使用 POST 方法发送 wxid
        await http.post('/wechat/deleteWechat', { wxid });

        // 更新本地数据
        wechatAccounts.value = wechatAccounts.value.filter((account) => account.wxid !== wxid);

        ElMessage.success('删除成功');
      } catch (error) {
        if (error !== 'cancel') {
          ElMessage.error('删除失败');
        }
      }
    };

    const editWechatAccount = (account: WechatAccount) => {
      // 将当前微信账号的内容加载到编辑表单中
      editWechat.value = {
        id: account.id,
        name: account.name,
        wxid: account.wxid,
        isuse: account.isuse,
        bindWechat: account.bindWechat,
        secret: account.secret,
      };
    };

    // 保存修改的方法
    const saveEditWechatAccount = async () => {
      if (!editWechat.value.name || !editWechat.value.wxid || !editWechat.value.secret || !editWechat.value.bindWechat) {
        ElMessage.warning('请填写完整信息');
        return;
      }
      try {
        // 使用 POST 请求提交修改
        await http.post(`/wechat/editWechat`, editWechat.value);
        ElMessage.success('修改成功');
        await fetchWechatAccounts();  // 更新微信号列表
        // 重置编辑表单
        editWechat.value = {
          id: 0,
          name: '',
          wxid: '',
          isuse: true,
          bindWechat: '',
          secret: '',
        };
      } catch (error) {
        ElMessage.error('修改失败');
      }
    };

    onMounted(() => {
      fetchWechatAccounts();
    });

    // 格式化日期
    const formatDate = (dateString: string): string => {
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
    };

    // 分页处理函数：页面大小变化
    const handleSizeChange = (newSize: number) => {
      pageSize.value = newSize;
      currentPage.value = 1; // 重置到第一页
    };

    // 分页处理函数：当前页变化
    const handleCurrentChange = (newPage: number) => {
      currentPage.value = newPage;
    };

    return {
      wechatAccounts,
      newWechat,
      editWechat,
      addWechatAccount,
      deleteWechatAccount,
      editWechatAccount,
      saveEditWechatAccount,
      paginatedWechatAccounts,
      currentPage,
      pageSize,
      handleSizeChange,
      handleCurrentChange,
      formatDate,
      formRules,
    };
  },
});
</script>

<style scoped>
.card {
  padding: 20px;
  background-color: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  margin-bottom: 20px; /* 给表格下方增加空隙 */
}

.form-wrapper {
  display: flex;
  gap: 20px; /* 两个表单之间增加间距 */
}

.form-container {
  flex: 1;
}

.vertical-form {
  display: flex;
  flex-direction: column; /* 表单项竖直排列 */
}

.el-form-item {
  margin-bottom: 20px; /* 控制表单项之间的间距 */
}

/* 分页组件样式 */
.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}
</style>
