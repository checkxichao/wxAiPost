// src/services/http.ts
import axios from 'axios';
import { ElMessage } from 'element-plus';

const http = axios.create({
    baseURL: 'http://localhost:8888', // 后端服务器地址
    timeout: 5000,
});

// 请求拦截器
http.interceptors.request.use(
    (config) => {
        const token = localStorage.getItem('token');
        if (token && config.headers) {
            config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
    },
    (error) => {
        return Promise.reject(error);
    }
);

// 响应拦截器
http.interceptors.response.use(
    (response) => {
        return response;
    },
    (error) => {
        if (error.response) {
            switch (error.response.status) {
                case 401:
                    ElMessage.error('未授权，请登录');
                    break;
                case 403:
                    ElMessage.error('拒绝访问');
                    break;
                case 404:
                    ElMessage.error('请求地址出错');
                    break;
                default:
                    ElMessage.error(error.response.data.message || '服务器错误');
                    break;
            }
        } else {
            ElMessage.error('网络错误');
        }
        return Promise.reject(error);
    }
);

export default http;
