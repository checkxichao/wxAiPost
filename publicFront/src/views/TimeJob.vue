<template>
  <div>

    <!-- 任务类型选择 -->
    <el-radio-group v-model="form.taskType" @change="onTaskTypeChange">
      <el-radio label="publish">发布任务</el-radio>
      <el-radio label="draft">草稿任务</el-radio>
    </el-radio-group>

    <!-- 批量操作按钮，仅发布任务时显示 -->
    <div v-if="form.taskType === 'publish'">
      <el-button
          type="primary"
          :disabled="selectedAccounts.length === 0"
          @click="batchPublishTask"
      >
        批量发布任务
      </el-button>
      <el-button
          type="success"
          :disabled="selectedAccounts.length === 0"
          @click="submitBatchScheduledTask"
      >
        批量设置定时任务
      </el-button>
    </div>

    <!-- 表格显示当前页的数据 -->
    <el-table
        ref="wechatTable"
        :data="paginatedWechatAccounts"
        style="width: 100%"
        stripe
        @selection-change="handleAccountSelectionChange"
        :default-expand-all="false"
        :row-key="getRowKey"
        v-loading="loading"

    >
      <!-- Expandable Column -->
      <el-table-column type="expand">
        <template #default="props">
          <div v-if="tasksByWxid[props.row.wxid] && tasksByWxid[props.row.wxid].length > 0">
            <el-table :data="tasksByWxid[props.row.wxid]" style="width: 100%">
              <el-table-column prop="id" label="任务ID" width="100"></el-table-column>
              <el-table-column prop="scheduledTime" label="定时发送时间" width="200">
                <template #default="scope">
                  {{ formatBeijingTime(scope.row.scheduledTime) }}
                </template>
              </el-table-column>
              <el-table-column prop="taskType" label="任务类型" width="100">
                <template #default="scope">
                  {{ translateTaskType(scope.row.taskType) }}
                </template>
              </el-table-column>
              <el-table-column prop="mode" label="模式" width="100"></el-table-column>
              <el-table-column prop="articleCount" label="文章数量" width="100"></el-table-column>
              <!-- 仅草稿任务显示 thumbId 和 templateId -->
              <el-table-column prop="thumbId" label="缩略图ID" width="200">
                <template #default="scope">
                  <span v-if="scope.row.taskType === 'draft'">{{ scope.row.thumbId }}</span>
                </template>
              </el-table-column>

              <el-table-column prop="templateId" label="模板ID" width="200"
              >
                <template #default="scope">
                  <span v-if="scope.row.taskType === 'draft'">{{ scope.row.templateId }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="status" label="状态" width="200">
                <template #default="scope">
                  <span>{{ translateStatus(scope.row.status) }}</span>
                  <el-button type="link" size="small" @click="deleteTask(scope.row.id)">删除任务</el-button>
                </template>
              </el-table-column>
            </el-table>

            <div style="margin-top: 10px;">
              <el-button
                  type="warning"
                  @click="closeSubTable(props.row)"
              >
                关闭此表格
              </el-button>
            </div>
          </div>
          <div v-else>
            <el-empty description="暂无任务"></el-empty>
          </div>
        </template>
      </el-table-column>

      <!-- 选择列 -->
      <el-table-column type="selection" width="55"></el-table-column>

      <!-- 其他列 -->
      <el-table-column prop="id" label="ID" width="80"></el-table-column>
      <el-table-column prop="name" label="公众号名" width="150"></el-table-column>
      <el-table-column prop="wxid" label="微信openid" width="180"></el-table-column>
      <el-table-column prop="secret" label="密钥" width="300"></el-table-column>
      <el-table-column prop="isuse" label="是否使用" width="80"></el-table-column>
      <el-table-column prop="bindWechat" label="绑定者" width="220"></el-table-column>
      <el-table-column prop="created_at" :formatter="formatDate" label="时间" width="200"></el-table-column>
      <el-table-column :label="'状态'" width="80" :formatter="statusFormatter"></el-table-column>
      <el-table-column label="操作" width="200">
        <template #default="scope">
          <el-button type="danger" size="small" @click="getTaskCount(scope.row.wxid)">获取任务</el-button>
          <el-button
              type="primary"
              size="small"
              @click="publishTask(scope.row.wxid, scope.row.secret)"
          >
            发布任务
          </el-button>
          <el-button
              type="primary"
              size="small"
              @click="openWriteDart(scope.row.wxid, scope.row.secret)"
              v-if="taskCount && taskCount.wxid === scope.row.wxid && scope.row.nowState !== 2"
          >
            草稿任务
          </el-button>
          <el-button
              type="warning"
              size="small"
              @click="stopTask(scope.row.wxid)"
              v-if="taskCount && taskCount.wxid === scope.row.wxid"
          >
            停止任务
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页组件 -->
    <el-pagination
        background
        layout="prev, pager, next, jumper"
        :current-page.sync="currentPage"
        :page-size="pageSize"
        :total="total"
        @current-change="handlePageChange"
        style="margin-top: 20px; text-align: right;"
    ></el-pagination>

    <!-- 底部的定时任务设置部分 -->
    <div class="scheduled-task-section">
      <h3>设置定时任务</h3>

      <!-- 显示选中的账号 -->
      <el-card class="box-card" style="margin-bottom: 20px;">
        <div slot="header" class="clearfix">
          <span>选中的账号</span>
        </div>
        <div>
          <el-tag
              v-for="account in selectedAccounts"
              :key="account.wxid"
              type="info"
              closable
              disable-transitions
              @close="removeAccount(account)"
          >
            {{ account.name }}
          </el-tag>
        </div>
      </el-card>

      <el-form :model="form" label-width="120px">
        <!-- 发布任务时的时间范围和任务数量 -->
        <el-form-item label="定时发送时间" required v-if="form.taskType === 'publish'">
          <el-date-picker
              v-model="form.timeRange"
              type="datetimerange"
              range-separator="至"
              start-placeholder="开始时间"
              end-placeholder="结束时间"
              style="width: 100%;"
              :disabled="selectedAccounts.length === 0"
          />
        </el-form-item>

        <el-form-item label="任务数量" required v-if="form.taskType === 'publish'">
          <el-input-number
              v-model="form.taskCount"
              :min="1"
              style="width: 100%;"
              :disabled="selectedAccounts.length === 0"
          />
        </el-form-item>

        <!-- 草稿任务时的单个时间选择 -->
        <el-form-item label="定时发送时间" required v-if="form.taskType === 'draft'">
          <el-date-picker
              v-model="form.scheduledTime"
              type="datetime"
              placeholder="请选择定时发送时间"
              style="width: 100%;"
              :disabled="selectedAccounts.length === 0"
          />
        </el-form-item>

        <el-form-item label="任务类型" required>
          <el-radio-group v-model="form.taskType" :disabled="false" @change="onTaskTypeChange">
            <el-radio label="draft">草稿任务</el-radio>
            <el-radio label="publish">发布任务</el-radio>
          </el-radio-group>
        </el-form-item>
        <div v-if="form.taskType === 'draft' && form.mode === 'manual'" style="margin-top: 20px;">
          <h4>选择标题</h4>
          <div style="margin-bottom: 10px; display: flex; align-items: center;">
            <el-input
                v-model="titleSearchQuery"
                placeholder="搜索标题"
                clearable
                style="width: 300px; margin-right: 10px;">
              <template #append>
                <el-button icon="el-icon-search" @click="handleTitleSearch">搜索</el-button>
              </template>
            </el-input>
            <el-button type="primary" @click="fetchManualTitles" :disabled="loadingManualTitles">
              刷新
            </el-button>
          </div>
          <el-table
              :data="paginatedManualTitles"
              style="width: 100%"
              stripe
              border
              @selection-change="handleTitleSelectionChange"
              v-loading="loadingManualTitles"
          >
            <el-table-column type="selection" width="55"></el-table-column>
            <el-table-column prop="id" label="ID" width="80"></el-table-column>
            <el-table-column prop="title" label="标题" width="300"></el-table-column>
            <el-table-column prop="createdAt" :formatter="formatDate" label="创建时间" width="180"></el-table-column>
          </el-table>
          <!-- 分页组件 for Manual Titles -->
          <el-pagination
              background
              layout="prev, pager, next, jumper"
              :current-page.sync="manualCurrentPage"
              :page-size="manualPageSize"
              :total="manualTotal"
              @current-change="handleManualPageChange"
              style="margin-top: 10px; text-align: right;"
          ></el-pagination>

        </div>
        <el-form-item label="模式" required>
          <el-radio-group v-model="form.mode" :disabled="selectedAccounts.length === 0">
            <el-radio label="random">随机</el-radio>
            <el-radio label="sequential">顺序</el-radio>
            <el-radio label="manual">手动</el-radio>
          </el-radio-group>
        </el-form-item>

        <!-- 仅在任务类型为 publish && 模式为 manual 时显示搜索框和表格 -->
        <div v-if="form.taskType === 'publish' && form.mode === 'manual'">
          <!-- 搜索框 -->
          <div style="margin-bottom: 10px; display: flex; align-items: center;">
            <el-input
                v-model="mediaSearchQuery"
                placeholder="搜索 Media ID 或标题"
                clearable
                style="width: 300px; margin-right: 10px;"
            >
              <template #append>
                <el-button icon="el-icon-search" @click="handleSearch">
                  搜索
                </el-button>
              </template>
            </el-input>
          </div>

          <!-- “清理无效草稿”按钮 -->
          <div style="margin-top: 10px; text-align: left;">
            <el-button
                type="warning"
                @click="cleanInvalidDrafts(currentPublishWxid, currentSecret)"
                :disabled="paginatedMediaList.length === 0"
            >
              清理无效草稿
            </el-button>
          </div>

          <!-- 表格 -->
          <el-table
              :data="paginatedMediaList"
              style="width: 100%"
              stripe
              border
              @selection-change="handleMediaSelectionChange"
              v-loading="loadingPublishMedia"
          >
            <el-table-column type="selection" width="30"></el-table-column>
            <el-table-column prop="id" label="ID" width="70"></el-table-column>
            <el-table-column prop="mediaId" label="Media ID" width="580"></el-table-column>
            <el-table-column prop="title" label="标题" width="580"></el-table-column>
            <el-table-column
                prop="createdAt"
                :formatter="formatDate"
                label="创建时间"
                width="140"
            ></el-table-column>
            <!-- 预览草稿列 -->
            <el-table-column label="预览草稿" width="170">
              <template #default="scope">
                <el-button-group>
                  <!-- 获取预览按钮：如果没有预览链接显示 -->
                  <el-button
                      type="primary"
                      size="small"
                      v-if="!previewLinks[scope.row.mediaId]"
                      @click="getPreviewLink(scope.row)"
                  >
                    获取预览
                  </el-button>

                  <!-- 删除草稿按钮：始终显示 -->
                  <el-button
                      type="danger"
                      size="small"
                      @click="deleteMediaId(scope.row.mediaId)"
                  >
                    删除草稿
                  </el-button>

                  <!-- 打开预览按钮：只有有预览链接时才显示 -->
                  <el-button
                      type="success"
                      size="small"
                      v-if="previewLinks[scope.row.mediaId]"
                      @click="openPreview(previewLinks[scope.row.mediaId])"
                  >
                    打开预览
                  </el-button>

                  <!-- 复制链接按钮：只有有预览链接时才显示 -->
                  <el-button
                      type="info"
                      size="small"
                      v-if="previewLinks[scope.row.mediaId]"
                      @click="copyPreviewLink(previewLinks[scope.row.mediaId])"
                  >
                    复制链接
                  </el-button>
                </el-button-group>
              </template>
            </el-table-column>
          </el-table>

          <!-- 分页组件 -->
          <div
              style="margin-top: 10px; display: flex; justify-content: center; align-items: center;"
          >
            <el-pagination
                @size-change="handlePageSizeChange"
                @current-change="handlePageChange"
                :current-page="currentPage"
                :page-sizes="[10, 20, 30, 50]"
                :page-size="pageSize"
                layout="total, sizes, prev, pager, next, jumper"
                :total="publishMediaTotal"
                background
            >
            </el-pagination>
          </div>
        </div>

        <el-form-item label="发送文章数量" required>
          <el-input-number
              v-model="form.articleCount"
              :min="1"
              style="width: 100%;"
              :disabled="selectedAccounts.length === 0"
          />
        </el-form-item>

        <el-form-item label="选择缩略图" required v-if="form.taskType === 'draft'">
          <el-select v-model="form.thumbId" placeholder="请选择缩略图" :disabled="selectedAccounts.length === 0">
            <el-option
                v-for="media in materials"
                :key="media.thumbMediaId"
                :label="media.note || '无备注'"
                :value="media.thumbMediaId"
            >
              <template #default>
                <div style="display: flex; align-items: center; justify-content: space-between; width: 100%;">
                  <div style="display: flex; align-items: center;">
                    <img :src="media.imgUrl" alt="缩略图"
                         style="width: 50px; height: 50px; object-fit: cover; margin-right: 10px;"/>
                    <div>
                      <div>{{ media.thumbMediaId }}</div>
                      <div>{{ media.note || '无备注' }}</div>
                    </div>
                  </div>
                  <el-button size="mini" type="primary" @click.stop="copyUrl(media.imgUrl)">
                    复制URL
                  </el-button>
                </div>
              </template>
            </el-option>
          </el-select>
        </el-form-item>

        <el-form-item label="选择模板" required v-if="form.taskType === 'draft'">
          <el-select
              v-model="form.templateId"
              placeholder="请选择模板"
              :disabled="selectedAccounts.length === 0"
          >
            <el-option
                v-for="template in templates"
                :key="template"
                :label="template"
                :value="template"
            ></el-option>
          </el-select>
        </el-form-item>

        <el-button
            type="primary"
            @click="submitBatchScheduledTask"
            :disabled="selectedAccounts.length === 0"
        >
          提交定时任务
        </el-button>
      </el-form>
    </div>
  </div>
</template>

<script lang="ts">
import {computed, defineComponent, onMounted, reactive, ref} from 'vue';
import http from '../services/http';
import {ElMessage, ElMessageBox, ElTable} from 'element-plus';
import copy from 'copy-to-clipboard';
import dayjs from 'dayjs';

interface MediaPost {
  id: number;
  mediaId: string;
  belongWxid: string;
  state: boolean;
  createdAt: string;
  updatedAt: string;
}
interface PublishMediaResponse {
  code: number;
  message: string;
  data: PublishMediaData;
}
interface PublishMediaData {
  data: MediaItem[];
  total: number;
}
interface DailyPostCount {
  date: string;
  belongWxid: string;
  count: number;
}

interface WechatAccount {
  id: number;
  name: string;
  wxid: string;
  secret: string;
  bindWechat: string;
  isuse: boolean;
  created_at: string;
  nowState: number;
}

interface TaskCount {
  wxid: string;
  count: number;
}

interface MediaItem {
  id: number;
  mediaId: string;
  belongWxid: string;
  state: boolean;
  created_at: string;
  updated_at: string;
}

interface Material {
  id: number;
  thumbMediaId: string;
  wxid: string;
  note: string;
  imgUrl: string;
  created_at: string;
}

interface TitleListResponse {
  code: number;
  message: string;
  data: {
    data: TitleItem[]; // 标题数组
    total: number; // 总记录数
  };
}

interface TitleItem {
  id: number;
  title: string;
}

interface Task extends CreateTask {
  id: number;
  status: string;
}

interface CreateTask {
  wxid: string;
  secret: string;
  scheduledTime: string;
  taskType: "publish" | "draft";
  mode: "random" | "sequential" | "manual"; // 小写
  articleCount: number;
  thumbId?: string;
  templateId?: string;
  selectedTitles?: string[]; // 修改为 string[]
  selectedArticles: string[]; // 根据实际需求调整类型
}


export default defineComponent({
  name: 'TimeJob',
  setup() {
    const wechatTable = ref<InstanceType<typeof ElTable> | null>(null);

    const filters = ref({
      belongWxid: '',
      dateRange: [],
    });

    const publishLog = ref<string>('');
    const tableData = ref<DailyPostCount[]>([]); // 存储统计结果
    const loading = ref(false);
    const uniqueWxids = ref<string[]>([]); // 存储所有唯一的 wxid
    const wxidToNameMap = ref<Record<string, string>>({}); // 存储 wxid 到 name 的映射
    const wechatAccounts = ref<WechatAccount[]>([]);
    const taskCount = ref<TaskCount | null>(null);
    const currentPublishWxid = ref<string>(''); // 当前发布任务的 wxid
    const currentSecret = ref<string>(''); // 当前发布任务的 secret
    const publishMediaList = ref<MediaItem[]>([]);
    const isPublishListVisible = ref(false);
    const selectedMediaIds = ref<string[]>([]);
    const previewLinks = reactive<Record<string, string>>({}); // 存储预览链接，key 为 mediaId
    const isWriteDartVisible = ref(false);
    const getRowKey = (row: WechatAccount) => row.wxid;

    // 批量定时任务相关
    const form = reactive({
      scheduledTime: '', // 定时任务的发送时间（草稿任务）
      timeRange: ['', ''] as [string, string], // 定时发送时间区间（发布任务）

      taskCount: 1, // 发布任务的数量
      wxid: '',
      secret: '',
      titleSelectionMethod: 'random' as 'random' | 'sequential' | 'manual', // 标题选择方法
      taskType: 'publish' as 'draft' | 'publish', // 任务类型
      mode: 'random' as 'random' | 'sequential' | 'manual', // 模式
      articleCount: 1, // 发送文章数量
      thumbId: '',
      templateId: '',
      selectedTitles: [] as string[], // 新增：存储选中的标题ID
      selectedArticles:[] as  string[],
    });

    const materials = ref<Material[]>([]);
    const templates = ref<string[]>([]); // 存储模板列表
    const titles = ref<TitleItem[]>([]); // 存储标题列表
    const manualTitles = ref<TitleItem[]>([]); // 手动选择模式下的标题列表
    const manualTotal = ref<number>(0); // 总记录数
    const manualCurrentPage = ref<number>(1); // 当前页码
    const manualPageSize = ref<number>(10); // 每页显示数量，适当增加以减少请求次数
    const wechatTotal = ref<number>(0); // 新增总记录数
    const wechatCurrentPage = ref<number>(1); // 当前页码
    const wechatPageSize = ref<number>(10); // 每页显示数量



    const tasksByWxid = ref<Record<string, Task[]>>({}); // 存储每个 wxid 的任务列表

    // 分页相关的响应式变量
    const currentPage = ref<number>(1); // 当前页码
    const pageSize = ref<number>(10); // 每页显示条数
    const total = computed(() => wechatAccounts.value.length); // 总记录数

    // 计算当前页需要显示的数据
    const paginatedWechatAccounts = computed(() => {
      const start = (currentPage.value - 1) * pageSize.value;
      const end = start + pageSize.value;
      return wechatAccounts.value.slice(start, end);
    });

    // 新增：计算手动模式下当前页需要显示的标题
    const paginatedManualTitles = computed(() => manualTitles.value);
    const selectedTitles = ref<TitleItem[]>([]);
    const handleTitleSelectionChange = (selection: TitleItem[]) => {
      selectedTitles.value = selection;
    };



    // 新增：定义 loadingManualTitles
    const loadingManualTitles = ref<boolean>(false);
    const publishMediaTotal = ref<number>(0); // 总记录数
    // 处理页码变化
    const handlePageChange = (page: number) => {
      currentPage.value = page;
      // 清除表格选择状态
      if (wechatTable.value) {
        wechatTable.value.clearSelection();
      }
    };

    const onTaskTypeChange = (newType: string) => {
      console.log('Task type changed to:', newType);
      if (newType !== 'draft') {
        form.thumbId = '';
        form.templateId = '';
        form.selectedTitles = []; // 清空选中的标题
      } else {
        // 清空已选择的多个账号
        if (selectedAccounts.value.length > 1) {
          ElMessage.warning('草稿任务只能选择一个账号');
          const first = selectedAccounts.value[0];
          selectedAccounts.value = [first];
          if (wechatTable.value) {
            wechatTable.value.clearSelection();
            wechatTable.value.toggleRowSelection(first, true);
          }
        }
        // 如果只选择了一个账号，加载缩略图
        if (selectedAccounts.value.length === 1) {
          fetchMaterials(selectedAccounts.value[0].wxid);
        }
      }
    };
    const selectedAccounts = ref<WechatAccount[]>([]);
    const handleAccountSelectionChange = (selection: WechatAccount[]) => {
      // 这段逻辑只处理公众号的选择
      // 1) 先移除当前页取消选中的项
      const currentPageWxids = paginatedWechatAccounts.value.map(acc => acc.wxid);
      selectedAccounts.value = selectedAccounts.value.filter(
          acc => !currentPageWxids.includes(acc.wxid)
      );

      // 2) 再添加当前页勾选的项
      selectedAccounts.value = [...selectedAccounts.value, ...selection];

      // 3) 去重
      selectedAccounts.value = Array.from(
          new Map(selectedAccounts.value.map(acc => [acc.wxid, acc])).values()
      );
    };

    // 选择变化处理
    const handleSelectionChange = (selection: WechatAccount[]) => {
      console.log('Selection changed:', selection);

      // 1) 先移除当前页取消选中的项
      const currentPageWxids = paginatedWechatAccounts.value.map(acc => acc.wxid);
      selectedAccounts.value = selectedAccounts.value.filter(
          acc => !currentPageWxids.includes(acc.wxid)
      );

      // 2) 再添加当前页勾选的项
      selectedAccounts.value = [...selectedAccounts.value, ...selection];

      // 3) 去重
      selectedAccounts.value = Array.from(
          new Map(selectedAccounts.value.map(acc => [acc.wxid, acc])).values()
      );

      console.log('Selected accounts:', selectedAccounts.value);

      // ===== 核心：判断是否只勾选了1个账号（单选）；如果允许多选，要自己定义逻辑 =====
      if (selection.length === 1) {
        // 判断是否从另一个账号切换到这个账号
        if (currentPublishWxid.value && currentPublishWxid.value !== selection[0].wxid) {
          // 说明用户勾选了一个新的账号，和之前不一样 => 做“切换”处理

          // 例如：清空上一个账号的媒体列表，防止混淆
          publishMediaList.value = [];
          // 清空上一个账号的预览链接
          for (const key in previewLinks) {
            delete previewLinks[key];
          }
          // 如果要马上加载新账号的媒体列表：
          // fetchPublishMediaList();
          // 或者 fetchMaterials(selection[0].wxid);
          // 根据你的需求选择
        }

        // 然后更新当前账号信息
        currentPublishWxid.value = selection[0].wxid;
        currentSecret.value = selection[0].secret;
      } else {
        // 如果多选或都没勾选，就清空，防止后续操作时出错
        currentPublishWxid.value = '';
        currentSecret.value = '';
      }

      // 如果是草稿模式，并且只勾选了一个账号，就加载它的缩略图
      if (form.taskType === 'draft' && selectedAccounts.value.length === 1) {
        fetchMaterials(selectedAccounts.value[0].wxid);
      }
    };



    // 格式化时间为北京时间
    const formatBeijingTime = (timeStr: string) => {
      const date = new Date(timeStr);
      const options: Intl.DateTimeFormatOptions = {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        hour12: false,
        timeZone: 'Asia/Shanghai',
      };
      return new Intl.DateTimeFormat('zh-CN', options).format(date);
    };

    // 移除选中的账号
    const removeAccount = (account: WechatAccount) => {
      selectedAccounts.value = selectedAccounts.value.filter(acc => acc.wxid !== account.wxid);
      if (wechatTable.value) {
        wechatTable.value.toggleRowSelection(account, false);
      }
    };

    // 批量发布任务
    const batchPublishTask = () => {

      selectedAccounts.value.forEach(account => {
        publishTask(account.wxid, account.secret);
      });
      ElMessage.success(`已提交 ${selectedAccounts.value.length} 个公众号的发布任务`);
    };

    // 批量提交定时任务
    const submitBatchScheduledTask = async () => {
      if (form.taskType === 'publish') {
        // 发布任务的逻辑
        if (!form.timeRange || form.timeRange.length !== 2) {
          ElMessage.error('请选择定时发送时间区间');
          return;
        }

        const [startTime, endTime] = form.timeRange;
        if (dayjs(endTime).isBefore(dayjs(startTime))) {
          ElMessage.error('结束时间必须在开始时间之后');
          return;
        }

        const taskCountNumber = form.taskCount;
        if (taskCountNumber < 1) {
          ElMessage.error('任务数量至少为1');
          return;
        }

        const totalMinutes = dayjs(endTime).diff(dayjs(startTime), 'minute');
        const minTotalMinutes = (taskCountNumber - 1) * 3; // 每个任务间隔至少3分钟
        if (totalMinutes < minTotalMinutes) {
          ElMessage.error(`时间区间过短，无法安排 ${taskCountNumber} 个任务，每个任务至少间隔3分钟`);
          return;
        }

        // 生成随机任务时间，确保每个任务至少间隔3分钟
        const scheduledTimes = generateRandomScheduledTimes(startTime, endTime, taskCountNumber, 3);

        if (scheduledTimes.length < taskCountNumber) {
          ElMessage.error('无法生成满足间隔要求的任务时间，请调整时间区间或任务数量');
          return;
        }

        // 构建任务数组，使用 mapMode 转换 mode 值
        const tasks: CreateTask[] = selectedAccounts.value.flatMap(account => {
          return scheduledTimes.map(time => ({
            wxid: account.wxid,
            secret: account.secret,
            scheduledTime: time,
            taskType: form.taskType,
            mode: mapMode(form.mode), // 使用 mapMode 转换
            articleCount: form.articleCount,
            selectedArticles: form.selectedArticles, // 根据需要调整
            // 仅在任务类型为 'draft' 时添加 thumbId 和 templateId
            ...(form.taskType === 'draft' && {
              thumbId: form.thumbId,
              templateId: form.templateId,
              ...(form.mode === 'manual' && {selectedTitles: form.selectedTitles}),
            }),
          }));
        });


        console.log('Tasks to be sent:', tasks); // 调试日志

        try {
          const response = await http.post('/send/batchScheduleTask', {tasks});

          if (response.data.code === 200) {
            ElMessage.success('批量定时任务已成功创建');
            // 重置表单
            form.timeRange = ['', ''];
            form.taskCount = 1;
            form.mode = 'random';
            form.articleCount = 1;

            selectedAccounts.value = [];
            materials.value = []; // 清空 materials
            if (wechatTable.value) {
              wechatTable.value.clearSelection();
            }
          } else {
            ElMessage.error('创建批量定时任务失败: ' + response.data.message);
          }
        } catch (error: any) {
          // 详细输出错误信息
          if (error.response && error.response.data && error.response.data.message) {
            ElMessage.error(`创建批量定时任务失败: ${error.response.data.message}`);
          } else if (error.message) {
            ElMessage.error(`创建批量定时任务失败: ${error.message}`);
          } else {
            ElMessage.error('创建批量定时任务失败: 未知错误');
          }
          console.error('Error:', error);
        }
      } else {
        // 草稿任务的逻辑
        if (!form.scheduledTime) {
          ElMessage.error('请选择定时发送时间');
          return;
        }

        if (selectedAccounts.value.length === 0) {
          ElMessage.error('请先选择至少一个账号');
          return;
        }

        if (form.taskType === 'draft') {
          if (!form.thumbId) {
            ElMessage.error('请选择一个缩略图');
            return;
          }
          if (!form.templateId) {
            ElMessage.error('请选择一个模板');
            return;
          }
          if (selectedAccounts.value.length > 1) {
            ElMessage.error('草稿任务只能选择一个账号');
            return;
          }
          // 如果模式为手动，且没有选中标题
          if (form.mode === 'manual' && form.selectedTitles.length === 0) {
            ElMessage.error('请至少选择一个标题');
            return;
          }
        }

        // 构建任务数组，使用 mapMode 转换 mode 值
        const tasks: CreateTask[] = selectedAccounts.value.map(account => {
          const task: CreateTask = {
            wxid: account.wxid,
            secret: account.secret,
            scheduledTime: form.scheduledTime,
            taskType: form.taskType,
            mode: mapMode(form.mode),
            articleCount: form.articleCount,
            selectedArticles: [], // 根据需要调整
          };

          // 仅在任务类型为 'draft' 时添加 thumbId 和 templateId
          if (form.taskType === 'draft') {
            task.thumbId = form.thumbId;
            task.templateId = form.templateId;
            // 如果模式为手动，添加选中的标题内容
            if (form.mode === 'manual') {
              task.selectedTitles = form.selectedTitles;
            }
          }

          return task;
        });

        console.log('Tasks to be sent:', tasks); // 调试日志

        try {
          const response = await http.post('/send/batchScheduleTask', {tasks});

          if (response.data.code === 200) {
            ElMessage.success('批量定时任务已成功创建');
            // 重置表单
            form.scheduledTime = '';
            form.taskType = 'publish';
            form.mode = 'random';
            form.articleCount = 1;
            form.thumbId = '';
            form.templateId = '';
            form.selectedTitles = []; // 重置选中的标题

            selectedAccounts.value = [];
            materials.value = []; // 清空 materials
            if (wechatTable.value) {
              wechatTable.value.clearSelection();
            }
          } else {
            ElMessage.error('创建批量定时任务失败: ' + response.data.message);
          }
        } catch (error: any) {
          // 详细输出错误信息
          if (error.response && error.response.data && error.response.data.message) {
            ElMessage.error(`创建批量定时任务失败: ${error.response.data.message}`);
          } else if (error.message) {
            ElMessage.error(`创建批量定时任务失败: ${error.message}`);
          } else {
            ElMessage.error('创建批量定时任务失败: 未知错误');
          }
          console.error('Error:', error);
        }
      }
    };


    // 生成随机定时发送时间，确保每个任务至少间隔minInterval分钟
    const generateRandomScheduledTimes = (
        startTime: string,
        endTime: string,
        taskCount: number,
        minInterval: number
    ): string[] => {
      const start = dayjs(startTime);
      const end = dayjs(endTime);
      const totalMinutes = end.diff(start, 'minute');
      const minTotalMinutes = (taskCount - 1) * minInterval;

      if (totalMinutes < minTotalMinutes) {
        return [];
      }

      // 随机分配任务时间，确保间隔
      const availableMinutes = totalMinutes - minTotalMinutes;
      const extraIntervals: number[] = [];

      for (let i = 0; i < taskCount; i++) {
        extraIntervals.push(0);
      }

      // 分配额外时间
      for (let i = 0; i < availableMinutes; i++) {
        const randomIndex = Math.floor(Math.random() * taskCount);
        extraIntervals[randomIndex]++;
      }

      const scheduledTimes: string[] = [];
      let currentTime = start;

      for (let i = 0; i < taskCount; i++) {
        const minutesToAdd = i === 0 ? 0 : minInterval + extraIntervals[i];
        currentTime = currentTime.add(minutesToAdd, 'minute');
        scheduledTimes.push(currentTime.toISOString());
      }

      return scheduledTimes;
    };
// 添加映射函数，将前端的 mode 值转换为后端期望的格式
    const mapMode = (mode: string): "random" | "sequential" | "manual" => {
      switch (mode.toLowerCase()) { // 确保转换为小写
        case 'random':
          return 'random';
        case 'sequential':
          return 'sequential';
        case 'manual':
          return 'manual';
        default:
          throw new Error(`未知的模式: ${mode}`);
      }
    };


    // 获取公众号列表的 API
    const fetchWechatList = async () => {
      try {
        const response = await http.get('/wechat/getWechat');
        if (response.data && response.data.info) {
          wechatAccounts.value = response.data.info;
          console.log('微信账号列表:', wechatAccounts.value); // 调试日志

          wxidToNameMap.value = response.data.info.reduce((map: Record<string, string>, item: any) => {
            map[item.wxid] = item.name;
            return map;
          }, {});
        } else {
          console.error('Unexpected response structure:', response.data);
        }
      } catch (error) {
        console.error('Error fetching Wechat list:', error);
      }
    };

    // 获取任务数并加载任务列表
    const getTaskCount = async (wxid: string) => {
      try {
        const response = await http.post('/send/getTasks', {wxid}); // 确保后端有此端点
        if (response.data.code === 200) {
          tasksByWxid.value[wxid] = response.data.data; // 存储任务
          taskCount.value = {wxid, count: response.data.data.length};

          // 找到对应的行数据并展开
          const row = wechatAccounts.value.find(account => account.wxid === wxid);
          if (row && wechatTable.value) {
            wechatTable.value.toggleRowExpansion(row, true); // 展开该行
          }
        } else {
          ElMessage.error(`获取任务失败,没有任务数`);
        }
      } catch (error) {
        ElMessage.error('获取任务数失败');
      }
    };

    const statusFormatter = (row: WechatAccount) => {
      switch (row.nowState) {
        case 0:
          return '无任务';
        case 1:
          return '任务中';
        case 2:
          return '任务中';
        default:
          return '未知状态';
      }
    };

    const fetchMediaPosts = async () => {
      loading.value = true;
      try {
        const response = await http.post('/wechat/getMediaPost', {
          belongWxid: filters.value.belongWxid || '',
          startDate: filters.value.dateRange[0] || '',
          endDate: filters.value.dateRange[1] || '',
        });
        if (response.data && response.data.data) {
          const rawData: MediaPost[] = response.data.data;
          tableData.value = processPostData(rawData);
          uniqueWxids.value = Array.from(new Set(rawData.map(item => item.belongWxid)));
          await fetchWechatList(); // 确保获取微信账号列表
        } else {
          console.error('Unexpected response structure:', response.data);
        }
      } catch (error) {
        console.error('Error fetching media posts:', error);
      } finally {
        loading.value = false;
      }
    };

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

    const translateStatus = (status: string) => {
      switch (status) {
        case 'pending':
          return '待处理';
        case 'in_progress':
          return '进行中';
        case 'completed':
          return '已完成';
        case 'failed':
          return '失败';
        default:
          return status;
      }
    };

    const dateShortcuts = [
      {
        text: '今天',
        value: () => {
          const today = new Date();
          return [
            new Date(today.setHours(0, 0, 0, 0)),
            new Date(today.setHours(23, 59, 59, 999)),
          ];
        },
      },
      {
        text: '昨日',
        value: () => {
          const yesterday = new Date();
          yesterday.setDate(yesterday.getDate() - 1);
          return [
            new Date(yesterday.setHours(0, 0, 0, 0)),
            new Date(yesterday.setHours(23, 59, 59, 999)),
          ];
        },
      },
      {
        text: '最近一周',
        value: () => {
          const end = new Date();
          const start = new Date();
          start.setDate(start.getDate() - 7);
          return [
            new Date(start.setHours(0, 0, 0, 0)),
            new Date(end.setHours(23, 59, 59, 999)),
          ];
        },
      },
      {
        text: '最近一个月',
        value: () => {
          const end = new Date();
          const start = new Date();
          start.setMonth(start.getMonth() - 1);
          return [
            new Date(start.setHours(0, 0, 0, 0)),
            new Date(end.setHours(23, 59, 59, 999)),
          ];
        },
      },
      {
        text: '最近三个月',
        value: () => {
          const end = new Date();
          const start = new Date();
          start.setMonth(start.getMonth() - 3);
          return [
            new Date(start.setHours(0, 0, 0, 0)),
            new Date(end.setHours(23, 59, 59, 999)),
          ];
        },
      },
      {
        text: '最近半年',
        value: () => {
          const end = new Date();
          const start = new Date();
          start.setMonth(start.getMonth() - 6);
          return [
            new Date(start.setHours(0, 0, 0, 0)),
            new Date(end.setHours(23, 59, 59, 999)),
          ];
        },
      },
    ];

    const totalCount = computed(() => {
      return tableData.value.reduce((total, item) => total + item.count, 0);
    });

    const formatDate = (row: any, column: any, cellValue: string) => {
      if (!cellValue) return '';
      // 使用 dayjs 格式化日期
      return dayjs(cellValue).format('YYYY-MM-DD HH:mm:ss');
    };
    const selectedMedia = ref<MediaItem[]>([]);


    const handleMediaSelectionChange = (selection: MediaItem[]) => {
      // 1) 把选中的稿件存起来
      selectedMedia.value = selection;

      // 2) 把选中的稿件ID放到 form.selectedArticles
      //   （mediaId 是字符串也没事，后端如果要数字，就得另外转一下）
      form.selectedArticles = selection.map(item => item.mediaId);

      console.log('手动模式选中的稿件ID:', form.selectedArticles);
    };

    const loadingPublishMedia = ref<boolean>(false); // 加载状态
    const fetchMaterials = async (wxid: string) => {
      loading.value = true;
      try {
        const response = await http.post('/send/getThumbList', {wxid}, {});
        console.log('响应数据:', response); // 输出响应内容，查看结构

        if (response.data && response.data.list && Array.isArray(response.data.list)) {
          materials.value = response.data.list; // 将素材列表赋值给 materials
        } else {
          ElMessage.error('获取素材失败: 无效的素材列表');
        }
      } catch (error) {
        console.error('请求失败:', error); // 输出详细错误信息
        ElMessage.error('获取素材失败');
      } finally {
        loading.value = false;
      }
    };

    const fetchTitles = async () => {
      try {
        const response = await http.get<TitleListResponse>('/send/getTitleList', {
          params: {
            page: wechatCurrentPage.value,
            pageSize: wechatPageSize.value,
          },
        });

        console.log('获取到的标题列表:', response.data); // 调试日志

        if (
            response.data.code === 200 &&
            response.data.data &&
            Array.isArray(response.data.data.data)
        ) {
          titles.value = response.data.data.data; // 更新标题数据
          wechatTotal.value = response.data.data.total; // 更新总记录数
        } else {
          ElMessage.error(`获取标题失败: ${response.data.message}`);
          titles.value = [];
          wechatTotal.value = 0;
        }
      } catch (error) {
        console.error('获取标题列表失败:', error);
        ElMessage.error('获取标题列表失败');
        titles.value = [];
        wechatTotal.value = 0;
      }
    };

    const fetchManualTitles = async () => {
      if (titleSearchQuery.value!=""){

        loadingManualTitles.value = true; // 开始加载
        try {
          const response = await http.post<TitleListResponse>('/send/getTitleListSearch', {

            page: manualCurrentPage.value, // 使用当前页码
            pageSize: manualPageSize.value,
            search: titleSearchQuery.value, // 搜索参数

          });

          console.log('获取到的标题列表:', response.data); // 调试日志

          if (
              response.data.code === 200 &&
              response.data.data &&
              Array.isArray(response.data.data.data)
          ) {
            manualTitles.value = response.data.data.data; // 设置为当前页的数据
            manualTotal.value = response.data.data.total;
          } else {
            ElMessage.error(`获取标题失败: ${response.data.message}`);
            manualTitles.value = [];
            manualTotal.value = 0;
          }
        } catch (error) {
          console.error('获取标题列表失败:', error);
          ElMessage.error('获取标题列表失败');
          manualTitles.value = [];
          manualTotal.value = 0;
        } finally {
          loadingManualTitles.value = false; // 加载完成
        }
      }else{
        loadingManualTitles.value = true; // 开始加载
        try {
          const response = await http.get<TitleListResponse>('/send/getTitleList', {
            params: {
              page: wechatCurrentPage.value,
              pageSize: wechatPageSize.value,
            },
          });

          console.log('获取到的标题列表:', response.data); // 调试日志

          if (
              response.data.code === 200 &&
              response.data.data &&
              Array.isArray(response.data.data.data)
          ) {
            manualTitles.value = response.data.data.data; // 设置为当前页的数据
            manualTotal.value = response.data.data.total;
          } else {
            ElMessage.error(`获取标题失败: ${response.data.message}`);
            manualTitles.value = [];
            manualTotal.value = 0;
          }
        } catch (error) {
          console.error('获取标题列表失败:', error);
          ElMessage.error('获取标题列表失败');
          manualTitles.value = [];
          manualTotal.value = 0;
        } finally {
          loadingManualTitles.value = false; // 加载完成
        }
      }


    };


    const openWriteDart = (wxid: string, secret: string) => {
      isWriteDartVisible.value = true;
      form.wxid = wxid;
      form.secret = secret;
      fetchMaterials(wxid); // 传递 wxid
      fetchTemplates(); // 获取模板列表
      // 手动选择模式下，加载标题列表
      if (form.mode === 'manual') {
        manualTitles.value = [];
        manualCurrentPage.value = 1;
        manualTotal.value = 0;
        fetchManualTitles();
      } else {
        fetchTitles(); // 获取标题列表
      }
    };

    const translateTaskType = (taskType: string) => {
      switch (taskType) {
        case 'draft':
          return '草稿';
        case 'publish':
          return '发布';
        default:
          return taskType;
      }
    };

    const fetchTemplates = async () => {
      try {
        const response = await http.get('/send/getTemplateList');
        if (response.data && Array.isArray(response.data.templates)) {
          templates.value = response.data.templates;
        } else {
          ElMessage.error('获取模板失败');
        }
      } catch (error) {
        console.error('获取模板失败:', error);
        ElMessage.error('获取模板失败');
      }
    };

    // 发布任务
    const publishTask = async (belongWxid: string, secret: string) => {
      // 记录当前发布任务的 wxid
      currentPublishWxid.value = belongWxid;
      currentSecret.value = secret;
      // 获取Media ID列表
      try {
        const response = await http.post('/send/getMedias', {belongWxid}); // 替换为实际的获取Media ID的接口路径
        if (response.data.code === 200 && Array.isArray(response.data.data)) {
          publishMediaList.value = response.data.data;
          isPublishListVisible.value = true;
          selectedMediaIds.value = []; // 重置选择
          // 清空之前的预览链接
          for (const mediaId of publishMediaList.value.map(item => item.mediaId)) {
            previewLinks[mediaId] = '';
          }
        } else {
          ElMessage.error(`获取Media ID失败: ${response.data.message}`);
        }
      } catch (error) {
        console.error('获取Media ID失败:', error);
        ElMessage.error('获取Media ID失败');
      }
    };
    const handlePageSizeChange = (newSize: number) => {
      pageSize.value = newSize;
      currentPage.value = 1; // 重置到第一页
    };
    const stopTask = async (wxid: string) => {
      try {
        // 发起 POST 请求到后端的 SetStateStop 端点
        const response = await http.post('/send/setStop', {wxid});

        if (response.data.code === 200) {
          // 更新发布日志
          publishLog.value += `停止任务成功！任务数: ${taskCount.value?.count || 0}\n`;
          // 显示成功消息
          ElMessage.success('任务已停止');

          // 更新对应公众号的状态为“无任务”
          const account = wechatAccounts.value.find(acc => acc.wxid === wxid);
          if (account) {
            account.nowState = 0; // 假设 0 表示“无任务”
          }

          // 更新任务数为 0
          if (taskCount.value && taskCount.value.wxid === wxid) {
            taskCount.value.count = 0;
          }
        } else {
          // 如果后端返回错误，更新发布日志并显示错误消息
          publishLog.value += `停止任务失败: ${response.data.message}\n`;
          ElMessage.error(`停止任务失败: ${response.data.message}`);
        }
      } catch (error: any) {
        // 处理请求错误
        const errorMessage = error.response?.data?.message || error.message || '未知错误';
        publishLog.value += `停止任务失败: ${errorMessage}\n`;
        ElMessage.error(`停止任务失败: ${errorMessage}`);
      }
    };
    // 获取预览链接方法
    const getPreviewLink = async (mediaItem: MediaItem) => {
      try {
        const payload = {
          wxid: currentPublishWxid.value,
          secret: currentSecret.value, // 添加 secret
          mediaId: mediaItem.mediaId,
        };
        const response = await http.post('/send/getPreviewLink', payload); // 替换为实际的获取预览链接的接口路径

        if (response.data.code === 200 && response.data.data) {
          previewLinks[mediaItem.mediaId] = response.data.data; // 直接使用 data 作为链接
          ElMessage.success('预览链接获取成功');
        } else {
          ElMessage.error(`获取预览链接失败: ${response.data.message}`);
        }
      } catch (error) {
        console.error('获取预览链接失败:', error);
        ElMessage.error('获取预览链接失败');
      }
    };

    // 打开预览链接
    const openPreview = (link: string) => {
      window.open(link, '_blank');
    };

    // 复制预览链接
    const copyPreviewLink = (link: string) => {
      navigator.clipboard.writeText(link)
          .then(() => {
            ElMessage.success('预览链接已复制到剪切板');
          })
          .catch((err) => {
            console.error('复制预览链接失败:', err);
            ElMessage.error('复制预览链接失败');
          });
    };
    // 提交定时任务（单个任务）
    const submitScheduledTask = async () => {
      if (form.taskType === 'publish') {
        // 发布任务的逻辑已经在 submitBatchScheduledTask 中处理
        // 这里可以保留或移除此方法，视具体需求而定
      } else {
        // 草稿任务的逻辑
        try {
          const payload: any = {
            wxid: form.wxid,
            secret: form.secret,
            scheduledTime: form.scheduledTime, // 设置的定时任务时间
            articleCount: form.articleCount, // 需要发送的文章数量
            taskType: form.taskType, // 添加任务类型
            mode: form.mode, // 添加模式
            selectedArticles: [], // 添加 selectedArticles 为空数组
          };

          // 仅在任务类型为 'draft' 时添加 thumbId 和 templateId
          if (form.taskType === 'draft') {
            payload.thumbId = form.thumbId;
            payload.templateId = form.templateId;
            // 如果模式为手动，添加选中的标题ID
            if (form.mode === 'manual') {
              payload.selectedTitles = form.selectedTitles;
            }
          }

          const response = await http.post('/send/scheduleTask', payload);

          if (response.data.code === 200) {
            ElMessage.success('定时任务已成功创建');
            isWriteDartVisible.value = false;
          } else {
            ElMessage.error('创建定时任务失败: ' + response.data.message);
          }
        } catch (error) {
          ElMessage.error('创建定时任务失败');
          console.error('Error:', error);
        }
      }
    };

    const copyUrl = (url: string) => {
      if (copy(url)) {
        ElMessage.success('URL 已复制到剪贴板');
      } else {
        ElMessage.error('复制失败');
      }
    };

    // 删除任务
    const deleteTask = async (taskId: number) => {
      try {
        const response = await http.post('/send/deleteScheduledTask', {taskID: taskId});
        if (response.data.code === 200) {
          ElMessage.success('任务已删除');
          // 移除前端的任务列表
          for (const wxid in tasksByWxid.value) {
            tasksByWxid.value[wxid] = tasksByWxid.value[wxid].filter(task => task.id !== taskId);
          }
        } else {
          ElMessage.error(`删除任务失败: ${response.data.message}`);
        }
      } catch (error) {
        ElMessage.error('删除任务失败');
        console.error('Error:', error);
      }
    };

    // 初始加载时获取数据
    onMounted(async () => {
      await fetchWechatList(); // 先获取公众号列表
      await fetchTemplates(); // 获取模板列表
      fetchMediaPosts(); // 然后获取媒体数据
    });

    const mediaPosts = ref<MediaPost[]>([]); // 定义一个响应式变量来存储媒体数据

    const getMediaPost = async (Wxid: string) => {
      try {
        const response = await http.post('/send/getThumbList', {Wxid});
        if (response.data.code === 200 && Array.isArray(response.data.list)) {
          mediaPosts.value = response.data.list; // 赋值给 mediaPosts
          console.log(`MediaPost for wxid ${Wxid}:`, mediaPosts.value);
        } else {
          ElMessage.error(`获取 MediaPost 失败: ${response.data.message || '未知错误'}`);
        }
      } catch (error) {
        ElMessage.error('获取 MediaPost 失败');
        console.error('Error:', error);
      }
    };

    // 新增：处理标题搜索查询
    const titleSearchQuery = ref<string>('');
    const deleteMediaId = async (mediaId: string) => {
      // 确保有选中的项
      if (!mediaId) {
        ElMessage.error('无效的Media ID');
        return;
      }

      const payload = {
        wxid: currentPublishWxid.value,
        secret: currentSecret.value,
        mediaId: mediaId,  // 确保通过 selectedMediaIds 获取到正确的 mediaId
      };

      try {
        const response = await http.post('/send/deleteMedia', payload);

        if (response.data.code === 200 && response.data.data) {
          ElMessage.success('该链接已删除');

          // 更新发布媒体列表
          publishMediaList.value = publishMediaList.value.filter(item => item.mediaId !== mediaId);
          // 移除预览链接
          delete previewLinks[mediaId];
        } else {
          ElMessage.error(`链接删除失败: ${response.data.message}`);
        }
      } catch (error) {
        console.error('删除草稿失败:', error);
        ElMessage.error('删除草稿失败');
      }
    };
    const handleTitleSearch = () => {
      // 清空当前手动标题列表并重置分页
      manualTitles.value = [];
      manualCurrentPage.value = 1;
      manualTotal.value = 0;
      fetchManualTitles();
    };

    // 新增：处理手动选择变化
    const handleManualSelectionChange = (selection: TitleItem[]) => {
      form.selectedTitles = selection.map(item => item.title); // 使用标题内容
      console.log('Selected titles:', form.selectedTitles);
    };
    const mediaSearchQuery = ref<string>(''); // 搜索查询

    // 新增：处理手动分页变化
    const handleManualPageChange = (page: number) => {
      manualCurrentPage.value = page;
      fetchManualTitles();
    };
    const fetchPublishMediaList = async () => {
      if (!currentPublishWxid.value) {
        ElMessage.error('当前没有选中的微信账号');
        return;
      }

      loadingPublishMedia.value = true;

      try {
        const response = await http.post<PublishMediaResponse>('/send/getMediaListByWxidP', {
          wxid: currentPublishWxid.value,
          q: mediaSearchQuery.value,
        });

        console.log('获取Media ID响应:', response.data);

        if (response.data.code === 200 && Array.isArray(response.data.data.data)) {
          publishMediaList.value = response.data.data.data;
          publishMediaTotal.value = response.data.data.total || 0;
          // 清空之前的预览链接
          publishMediaList.value.forEach(item => {
            previewLinks[item.mediaId] = '';
          });
          // 重置分页
          currentPage.value = 1;
          // 显示发布列表
          isPublishListVisible.value = true;
          selectedMediaIds.value = []; // 重置选择
        } else {
          ElMessage.error(`获取Media ID失败: ${response.data.message}`);
        }

      } catch (error: any) {
        console.error('获取Media ID失败:', error);
        const errorMessage = error.response?.data?.message || error.message || '未知错误';
        ElMessage.error(`获取Media ID失败: ${errorMessage}`);
      } finally {
        loadingPublishMedia.value = false;
      }
    };
    const handleSearch = () => {
      publishMediaCurrentPage.value = 1; // 重置到第一页
      fetchPublishMediaList();
    };
    const publishMediaCurrentPage = ref<number>(1); // 当前页码
    const paginatedMediaList = computed(() => {
      const start = (currentPage.value - 1) * pageSize.value;
      const end = start + pageSize.value;
      return publishMediaList.value.slice(start, end);
    });
    const cleanInvalidDrafts = async (wxid: string, secret: string) => {
      // 确认操作
      try {
        await ElMessageBox.confirm(
            '确认要清理所有无效的草稿吗？此操作无法撤销。',
            '提示',
            {
              confirmButtonText: '确定',
              cancelButtonText: '取消',
              type: 'warning',
            }
        );
      } catch {
        // 用户取消操作
        return;
      }

      // 收集所有 mediaId
      const mediaIds = publishMediaList.value.map(item => item.mediaId);

      if (mediaIds.length === 0) {
        ElMessage.warning('当前没有可清理的草稿。');
        return;
      }

      // 构建请求负载
      const payload = {
        wxid: wxid,
        secret: secret,
        mediaIds: mediaIds,
      };

      try {
        const response = await http.post('/send/cleanInvalidDrafts', payload, {
          timeout: 0,
        });

        if (response.data.code === 200) {
          // 假设所有 mediaIds 都已被清理
          publishMediaList.value = publishMediaList.value.filter(
              item => !mediaIds.includes(item.mediaId)
          );

          // 移除对应的预览链接
          mediaIds.forEach(mediaId => {
            delete previewLinks[mediaId];
          });

          ElMessage.success(`已清理无效草稿。`);
          window.location.reload();
        } else {
          ElMessage.error(`清理无效草稿失败: ${response.data.message}`);
        }
      } catch (error: any) {
        console.error('清理无效草稿失败:', error);
        const errorMessage = error.response?.data?.message || error.message || '未知错误';
        ElMessage.error(`清理无效草稿失败: ${errorMessage}`);
      }
    };
    const closeSubTable = (row: WechatAccount) => {
      if (wechatTable.value) {
        // 折叠这一行
        wechatTable.value.toggleRowExpansion(row, false);
      }
      // 如果你想把 tasksByWxid[row.wxid] 清空，防止再次展开还是旧数据
      // tasksByWxid.value[row.wxid] = [];
      ElMessage.info(`已关闭 ${row.name} 的任务列表`);
    };
    return {
      closeSubTable,
      cleanInvalidDrafts,
      wechatTable,
      formatBeijingTime,
      filters,
      tableData,
      mediaPosts,
      loading,
      uniqueWxids,
      formatDate,
      getRowKey,
      fetchMediaPosts,
      wxidToNameMap,
      dateShortcuts,
      totalCount,
      paginatedWechatAccounts,
      statusFormatter,
      getTaskCount,
      publishTask,
      taskCount,
      openWriteDart, // 确保只定义一次
      stopTask,
      form,
      publishMediaList,
      submitScheduledTask,
      wechatAccounts,
      onTaskTypeChange,
      // 批量定时任务
      selectedAccounts,
      handleSelectionChange,
      removeAccount,
      batchPublishTask,
      submitBatchScheduledTask,
      materials,
      templates,
      tasksByWxid,
      translateTaskType,
      translateStatus,
      deleteTask,
      handleAccountSelectionChange,
      getMediaPost, // 新增
      copyUrl,
      handleSearch,
      // 分页相关
      currentPage,
      pageSize,
      total,
      handlePageChange,
      mapMode,
      // 新增：标题搜索相关
      titleSearchQuery,
      handleTitleSearch,
      manualTitles,
      manualTotal,
      manualCurrentPage,
      manualPageSize,
      paginatedManualTitles, // 新增计算属性
      handleManualSelectionChange,
      handleManualPageChange,
      loadingManualTitles, // 新增加载状态
      fetchManualTitles,
      handleTitleSelectionChange,
      mediaSearchQuery,
      currentPublishWxid,
      currentSecret,
      paginatedMediaList,
      loadingPublishMedia,
      handleMediaSelectionChange,
      previewLinks,
      getPreviewLink,
      openPreview,
      copyPreviewLink,
      deleteMediaId,
      handlePageSizeChange,
      publishMediaTotal,
      publishMediaCurrentPage,
      fetchPublishMediaList
    };
  },
});
</script>


<style scoped>
/* 添加一些样式以美化分页组件 */
.el-pagination {
  margin-top: 20px;
  text-align: right;
}
</style>
