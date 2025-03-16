<template>
  <div>

    <el-form :inline="true" :model="filters" class="filter-form" style="margin-bottom: 20px;">
      <el-form-item label="选择微信ID">
        <el-select
            v-model="filters.belongWxid"
            placeholder="选择微信ID"
            clearable
            style="width: 200px;"
            filterable
            remote
            remote-method="handleWechatSearch"
            :loading="selectLoading"
        >
          <el-option
              v-for="wxid in uniqueWxids"
              :key="wxid"
              :label="wxidToNameMap[wxid] || wxid"
              :value="wxid"
          />
        </el-select>
      </el-form-item>

      <!-- 显示已选择的 wxid 对应的名称 -->
      <p>选择的公众号: {{ wxidToNameMap[filters.belongWxid] || filters.belongWxid }}</p>

      <el-form-item label="日期范围">
        <el-date-picker
            v-model="filters.dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            format="yyyy-MM-dd"
            clearable
            :shortcuts="dateShortcuts"
        />
      </el-form-item>

      <el-button type="primary" @click="fetchMediaPosts">查询</el-button>
    </el-form>

    <el-table :data="tableData" style="width: 100%; margin-top: 20px;" border stripe>
      <!-- 显示公众号名称而不是 wxid -->
      <el-table-column label="微信ID" width="150">
        <template #default="{ row }">
          <!-- 使用映射表 wxidToNameMap 获取公众号名称 -->
          <span>{{ wxidToNameMap[row.belongWxid] || row.belongWxid }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="date" label="日期" width="150"/>
      <el-table-column prop="count" label="发布数量" width="150"/>
    </el-table>

    <!-- 发布数量总和 -->
    <div style="margin-top: 20px;">
      <strong>发布总数量: {{ totalCount }}</strong>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, onMounted, computed } from 'vue';
import http from '../services/http'; // HTTP 服务
import debounce from 'lodash/debounce'; // 引入 debounce 函数

interface MediaPost {
  id: number;
  mediaId: string;
  belongWxid: string;
  state: boolean;
  createdAt: string;
  updatedAt: string;
}

interface DailyPostCount {
  date: string;
  belongWxid: string;
  count: number;
}

