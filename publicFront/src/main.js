// src/main.ts
import { createApp } from 'vue';
import App from './App.vue';
import router from './router';
import ElementPlus, { ElMessage } from 'element-plus';
import 'element-plus/dist/index.css';
import { createPinia } from 'pinia';
// 如果选择创建 styles.css，可以保留这行
// import './assets/styles.css';

const app = createApp(App);

// 全局挂载 ElMessage
app.config.globalProperties.$message = ElMessage;

// 使用 Pinia
const pinia = createPinia();
app.use(pinia);

// 使用 Element Plus
app.use(ElementPlus);

// 使用路由
app.use(router);

// 挂载应用
app.mount('#app');
// 捕获并忽略 ResizeObserver 的警告
const observerError = window.console.error;
window.console.error = (...args) => {
    if (args[0] && typeof args[0] === 'string' && args[0].includes('ResizeObserver loop completed with undelivered notifications')) {
        return; // 忽略警告
    }
    observerError(...args);
};
