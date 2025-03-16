<!-- src/views/Index.vue -->
<template>
  <el-container style="height: 100vh;">
    <el-aside width="200px" style="background-color: #545c64;">
      <el-menu
          :default-active="activeMenu"
          class="el-menu-vertical-demo"
          @select="handleSelect"
          background-color="#545c64"
          text-color="#fff"
          active-text-color="#ffd04b"
      >
        <el-menu-item index="1">
          <i class="el-icon-menu"></i>
          <span>仪表板</span>
        </el-menu-item>
        <el-menu-item index="2">
          <i class="el-icon-setting"></i>
          <span>设置</span>
        </el-menu-item>
        <el-menu-item index="3">
          <i class="el-icon-document"></i>
          <span>文档</span>
        </el-menu-item>
        <el-menu-item index="4">
          <i class="el-icon-document"></i>
          <span>添加微信号</span>
        </el-menu-item>
        <el-menu-item index="5">
          <i class="el-icon-document"></i>
          <span>设置文章模板</span>
        </el-menu-item>
        <el-menu-item index="6">
          <i class="el-icon-document"></i>
          <span>设置gpt</span>
        </el-menu-item>
        <el-menu-item index="7">
          <i class="el-icon-document"></i>
          <span>用户管理</span>
        </el-menu-item>
        <el-menu-item index="8">
          <i class="el-icon-document"></i>
          <span>微信发文</span>
        </el-menu-item>
        <el-menu-item index="9">
          <i class="el-icon-document"></i>
          <span>标题设置</span>
        </el-menu-item>
        <el-menu-item index="10">
          <i class="el-icon-document"></i>
          <span>定时任务</span>
        </el-menu-item>
        <el-menu-item index="11">
          <i class="el-icon-switch-button"></i>
          <span @click="handleLogout">退出</span>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header style="background-color: #fff; padding: 0 20px; display: flex; justify-content: flex-end; align-items: center;">
        <span>欢迎, {{ username }}</span>
      </el-header>
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script lang="ts">
import { defineComponent, ref, onMounted, watch } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { ElMessage } from 'element-plus';
import { useUserStore } from '@/store/user';

export default defineComponent({
  name: 'IndexPage',
  setup() {
    const router = useRouter();
    const route = useRoute();
    const activeMenu = ref<string>('1');
    const userStore = useUserStore();
    const username = ref<string>('');

    const setActiveMenu = () => {
      switch (route.path) {
        case '/dashboard':
          activeMenu.value = '1';
          break;
        case '/settings':
          activeMenu.value = '2';
          break;
        case '/documents':
          activeMenu.value = '3';
          break;
        case '/addWechat':
          activeMenu.value = '4';
          break;
        case '/setPage':
          activeMenu.value = '5';
          break;
        case '/setGPT':
          activeMenu.value = '6';
          break;
        case '/userManage':
          activeMenu.value='7';
          break;
        case '/postPage':
          activeMenu.value='8';
          break;
        case '/titleSet':
          activeMenu.value='9';
          break;
        case '/timeJob':
          activeMenu.value='10';
          break;
        default:
          activeMenu.value = '1';
          break;
      }
    };

    onMounted(() => {
      if (userStore.token) {
        username.value = userStore.username;
        setActiveMenu();
      } else {
        router.push('/login');
      }
    });

    watch(
        () => route.path,
        () => {
          setActiveMenu();
        }
    );

    const handleSelect = (key: string) => {
      switch (key) {
        case '1':
          router.push('/dashboard');
          break;
        case '2':
          router.push('/settings');
          break;
        case '3':
          router.push('/documents');
          break;
        case '4':
          router.push('/addWechat');
          break;
        case '5':
          router.push('/setPage');
          break;
        case '6':
          router.push('/setGPT');
          break;
        case '7':
          router.push('/userManage');
          break;
        case '8':
          router.push('/postPage');
          break;
        case '9':
          router.push('/titleSet');
          break;
        case '10':
          router.push('/timeJob');
          break;
        default:
          break;
      }
    };

    const handleLogout = async () => {
      try {
        await userStore.logout(); // 清除 store 中的 token 和用户名
        ElMessage.success('已成功退出');

        router.push('/login'); // 跳转到登录页
      } catch (error) {
        ElMessage.error('退出失败');
        console.error('Logout error:', error);
      }
    };


    return {
      activeMenu,
      handleSelect,
      handleLogout,
      username,
    };
  },
});
</script>

<style scoped>
.el-aside {
  border-right: 1px solid #ebeef5;
}
</style>