export default defineComponent({
  name: 'DashboardView',
  setup() {
    const filters = ref({
      belongWxid: '',
      dateRange: [],
    });

    const tableData = ref<DailyPostCount[]>([]); // 存储统计结果
    const loading = ref(false);
    const uniqueWxids = ref<string[]>([]); // 存储所有唯一的 wxid
    const wxidToNameMap = ref<Record<string, string>>({});  // 存储 wxid 到 name 的映射
    const selectLoading = ref(false); // 选择框的加载状态

    // 获取公众号列表的 API
    const fetchWechatList = async () => {
      try {
        const response = await http.get('/wechat/getWechat');
        if (response.data && response.data.info) {
          // 创建 wxid -> name 的映射
          const wxidMap = response.data.info.reduce((map: Record<string, string>, item: any) => {
            map[item.wxid] = item.name;
            return map;
          }, {});
          wxidToNameMap.value = wxidMap;

          // 更新 uniqueWxids
          uniqueWxids.value = Object.keys(wxidMap);
        } else {
          console.error('Unexpected response structure:', response.data);
        }
      } catch (error) {
        console.error('Error fetching Wechat list:', error);
      }
    };

    // 远程搜索方法，防抖处理
    const handleWechatSearch = debounce(async (query: string) => {
      if (query === '') {
        // 如果查询为空，获取所有公众号
        await fetchWechatList();
        return;
      }
      selectLoading.value = true;
      try {
        const response = await http.get('/wechat/search', { params: { q: query } });
        if (response.data && response.data.info) {
          // 更新 wxidToNameMap 和 uniqueWxids
          const wxidMap = response.data.info.reduce((map: Record<string, string>, item: any) => {
            map[item.wxid] = item.name;
            return map;
          }, {});
          wxidToNameMap.value = wxidMap;
          uniqueWxids.value = Object.keys(wxidMap);
        } else {
          console.error('Unexpected response structure:', response.data);
        }
      } catch (error) {
        console.error('Error searching Wechat:', error);
      } finally {
        selectLoading.value = false;
      }
    }, 300); // 300ms 防抖延迟

    // 获取媒体数据并处理
    const fetchMediaPosts = async () => {
      loading.value = true;
      try {
        const response = await http.post('/wechat/getMediaPost', {
          belongWxid: filters.value.belongWxid || '',  // 如果用户没选择，传空字符串
          startDate: filters.value.dateRange[0] || '',
          endDate: filters.value.dateRange[1] || '',
        });
        if (response.data && response.data.data) {
          const rawData: MediaPost[] = response.data.data;
          const dailyPostCounts: DailyPostCount[] = processPostData(rawData);
          tableData.value = dailyPostCounts;

          // 获取唯一的 wxid
          uniqueWxids.value = Array.from(
              new Set(rawData.map(item => item.belongWxid))
          );

          // 获取公众号列表并创建 wxid -> name 映射
          fetchWechatList();  // 在获取数据后调用此函数
        } else {
          console.error('Unexpected response structure:', response.data);
        }
      } catch (error) {
        console.error('Error fetching media posts:', error);
      } finally {
        loading.value = false;
      }
    };

    // 处理原始数据，按 wxid 和日期分组，统计每个组的数量
    const processPostData = (data: MediaPost[]): DailyPostCount[] => {
      const grouped: Record<string, Record<string, MediaPost[]>> = data.reduce(
          (acc, post) => {
            const date = post.createdAt.split('T')[0]; // 提取日期部分
            const wxid = post.belongWxid;

            if (!acc[wxid]) acc[wxid] = {};
            if (!acc[wxid][date]) acc[wxid][date] = [];
            acc[wxid][date].push(post);

            return acc;
          },
          {} as Record<string, Record<string, MediaPost[]>>
      );

      const counts: DailyPostCount[] = [];
      Object.keys(grouped).forEach(wxid => {
        Object.keys(grouped[wxid]).forEach(date => {
          counts.push({
            date,
            belongWxid: wxid,
            count: grouped[wxid][date].length,
          });
        });
      });

      counts.sort((a, b) => {
        if (a.date < b.date) return 1;
        if (a.date > b.date) return -1;
        return 0;
      });

      return counts;
    };

    // 日期快捷选择
    const dateShortcuts = [
      {
        text: '今天',
        value: () => {
          const today = new Date();
          return [new Date(today.setHours(0, 0, 0, 0)), new Date(today.setHours(23, 59, 59, 999))];
        }
      },
      {
        text: '昨日',
        value: () => {
          const yesterday = new Date();
          yesterday.setDate(yesterday.getDate() - 1);
          return [new Date(yesterday.setHours(0, 0, 0, 0)), new Date(yesterday.setHours(23, 59, 59, 999))];
        }
      },
      {
        text: '最近一周',
        value: () => {
          const end = new Date();
          const start = new Date();
          start.setDate(start.getDate() - 7);
          return [new Date(start.setHours(0, 0, 0, 0)), new Date(end.setHours(23, 59, 59, 999))];
        }
      },
      {
        text: '最近一个月',
        value: () => {
          const end = new Date();
          const start = new Date();
          start.setMonth(start.getMonth() - 1);
          return [new Date(start.setHours(0, 0, 0, 0)), new Date(end.setHours(23, 59, 59, 999))];
        }
      },
      {
        text: '最近三个月',
        value: () => {
          const end = new Date();
          const start = new Date();
          start.setMonth(start.getMonth() - 3);
          return [new Date(start.setHours(0, 0, 0, 0)), new Date(end.setHours(23, 59, 59, 999))];
        }
      },
      {
        text: '最近半年',
        value: () => {
          const end = new Date();
          const start = new Date();
          start.setMonth(start.getMonth() - 6);
          return [new Date(start.setHours(0, 0, 0, 0)), new Date(end.setHours(23, 59, 59, 999))];
        }
      }
    ];

    // 计算发布数量的总和
    const totalCount = computed(() => {
      return tableData.value.reduce((total, item) => total + item.count, 0);
    });

    // 格式化日期
    const formatDate = (dateStr: string) => {
      const date = new Date(dateStr);
      return date.toLocaleString();
    };

    // 初始加载时获取数据
    onMounted(() => {
      fetchMediaPosts();
    });

    return {
      filters,
      tableData,
      loading,
      uniqueWxids,
      formatDate,
      fetchMediaPosts,
      wxidToNameMap,
      dateShortcuts,
      totalCount,
      handleWechatSearch,
      selectLoading
    };
  },
});
</script>

<style scoped>
h2 {
  margin-bottom: 20px;
}
</style>
