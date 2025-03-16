<template>
  <div>

    <el-card>
      <p>用户面板</p>
      <!-- 判断是否显示错误消息 -->
      <el-alert
          v-if="errorMessage"
          type="error"
          title="权限不足"
          :description="errorMessage"
          show-icon
      ></el-alert>

      <el-table
          :data="userList"
          style="width: 100%; margin-top: 20px;"
          border
          stripe
      >
        <el-table-column prop="id" label="ID" width="50"/>
        <el-table-column prop="username" label="用户名" width="150"/>
        <el-table-column prop="updated_at" label="更新时间" width="250"/>
        <el-table-column prop="power" label="权限" width="100"/>
        <el-table-column label="操作" width="180">
          <template #default="scope">
            <el-button type="danger" size="small" @click="deleteUser(scope.row.id)">删除</el-button>
            <el-button type="primary" size="small" @click="editUser(scope.row)">编辑</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 编辑用户区域 -->
    <el-card class="card form-container" style="width: 50%; margin-top: 20px;">
      <el-form :model="editUserData" label-width="120px" class="vertical-form">
        <el-form-item label="用户名">
          <el-input v-model="editUserData.username" placeholder="请输入用户名"/>
        </el-form-item>
        <el-form-item label="密码" :rules="{ required: true, message: '请输入密码', trigger: 'blur' }">
          <el-input v-model="editUserData.password" type="password" placeholder="请输入密码"></el-input>
        </el-form-item>

        <el-form-item label="权限">
          <el-input v-model="editUserData.power" type="number" placeholder="请输入权限等级"/>
        </el-form-item>
        <el-form-item style="text-align: center;">
          <el-button type="primary" @click="saveUserEdit">保存修改</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script lang="ts">
import {defineComponent, ref, onMounted} from 'vue';
import http from '../services/http'; // 确保你已经配置了 http 请求
import {useUserStore} from '@/store/user';
import {ElMessageBox,ElMessage} from "element-plus";

interface User {
  id: number;
  username: string;
  password: string;
  updated_at: string;
  power: number;
}

export default defineComponent({
  name: 'DashboardView',
  setup() {
    const userList = ref<User[]>([]); // 用来存储从后端获取到的用户数据
    const errorMessage = ref<string>(''); // 用来存储错误信息
    const userStore = useUserStore();

    // 编辑表单数据
    const editUserData = ref<User>({
      id: 0,
      username: '',
      updated_at: '',
      password: '',
      power: 0,
    });

    // 获取当前用户名
    const username = userStore.username;

    // 获取用户数据
    const getUsers = async () => {
      try {
        const response = await http.post('/user/getUser', {username});

        // 如果返回成功并且有数据，更新用户列表
        if (response.data && response.data.user) {
          userList.value = response.data.user;
          errorMessage.value = ''; // 清空错误信息
        }

        // 如果权限不足，显示错误信息
        if (response.data && response.data.error) {
          errorMessage.value = response.data.error;
          userList.value = []; // 清空用户列表
        }
      } catch (error) {
        console.error('获取用户数据失败', error);
        errorMessage.value = '获取用户数据失败，请稍后再试';
      }
    };

    // 删除用户
    const deleteUser = async (userId: number) => {
      try {
        // 使用 ElMessageBox.confirm 来确认删除操作
        const result = await ElMessageBox.confirm('确定要删除该用户吗?', '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning',
        });

        if (result === 'confirm') {
          // 如果用户确认删除，调用删除接口
          await http.post('/user/deleteUser', { id: userId });
          // 删除成功后更新用户列表
          userList.value = userList.value.filter((user) => user.id !== userId);
          ElMessage.success('用户删除成功');
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error('删除失败', error);
          ElMessage.error('删除用户失败');
        }
      }
    };


    // 编辑用户
    const editUser = (user: User) => {
      // 将选中的用户数据赋值到编辑表单
      editUserData.value = {...user};
    };

    const saveUserEdit = async () => {
      try {
        // 使用 ElMessageBox.confirm 来确认提交更改
        const result = await ElMessageBox.confirm('确定要保存更改吗?', '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'info',
        });

        if (result === 'confirm') {
          // 如果用户确认提交，调用编辑接口
          await http.post('/user/editUser', editUserData.value);
          // 编辑成功后，刷新用户数据
          await getUsers();
          ElMessage.success('用户信息更新成功');
        }
      } catch (error) {
        if (error !== 'cancel') {
          console.error('编辑失败', error);
          ElMessage.error('编辑用户失败');
        }
      }
    };


    // 初始化时获取用户数据
    onMounted(() => {
      getUsers();
    });

    return {
      userList,
      errorMessage,
      editUserData,
      getUsers,
      deleteUser,
      editUser,
      saveUserEdit
    };
  },
});
</script>

<style scoped>
h2 {
  margin-bottom: 20px;
}

.card {
  padding: 20px;
  background-color: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.form-container {
  margin-top: 20px;
}

.vertical-form {
  display: flex;
  flex-direction: column;
}

.el-form-item {
  margin-bottom: 20px;
}
</style>
