import { defineStore } from 'pinia';

export const useUserStore = defineStore('user', {
    state: () => ({
        token: localStorage.getItem('token') || '',
        username: localStorage.getItem('username') || '',
        role: localStorage.getItem('role') || '',
    }),
    actions: {
        setUser(user: { token: string; username: string; role: string }) {
            this.token = user.token;
            this.username = user.username;
            this.role = user.role;

            // 存储到 localStorage
            localStorage.setItem('token', user.token);
            localStorage.setItem('username', user.username);
            localStorage.setItem('role', user.role);
        },
        logout() {
            this.token = '';
            this.username = '';
            this.role = '';

            // 从 localStorage 中移除
            localStorage.removeItem('token');
            localStorage.removeItem('username');
            localStorage.removeItem('role');
        },
    },
});
