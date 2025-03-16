// src/router/index.ts
import {createRouter, createWebHistory, RouteRecordRaw} from 'vue-router';
import UserLogin from '../components/Login.vue';
import UserRegister from '../components/Register.vue';
import Index from '../views/Index.vue';
import DashboardView from '../views/Dashboard.vue';
import SettingsView from '../views/Settings.vue';
import Documents from '../views/Documents.vue';
import SetGPT from "@/views/SetGPT.vue";
import SetArticleTemplate from "@/views/SetArticleTemplate.vue";
import AddWechat from "@/views/AddWechat.vue";
import UserManagement from '@/views/UserManagement.vue';
import PostPage from "@/views/PostPage.vue";
import TitlePage from "@/views/TitlePage.vue";
import TimeJob from "@/views/TimeJob.vue";  // 新增用户管理页面

const routes: Array<RouteRecordRaw> = [
    {
        path: '/login',
        name: 'Login',
        component: UserLogin,
    },
    {
        path: '/register',
        name: 'Register',
        component: UserRegister,
    },
    {
        path: '/',
        name: 'Index',
        component: Index,
        children: [
            {
                path: '/dashboard',
                name: 'DashboardView',
                component: DashboardView,
                meta: {requiresAuth: true},
            },
            {
                path: '/settings',
                name: 'SettingsView',
                component: SettingsView,
                meta: {requiresAuth: true},
            },
            {
                path: '/documents',
                name: 'Documents',
                component: Documents,
            },
            {
                path: '',
                redirect: 'dashboard',
            }, {
                path: '/addWechat',
                name: 'AddWechat',
                component: AddWechat,
            },
            {
                path: '/setPage',
                name: 'SetArticleTemplate',
                component: SetArticleTemplate,
            },
            {
                path: '/setGPT',
                name: 'SetGPT',
                component: SetGPT,
            },
            {path: '/userManage', component: UserManagement},  // 新增路由
            {path: '/postPage', component: PostPage},  // 新增路由
            {path: '/titleSet',  component: TitlePage},  // 新增路由
            {path: '/timeJob',  component: TimeJob},  // 新增路由
        ],
        meta: {requiresAuth: true},
    },
    {
        path: '/:pathMatch(.*)*',
        redirect: '/login',
    },
];

const router = createRouter({
    history: createWebHistory(),
    routes,
});

// 路由守卫
router.beforeEach((to, from, next) => {
    const token = localStorage.getItem('token');
    if (to.matched.some(record => record.meta.requiresAuth)) {
        if (!token) {
            next('/login');
        } else {
            next();
        }
    } else {
        next();
    }
});

export default router;
