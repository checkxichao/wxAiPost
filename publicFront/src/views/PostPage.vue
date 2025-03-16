<template>
  <div>
    <!-- 微信账号列表卡片 -->
    <el-card class="card">
      <el-table :data="paginatedWechatAccounts" style="width: 100%" stripe>
        <el-table-column prop="id" label="ID" width="80"></el-table-column>
        <el-table-column prop="name" label="公众号名" width="150"></el-table-column>
        <el-table-column prop="wxid" label="微信openid" width="180"></el-table-column>
        <el-table-column prop="secret" label="密钥" width="300"></el-table-column>
        <el-table-column prop="bindWechat" label="绑定者" width="180"></el-table-column>
        <el-table-column prop="created_at" :formatter="formatDate" label="时间" width="300"></el-table-column>
        <el-table-column :label="'状态'" width="80" :formatter="statusFormatter"></el-table-column>
        <el-table-column label="操作" width="200">
          <template #default="scope">
            <el-button type="danger" size="small" @click="getTaskCount(scope.row.wxid)">获取任务</el-button>
            <el-button
                type="primary"
                size="small"
                @click="publishTask(scope.row.wxid, scope.row.secret)"
                v-if="taskCount && taskCount.wxid === scope.row.wxid && taskCount.count > 0">
              发布任务
            </el-button>

            <el-button
                type="primary"
                size="small"
                @click="openWriteDart(scope.row.wxid, scope.row.secret)"
                v-if="taskCount && taskCount.wxid === scope.row.wxid && scope.row.nowState !== 2">
              草稿任务
            </el-button>
            <el-button
                type="warning"
                size="small"
                @click="stopTask(scope.row.wxid)"
                v-if="taskCount && taskCount.wxid === scope.row.wxid">
              停止任务
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页组件 -->
      <div class="pagination-wrapper">
        <el-pagination
            @size-change="handleWechatPageSizeChange"
            @current-change="handleWechatCurrentChange"
            :current-page="wechatCurrentPage"
            :page-sizes="[10, 20, 30, 50]"
            :page-size="wechatPageSize"
            layout="total, sizes, prev, pager, next, jumper"
            :total="wechatTotal">
        </el-pagination>
      </div>
    </el-card>

    <!-- 任务计数显示卡片 -->
    <el-card class="card" v-if="taskCount !== null">
      <div>
        {{ taskCount.wxid }}: 当前待发布文章数：<span>{{ taskCount.count }}</span>
      </div>

      <div style="margin-top: 10px;">
        <el-input
            v-model="publishLog"
            type="textarea"
            placeholder="记录发布日志"
            rows="4"
            readonly>
        </el-input>
      </div>
    </el-card>

    <!-- 发布媒体列表卡片 -->
    <el-card class="card" v-if="isPublishListVisible">
      <div style="margin-bottom: 20px;">
        <h3>发布媒体任务</h3>

        <!-- 选择模式和选择数量的控件 -->
        <div class="publish-controls">
          <el-radio-group v-model="publishSelectionMode" @change="handlePublishSelectionModeChange">
            <el-radio label="random">随机模式</el-radio>
            <el-radio label="sequential">顺序模式</el-radio>
            <el-radio label="manual">手动选择</el-radio>
          </el-radio-group>


          <!-- 随机模式和顺序模式的数量输入 -->
          <template v-if="form.titleSelectionMethod !== 'manual'">
            <el-input-number
                v-model="publishSelectionCount"
                :min="1"
                :max="publishSelectionCountMax"
                style="margin-left: 20px;"
                :disabled="isPublishSelectionCountDisabled"
                @change="handlePublishSelectionCountChange"
            />
            <span style="margin-left: 10px;">选择数量</span>
          </template>

        </div>

        <!-- 手动选择媒体 ID -->
        <div v-if="publishSelectionMode === 'manual'" style="margin-top: 20px;">
          <h4>可发布的 Media ID 列表</h4>
          <div style="margin-bottom: 10px; display: flex; align-items: center;">
            <el-input
                v-model="mediaSearchQuery"
                placeholder="搜索 Media ID 或标题"
                clearable
                style="width: 300px; margin-right: 10px;">
              <template #append>
                <el-button icon="el-icon-search" @click="handleSearch">搜索</el-button>
              </template>
            </el-input>
          </div>
          <div style="margin-top: 10px; text-align: left;">
            <el-button
                type="warning"
                @click="cleanInvalidDrafts(currentPublishWxid, currentSecret)"
                :disabled="paginatedMediaList.length === 0">
              清理无效草稿
            </el-button>
          </div>
          <el-table
              :data="paginatedMediaList"
              style="width: 100%"
              stripe
              border
              @selection-change="handleSelectionChange"
              v-loading="loadingPublishMedia">
            <el-table-column type="selection" width="30"></el-table-column>
            <el-table-column prop="id" label="ID" width="70"></el-table-column>
            <el-table-column prop="mediaId" label="Media ID" width="580"></el-table-column>
            <el-table-column prop="title" label="标题" width="580"></el-table-column>
            <el-table-column prop="createdAt" :formatter="formatDate" label="创建时间" width="140"></el-table-column>
            <!-- 预览草稿列 -->
            <el-table-column label="预览草稿" width="170">
              <template #default="scope">
                <el-button-group>
                  <!-- 获取预览按钮：如果没有预览链接显示 -->
                  <el-button
                      type="primary"
                      size="small"
                      v-if="!previewLinks[scope.row.mediaId]"
                      @click="getPreviewLink(scope.row)">
                    获取预览
                  </el-button>

                  <!-- 删除草稿按钮：始终显示 -->
                  <el-button
                      type="danger"
                      size="small"
                      @click="deleteMediaId(scope.row.mediaId)">
                    删除草稿
                  </el-button>

                  <!-- 打开预览按钮：只有有预览链接时才显示 -->
                  <el-button
                      type="success"
                      size="small"
                      v-if="previewLinks[scope.row.mediaId]"
                      @click="openPreview(previewLinks[scope.row.mediaId])">
                    打开预览
                  </el-button>

                  <!-- 复制链接按钮：只有有预览链接时才显示 -->
                  <el-button
                      type="info"
                      size="small"
                      v-if="previewLinks[scope.row.mediaId]"
                      @click="copyPreviewLink(previewLinks[scope.row.mediaId])">
                    复制链接
                  </el-button>
                </el-button-group>
              </template>
            </el-table-column>
          </el-table>

          <!-- 分页控件 -->
          <div style="margin-top: 10px; display: flex; justify-content: center; align-items: center;">
            <el-pagination
                @size-change="handlePageSizeChange"
                @current-change="handlePageChange"
                :current-page="currentPage"
                :page-sizes="[10, 20, 30, 50]"
                :page-size="pageSize"
                layout="total, sizes, prev, pager, next, jumper"
                :total="publishMediaTotal"
                background>
            </el-pagination>
          </div>
        </div>


        <!-- 随机模式和顺序模式的媒体列表显示（只读） -->
        <div v-if="publishSelectionMode !== 'manual'" style="margin-top: 20px;">
          <h4>可发布的 Media ID 列表</h4>
          <el-table
              :data="selectedMediaToPublish"
              style="width: 100%"
              stripe
              border>
            <el-table-column prop="id" label="ID" width="60"></el-table-column>
            <el-table-column prop="mediaId" label="Media ID" width="580"></el-table-column>
            <el-table-column prop="title" label="标题" width="600"></el-table-column>
            <el-table-column prop="createdAt" :formatter="formatDate" label="创建时间" width="140"></el-table-column>
            <!-- 预览草稿列 -->
            <el-table-column label="预览草稿" width="170">
              <template #default="scope">
                <el-button-group>
                  <!-- 获取预览按钮：如果没有预览链接显示 -->
                  <el-button
                      type="primary"
                      size="small"
                      v-if="!previewLinks[scope.row.mediaId]"
                      @click="getPreviewLink(scope.row)">
                    获取预览
                  </el-button>

                  <!-- 删除草稿按钮：始终显示 -->
                  <el-button
                      type="danger"
                      size="small"
                      @click="deleteMediaId(scope.row.mediaId)">
                    删除草稿
                  </el-button>

                  <!-- 打开预览按钮：只有有预览链接时才显示 -->
                  <el-button
                      type="success"
                      size="small"
                      v-if="previewLinks[scope.row.mediaId]"
                      @click="openPreview(previewLinks[scope.row.mediaId])">
                    打开预览
                  </el-button>

                  <!-- 复制链接按钮：只有有预览链接时才显示 -->
                  <el-button
                      type="info"
                      size="small"
                      v-if="previewLinks[scope.row.mediaId]"
                      @click="copyPreviewLink(previewLinks[scope.row.mediaId])">
                    复制链接
                  </el-button>
                </el-button-group>
              </template>
            </el-table-column>


          </el-table>
        </div>

        <div style="text-align: right; margin-top: 20px;">
          <el-button @click="cancelPublish">取消</el-button>
          <el-button
              type="primary"
              @click="confirmPublish"
              :disabled="isConfirmPublishDisabled">
            确认发布
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- 草稿任务表单 -->
    <div v-if="isWriteDartVisible">
      <el-card class="card">
        <el-form :model="form" ref="formRef" :rules="formRules" label-width="100px">
          <!-- 素材ID 输入框 -->
          <el-form-item label="素材ID" prop="materialId">
            <el-select v-model="form.materialId" placeholder="请选择素材ID" filterable clearable :loading="loading"
                       @change="onMaterialChange">
              <el-option
                  v-for="item in materials"
                  :key="item.thumbMediaId"
                  :label="`${item.thumbMediaId} - ${item.note || '无备注'}`"
                  :value="item.thumbMediaId">
                <template #default>
                  <img :src="item.imgUrl || 'default-image-path.jpg'" alt="素材图片"
                       style="width: 30px; height: 30px; margin-right: 10px;"/>
                  {{ item.thumbMediaId }} - {{ item.note || '无备注' }}
                </template>
              </el-option>
            </el-select>
          </el-form-item>

          <!-- 显示素材图片 -->
          <el-form-item v-if="selectedMaterialImgUrl" label="素材图片">
            <div style="display: flex; align-items: center;">
              <img
                  :src="selectedMaterialImgUrl"
                  alt="素材图片"
                  style="width: 100px; height: 100px; object-fit: cover; margin-right: 10px;"
              />
              <el-button
                  type="primary"
                  size="small"
                  @click="copyImageUrl"
                  style="margin-left: 10px;"
              >
                复制图片地址
              </el-button>
              <el-button
                  type="danger"
                  size="small"
                  @click="deleteImg"
                  style="margin-left: 10px;">
                删除这个素材
              </el-button>
            </div>
          </el-form-item>

          <el-form-item label="草稿数量" prop="draftCount">
            <el-input-number
                v-model="form.draftCount"
                :min="1"
                :max="publishSelectionMode !== 'manual' ? draftTaskCountMax : draftCountMax"
                :disabled="publishSelectionMode === 'manual'"
                placeholder="请输入草稿数量">
            </el-input-number>

            <div v-if="publishSelectionMode === 'manual'" style="margin-top: 5px;">
              <span v-if="selectedTitleIds.length > 0">当前选择标题数量：{{ form.draftCount }}</span>
              <span v-else style="color: red;">当前没有可用的标题，请先添加标题。</span>
            </div>
          </el-form-item>


          <el-form-item label="备注" prop="note">
            <el-input v-model="form.note" placeholder="请输入图片备注"></el-input>
          </el-form-item>

          <el-form-item label="模板" prop="template">
            <el-select v-model="form.template" placeholder="请选择模板" clearable>
              <el-option
                  v-for="template in templates"
                  :key="template"
                  :label="template"
                  :value="template">
              </el-option>
            </el-select>
          </el-form-item>

          <el-form-item label="标题选择" prop="titleSelectionMethod">
            <el-radio-group v-model="form.titleSelectionMethod" @change="handleTitleSelectionMethodChange">
              <el-radio label="random">随机</el-radio>
              <el-radio label="sequential">顺序</el-radio>
              <el-radio label="manual">手动</el-radio>
            </el-radio-group>


          </el-form-item>

          <div v-if="form.titleSelectionMethod === 'manual'" style="margin-top: 20px;">
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
            <el-form-item label="选择标题" prop="titleId" v-if="form.titleSelectionMethod === 'manual'">
              <el-table
                  ref="titleTable"
                  :data="manualTitles"
                  style="width: 100%"
                  stripe
                  border
                  size="small"
                  @selection-change="handleTitleSelectionChangeForWriteDart"
                  v-loading="loading"
              >
                <el-table-column type="selection" width="55"></el-table-column>
                <el-table-column prop="id" label="ID" width="80"></el-table-column>
                <el-table-column prop="title" label="标题" width="300"></el-table-column>
              </el-table>

              <el-pagination
                  @size-change="handleManualTitlePageSizeChange"
                  @current-change="handleManualTitleCurrentChange"
                  :current-page="manualCurrentPage"
                  :page-sizes="[10, 20, 50, 100]"
                  :page-size="manualPageSize"
                  layout="total, sizes, prev, pager, next, jumper"
                  :total="manualTotal"
                  style="margin-top: 10px;"
              >
              </el-pagination>

            </el-form-item>
          </div>

          <el-form-item label="模板模式" prop="templateSelectionMethod">
            <el-radio-group v-model="form.templateSelectionMethod">
              <el-radio label="intro">介绍模式</el-radio>
              <el-radio label="ranking">排行榜模式</el-radio>
            </el-radio-group>
          </el-form-item>

          <el-form-item label="上传封面图片">
            <el-upload
                action=""
                :before-upload="handleBeforeUpload"
                :show-file-list="false"
                accept="image/*"
                drag
                :on-error="handleUploadError"
                :on-progress="handleUploadProgress"
            >
              <i class="el-icon-upload"></i>
              <div class="el-upload__text">点击上传(封面图片可复用)</div>
            </el-upload>

            <el-progress v-if="uploadProgress > 0" :percentage="uploadProgress" status="active"/>
          </el-form-item>

          <el-form-item>
            <el-button type="primary" @click="submitWriteDart"
                       :disabled="!form.materialId || !form.draftCount || !form.template">
              确定
            </el-button>
            <el-button @click="closeWriteDart">取消</el-button>
          </el-form-item>
        </el-form>
      </el-card>
    </div>
  </div>
</template>

<script lang="ts">
import {defineComponent, reactive, ref, onMounted, nextTick, computed, onBeforeUnmount,watch} from 'vue';
import {
  ElMessage,
  ElButton,
  ElCard,
  ElForm,
  ElInput,
  ElInputNumber,
  ElSelect,
  ElOption,
  ElUpload,
  ElProgress,
  ElRadioGroup,
  ElRadio,
  ElTable,
  ElPagination,
  ElMessageBox,
} from 'element-plus';
import http from '../services/http';
import dayjs from 'dayjs';
import debounce from "lodash/debounce"; // 引入 dayjs 库用于日期格式化
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

interface Material {
  thumbMediaId: string;
  imgUrl: string;
  note: string;
}

interface TitleItem {
  id: number;
  title: string;
}
interface PublishMediaData {
  data: MediaItem[];
  total: number;
}
interface PublishMediaResponse {
  code: number;
  message: string;
  data: PublishMediaData;
}

interface TitleListResponse {
  code: number;
  message: string;
  data: {
    data: TitleItem[]; // 标题数组
    total: number;     // 总记录数
  };
}


interface MediaItem {
  id: number;
  mediaId: string;
  belongWxid: string;
  state: boolean;
  title:string;
  createdAt: string;
  updated_at: string;
}

export default defineComponent({
  name: 'PostPage',
  setup() {

    const wechatAccounts = ref<WechatAccount[]>([]);
    const taskCount = ref<TaskCount | null>(null);
    const publishLog = ref<string>('');
    const isWriteDartVisible = ref(false);
    const wechatTotal = ref<number>(0); // 新增总记录数
    const manualTitles = ref<TitleItem[]>([]); // 手动选择模式下的标题列表
    const manualCurrentPage = ref<number>(1); // 当前页码
    const manualPageSize = ref<number>(10); // 每页显示数量，适当增加以减少请求次数
    const manualTotal = ref<number>(0); // 总记录数
    const draftCountMax = computed(() => {
      // 仅在发布任务模式下使用
      if (isPublishListVisible.value) {
        if (publishSelectionMode.value === 'manual') {
          return selectedTitleIds.value.length;
        } else {
          return Math.max(1, publishMediaList.value.length);
        }
      }
      return 1; // 默认值
    });
    const titleSearchQuery = ref<string>('');
    const formatDate = (row: any, column: any, cellValue: string) => {
      if (!cellValue) return '';
      // 使用 dayjs 格式化日期
      return dayjs(cellValue).format('YYYY-MM-DD HH:mm:ss');
    };

    const allPublishMediaList = ref<MediaItem[]>([]); // 存储所有媒体数据
    const loadingManualTitles = ref<boolean>(false);
    const isPublishSelectionCountDisabled = computed(() => {
      // 在手动模式下，禁用输入框
      if (publishSelectionMode.value === 'manual') {
        return true;
      }
      // 在随机和顺序模式下，如果没有可发布的媒体，则禁用输入框
      return publishMediaList.value.length === 0;
    });
    const publishSelectionCountMax = computed(() => {
      return publishMediaList.value.length > 0 ? publishMediaList.value.length : 1;
    });

    const selectedTitleIds = ref<number[]>([]);
    const titleTable = ref<any>(null); // 用于引用el-table组件
    const handleTitleSelectionMethodChange = (method: SelectionMode) => {
      if (method === 'manual') {
        manualTitles.value = [];
        manualCurrentPage.value = 1;
        manualTotal.value = 0;
        selectedTitleIds.value = [];
        fetchManualTitles();
      } else {
        form.titleId = [];
        selectedTitleIds.value = [];
        form.draftCount = 1; // 重置草稿数量
      }
    };


// 处理草稿任务标题选择变化

    const handleTitleSelectionChangeForWriteDart = (selection: TitleItem[]) => {
      selectedTitleIds.value = selection.map(item => item.id);
      form.titleId = selectedTitleIds.value;

      if (form.titleSelectionMethod === 'manual') {
        form.draftCount = selection.length; // 动态更新草稿数量
        console.log('选中的标题ID:', selectedTitleIds.value);
        console.log('草稿数量已更新为:', form.draftCount);
      }
    };


    const handleManualTitlePageSizeChange = (newSize: number) => {
      manualPageSize.value = newSize;
      manualCurrentPage.value = 1;
      manualTitles.value = [];
      fetchManualTitles();
    };

    const handleManualTitleCurrentChange = (newPage: number) => {
      manualCurrentPage.value = newPage;
      fetchManualTitles();
    };

    const form = reactive({
      materialId: '',  // 素材ID
      draftCount: 1,    // 草稿任务数量
      wxid: '',
      secret: '',
      note: '',         // 备注字段
      template: '',     // 模板字段
      titleId: [] as number[], // 选择的标题ID，改为数组
      titleSelectionMethod: 'random' as 'random' | 'sequential' | 'manual',  // 标题选择方法
      templateSelectionMethod: 'intro' as 'intro' | 'ranking',  // 模板选择方法，'intro' 或 'ranking'
    });

    const materials = ref<Material[]>([]);
    const templates = ref<string[]>([]); // 存储模板列表
    const titles = ref<TitleItem[]>([]); // 存储标题列表
    const loading = ref(false);
    const selectedMaterialImgUrl = ref<string>(''); // 存储选择素材的图片 URL
    const uploadProgress = ref<number>(0); // 用于显示上传进度
    const mediaSearchQuery = ref<string>(''); // 搜索查询
    // 新增的数据属性
    const isPublishListVisible = ref(false);
    const publishMediaList = ref<MediaItem[]>([]);
    const selectedMediaIds = ref<string[]>([]);
    const currentPublishWxid = ref<string>(''); // 当前发布任务的wxid
    const currentSecret = ref<string>(''); // 当前发布任务的 secret
    const previewLinks = reactive<Record<string, string>>({}); // 存储预览链接，key 为 mediaId
    const publishMediaCurrentPage = ref<number>(1); // 当前页码
    type SelectionMode = 'random' | 'sequential' | 'manual';
    const currentPage = ref<number>(1);
    const pageSize = ref<number>(10);
    const totalPages = computed(() => Math.ceil(publishMediaTotal.value / pageSize.value));
    const paginatedMediaList = computed(() => {
      const start = (currentPage.value - 1) * pageSize.value;
      const end = start + pageSize.value;
      return publishMediaList.value.slice(start, end);
    });

    const filteredMediaList = computed(() => {
      if (!mediaSearchQuery.value) {
        return allPublishMediaList.value;
      }
      const query = mediaSearchQuery.value.toLowerCase();
      return allPublishMediaList.value.filter(item =>
          item.mediaId.toLowerCase().includes(query) ||
          item.title.toLowerCase().includes(query)
      );
    });
    // 更新分页数据
    const updatePaginatedMediaList = () => {
      publishMediaTotal.value = filteredMediaList.value.length;
      // 如果当前页码超过总页数，重置到最后一页
      const totalPages = Math.ceil(publishMediaTotal.value / pageSize.value);
      if (currentPage.value > totalPages) {
        currentPage.value = totalPages || 1;
      }
    };
    const handleSearch = () => {
      publishMediaCurrentPage.value = 1; // 重置到第一页
      fetchPublishMediaList();
    };

// 分页处理函数：页码变化
    const handlePageChange = (newPage: number) => {
      currentPage.value = newPage;
    };

// 分页处理函数：页面大小变化
    const handlePageSizeChange = (newSize: number) => {
      pageSize.value = newSize;
      currentPage.value = 1; // 重置到第一页
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

    const publishMediaTotal = ref<number>(0); // 总记录数
    const loadingPublishMedia = ref<boolean>(false); // 加载状态
    const publishSelectionMode = ref<SelectionMode>('random'); // 默认随机模式
    const publishMediaPageSize = ref<number>(10); // 每页显示数量
    const publishSelectionCount = ref<number>(1); // 默认选择数量为1
    watch(selectedMediaIds, (newVal) => {
      if (publishSelectionMode.value === 'manual') {
        publishSelectionCount.value = newVal.length;
      }
    });
    const formRules = reactive({
      materialId: [{required: true, message: '请填写素材ID', trigger: 'blur'}],
      draftCount: [
        {
          validator: (rule: any, value: any, callback: any) => {
            if (form.titleSelectionMethod === 'manual') {
              if (value !== selectedTitleIds.value.length) {
                callback(new Error(`草稿数量必须等于选中的标题数量 (${selectedTitleIds.value.length})`));
                return;
              }
            } else {
              if (value < 1) {
                callback(new Error('草稿数量至少为1'));
                return;
              }
            }
            callback();
          },
          trigger: 'change'
        }
      ],
      template: [{required: true, message: '请选择模板', trigger: 'change'}],
      titleSelectionMethod: [{required: true, message: '请选择标题选择方法', trigger: 'change'}],
      titleId: [
        {
          required: () => publishSelectionMode.value === 'manual',
          message: '请选择标题',
          trigger: 'change',
          validator: (rule: any, value: any, callback: any) => {
            if (publishSelectionMode.value === 'manual' && (!value || value.length === 0)) {
              callback(new Error('请选择标题'));
            } else {
              callback();
            }
          }
        }
      ]
    });


    const deleteImg = async () => {
      // 检查必要的信息是否存在
      if (!form.wxid || !form.secret || !form.materialId) {
        ElMessage.error("缺少必要的信息，无法删除素材");
        return;
      }

      try {
        // 弹出确认对话框
        await ElMessageBox.confirm(
            '确认要删除这个素材吗？此操作无法撤销。',
            '提示',
            {
              confirmButtonText: '确定',
              cancelButtonText: '取消',
              type: 'warning',
            }
        );
      } catch {
        // 用户取消删除操作
        return;
      }

      try {
        // 构建请求负载
        const payload = {
          wxid: form.wxid,
          secret: form.secret,
          thumbMediaId: form.materialId,
        };

        // 发送删除请求到后端
        const response = await http.post('/send/deleteThumb', payload);

        if (response.data.code === 200) {
          ElMessage.success('素材删除成功');
          // 清除当前选择的素材信息
          form.materialId = '';
          selectedMaterialImgUrl.value = '';
          window.location.reload();
        } else {
          ElMessage.error(`删除素材失败: ${response.data.message}`);
        }
      } catch (error: any) {
        // 处理请求错误
        const errorMessage = error.response?.data?.message || error.message || '未知错误';
        ElMessage.error(`删除素材失败: ${errorMessage}`);
      }
    };


    // 复制图片地址的方法
    const copyImageUrl = () => {
      if (!selectedMaterialImgUrl.value) {
        ElMessage.error("没有图片地址可复制！");
        return;
      }

      navigator.clipboard
          .writeText(selectedMaterialImgUrl.value)
          .then(() => {
            ElMessage.success("图片地址已复制到剪切板，请粘贴到浏览器新标签页中打开！");
          })
          .catch((err) => {
            console.error("复制失败：", err);
            ElMessage.error("复制失败，请手动复制图片地址！");
          });
    };
    const handleDropdownScroll = (event: Event) => {
      const target = event.target as HTMLElement;
      // 当滚动到下方90%时加载更多数据
      if (target.scrollTop + target.clientHeight >= target.scrollHeight - 50) {
        fetchManualTitles();
      }
    };

    const onSelectDropdownChange = (visible: boolean) => {
      if (publishSelectionMode.value === 'manual' && visible) {
        // 当下拉框打开时，初始化manualTitles
        manualTitles.value = [];
        manualCurrentPage.value = 1;
        manualTotal.value = 0;
        fetchManualTitles(); // 加载第一页

        // 使用nextTick确保DOM已更新
        nextTick(() => {
          const dropdown = document.querySelector('.el-select-dropdown') as HTMLElement;
          if (dropdown) {
            dropdown.addEventListener('scroll', handleDropdownScroll);
          }
        });
      } else if (publishSelectionMode.value === 'manual' && !visible) {
        // 当下拉框关闭时，移除滚动事件监听
        const dropdown = document.querySelector('.el-select-dropdown') as HTMLElement;
        if (dropdown) {
          dropdown.removeEventListener('scroll', handleDropdownScroll);
        }
      }
    };

// 组件卸载前，确保移除事件监听
    onBeforeUnmount(() => {
      const dropdown = document.querySelector('.el-select-dropdown') as HTMLElement;
      if (dropdown) {
        dropdown.removeEventListener('scroll', handleDropdownScroll);
      }
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

    const fetchManualTitles = async () => {
      if(titleSearchQuery.value!=""){
        try {

          const response = await http.post<TitleListResponse>('/send/getTitleListSearch', {

            page: manualCurrentPage.value,
            pageSize: manualPageSize.value,
            search: titleSearchQuery.value, // 搜索参数
          });

          if (
              response.data.code === 200 &&
              response.data.data &&
              Array.isArray(response.data.data.data)
          ) {
            manualTitles.value = response.data.data.data; // 替换为当前页的数据
            manualTotal.value = response.data.data.total;
            // manualCurrentPage.value += 1; // 移除错误的递增

            // 使用 nextTick 确保 DOM 已更新
            nextTick(() => {
              if (titleTable.value) {
                manualTitles.value.forEach(title => {
                  if (selectedTitleIds.value.includes(title.id)) {
                    titleTable.value.toggleRowSelection(title, true);
                  }
                });
              }
            });
          } else {
            ElMessage.error(`获取标题失败: ${response.data.message}`);
          }
        } catch (error) {
          console.error('获取标题列表失败:', error);
          ElMessage.error('获取标题列表失败');
        }
      }else{
        try {

          const response = await http.get<TitleListResponse>('/send/getTitleList', {
            params: {
              page: manualCurrentPage.value,
              pageSize: manualPageSize.value,
            },
          });

          if (
              response.data.code === 200 &&
              response.data.data &&
              Array.isArray(response.data.data.data)
          ) {
            manualTitles.value = response.data.data.data; // 替换为当前页的数据
            manualTotal.value = response.data.data.total;
            // manualCurrentPage.value += 1; // 移除错误的递增

            // 使用 nextTick 确保 DOM 已更新
            nextTick(() => {
              if (titleTable.value) {
                manualTitles.value.forEach(title => {
                  if (selectedTitleIds.value.includes(title.id)) {
                    titleTable.value.toggleRowSelection(title, true);
                  }
                });
              }
            });
          } else {
            ElMessage.error(`获取标题失败: ${response.data.message}`);
          }
        } catch (error) {
          console.error('获取标题列表失败:', error);
          ElMessage.error('获取标题列表失败');
        }
      }

    };


    // 获取微信账号列表
    const fetchWechatAccounts = async () => {
      try {
        const response = await http.get('/wechat/getWechat');
        wechatAccounts.value = response.data.info;
        wechatTotal.value = response.data.info.length;
      } catch (error) {
        ElMessage.error('获取微信信息失败');
      }
    };

    // 获取素材列表
    const fetchMaterials = async (wxid: string) => {
      loading.value = true;
      try {
        const response = await http.post('/send/getThumbList', {wxid}, {});
        console.log('响应数据:', response); // 输出响应内容，查看结构

        if (response.data && response.data.data.list && Array.isArray(response.data.data.list)) {
          materials.value = response.data.data.list; // 将素材列表赋值给 materials

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

    // 获取模板列表
    const fetchTemplates = async () => {
      try {
        const response = await http.get('/send/getTemplateList'); // 替换为实际的模板接口路径
        if (response.data && Array.isArray(response.data.data.templates)) {
          templates.value = response.data.data.templates;
        } else {
          ElMessage.error('获取模板失败');
        }
      } catch (error) {
        console.error('获取模板失败:', error);
        ElMessage.error('获取模板失败');
      }
    };

    // 获取标题列表
    const fetchTitles = async () => {
      try {
        const response = await http.get<TitleListResponse>('/send/getTitleList', {
          params: {
            page: manualCurrentPage.value,
            pageSize: manualPageSize.value,
          },
        });

        console.log('获取到的标题列表:', response.data); // 调试日志

        if (response.data.code === 200 && response.data.data && Array.isArray(response.data.data.data)) {
          manualTitles.value = response.data.data.data; // 更新标题数据
          manualTotal.value = response.data.data.total; // 更新总记录数
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


    // 监听素材选择变化，更新图片显示
    const onMaterialChange = () => {
      const selectedMaterial = materials.value.find(item => item.thumbMediaId === form.materialId);
      if (selectedMaterial) {
        nextTick(() => {
          selectedMaterialImgUrl.value = selectedMaterial.imgUrl || 'default-image-path.jpg';
        });
      }else{
        selectedMaterialImgUrl.value = '';
      }
    };

    // 上传图片并获取微信返回的URL和Media ID
    const uploadImage = async (wxid: string, secret: string, file: File, note: string = '') => {
      // 创建 FormData 用于文件上传
      const formData = new FormData();
      formData.append("wxid", wxid); // 必须传递 wxid
      formData.append("secret", secret); // 传递 secret
      formData.append("note", note); // 传递备注，无论是否为空
      formData.append("image", file); // 传递图片文件

      try {
        // 发起 POST 请求
        const response = await http.post('/send/uploadThumb', formData, {
          headers: {
            "Content-Type": "multipart/form-data", // 确保设置文件上传类型
          },
        });

        // 检查返回结果
        if (response.data && response.data.message === "上传成功") {
          ElMessage.success(`上传成功，URL: ${response.data.url}`);
          return {
            url: response.data.url,
            media_id: response.data.media_id,
          };
        } else {
          ElMessage.error(`上传失败: ${response.data.message || '未知错误'}`);
          return null;
        }
      } catch (error) {
        console.error("上传失败:", error);
        ElMessage.error("上传失败，请检查网络或服务器");
        return null;
      }
    };
    const draftTaskCountMax = computed(() => {
      if (publishSelectionMode.value === 'manual') {
        return selectedTitleIds.value.length;
      } else {
        return 10000; // 设置一个合理的最大值
      }
    });
// 添加或更新 handlePublishSelectionModeChange 方法
    const handlePublishSelectionModeChange = (newMode: SelectionMode) => {
      publishSelectionMode.value = newMode;

      if (newMode === 'manual') {
        // 切换到手动模式时，重置搜索查询并加载全部数据
        mediaSearchQuery.value = ''; // 清空搜索查询
        fetchPublishMediaList(); // 自动加载数据
      }else {
        // 对于随机和顺序模式，清空选择并可能加载相关数据
        selectedMediaIds.value = []; // 清空选择
        // 根据需要，可以在这里调用方法加载随机或顺序模式的数据
      }
    };


    // 在图片上传前处理
    const handleBeforeUpload = async (file: File) => {
      const wxid = form.wxid;
      const secret = form.secret;
      let note = form.note.trim(); // 去掉空格

      if (!wxid || !secret) {
        ElMessage.warning("请先填写微信账号信息");
        return false;
      }

      try {
        // 如果没有输入备注，直接传递空字符串
        const result = await uploadImage(wxid, secret, file, note ? note : '');
        if (result) {
          form.materialId = result.media_id; // 将返回的 media_id 赋值给表单
          ElMessage.success("图片上传成功！");
          form.note = ''; // 清空备注字段

          // 重新获取素材列表，刷新数据
          await fetchMaterials(wxid);

          // 重新获取模板列表，确保模板数据是最新的
          await fetchTemplates();

          // 重新获取标题列表，确保标题数据是最新的
          await fetchTitles();

          // 重置上传进度
          uploadProgress.value = 0;

          // 可选：显示上传确认提示
          ElMessage({
            message: '图片上传成功，您可以继续其他操作。',
            type: 'success',
            duration: 3000,
          });
        } else {
          ElMessage.error("图片上传失败！");
        }
      } catch (error) {
        ElMessage.error("图片上传过程中出现错误！");
      }

      // 阻止默认上传行为
      return false;
    };

    // 上传失败回调
    const handleUploadError = (error: any) => {
      console.error('上传失败:', error);
      ElMessage.error('图片上传失败');
    };

    // 上传进度回调
    const handleUploadProgress = (event: ProgressEvent) => {
      uploadProgress.value = Math.round((event.loaded / event.total) * 100);
    };

    // 获取任务数
    const getTaskCount = async (wxid: string) => {
      try {
        const response = await http.post('/send/getMedias', {wxid});
        taskCount.value = {wxid, count: response.data.data.length};
      } catch (error) {
        ElMessage.error('获取任务数失败');
      }
    };
// 状态重置函数
    const resetPublishState = () => {
      publishSelectionMode.value = 'random'; // 或者根据需求设置默认模式
      publishSelectionCount.value = 1;
      selectedMediaIds.value = [];
      mediaSearchQuery.value = '';
      currentPage.value = 1;
      pageSize.value = 10;
      publishMediaList.value = [];
      publishMediaTotal.value = 0;
      Object.keys(previewLinks).forEach(key => delete previewLinks[key]);
      isPublishListVisible.value = false;
    };

    // 发布任务
    const publishTask = async (wxid: string, secret: string) => {
      resetPublishState();
      if (currentPublishWxid.value !== wxid) {
        currentPublishWxid.value = wxid;
        currentSecret.value = secret;

        // 重置发布媒体相关状态
        publishMediaList.value = [];
        publishMediaTotal.value = 0;
        Object.keys(previewLinks).forEach(key => delete previewLinks[key]);
        selectedMediaIds.value = [];
        currentPage.value = 1;
        pageSize.value = 10;

        // 清空草稿任务相关状态
        form.materialId = '';
        selectedMaterialImgUrl.value = '';
      }
      // 获取Media ID列表
      try {
        const response = await http.post('/send/getMedias', {wxid}); // 替换为实际的获取Media ID的接口路径
        if (response.data.code === 200 && Array.isArray(response.data.data)) {
          publishMediaList.value = response.data.data;
          console.log('publishMediaList.length:', publishMediaList.value.length);
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

    // 处理媒体ID选择变化（手动模式）
    const handleSelectionChange = (selection: MediaItem[]) => {
      selectedMediaIds.value = selection.map(item => item.mediaId);
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

    // 删除草稿
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

    // 获取选中的媒体 ID 进行发布，根据模式和数量
    const getSelectedMediaIds = (): string[] => {
      const count = publishSelectionCount.value;
      const mode = publishSelectionMode.value;
      let selected: string[] = [];

      if (mode === 'manual') {
        // 手动选择模式
        selected = selectedMediaIds.value;
      } else {
        const available = publishMediaList.value.map(item => item.mediaId);

        if (count > available.length) {
          ElMessage.warning(`选择数量超过可发布媒体数量。最多可选 ${available.length} 个。`);
          return available;
        }

        if (mode === 'random') {
          // 随机打乱数组并取前count个
          const shuffled = [...available].sort(() => 0.5 - Math.random());
          selected = shuffled.slice(0, count);
        } else if (mode === 'sequential') {
          // 顺序取前count个
          selected = available.slice(0, count);
        }
      }

      return selected;
    };

    // 计算是否禁用确认发布按钮
    const isConfirmPublishDisabled = computed(() => {
      if (publishSelectionMode.value === 'manual') {
        return selectedMediaIds.value.length === 0;
      } else {
        return publishSelectionCount.value < 1 || publishMediaList.value.length === 0;
      }
    });

    // 计算已选择的媒体 ID（用于随机和顺序模式）
    const selectedMediaToPublish = computed(() => {
      if (publishSelectionMode.value === 'manual') {
        return publishMediaList.value.filter(item => selectedMediaIds.value.includes(item.mediaId));
      } else {
        const selectedIds = getSelectedMediaIds();
        return publishMediaList.value.filter(item => selectedIds.includes(item.mediaId));
      }
    });


    // 确认发布方法
    const confirmPublish = async () => {
      const selectedToPublish = getSelectedMediaIds();

      if (selectedToPublish.length === 0) {
        ElMessage.warning('请选择至少一个Media ID进行发布');
        return;
      }

      try {
        const payload = {
          wxid: currentPublishWxid.value,
          mediaIds: selectedToPublish,
        };
        const response = await http.post('/send/postDart', payload); // 替换为实际的发布任务接口

        if (response.data.code === 200) {
          publishLog.value += `成功发布任务！Media IDs: ${selectedToPublish.join(', ')}\n`;
          ElMessage.success('任务发布成功');
          isPublishListVisible.value = false;
          // 更新任务数
          await getTaskCount(currentPublishWxid.value);
          // 清空预览链接
          selectedToPublish.forEach(mediaId => {
            delete previewLinks[mediaId];
          });
          // 如果是手动选择，清空选择
          if (publishSelectionMode.value === 'manual') {
            selectedMediaIds.value = [];
          }
          // 清空发布媒体列表中已发布的媒体
          publishMediaList.value = publishMediaList.value.filter(item => !selectedToPublish.includes(item.mediaId));
        } else {
          publishLog.value += `发布任务失败: ${response.data.message}\n`;
          ElMessage.error(`发布任务失败: ${response.data.message}`);
        }
      } catch (error) {
        console.error('发布任务失败:', error);

        if (error instanceof Error) {
          publishLog.value += `发布任务失败: ${error.message || '未知错误'}\n`;
          ElMessage.error(`发布任务失败: ${error.message}`);
        } else {
          publishLog.value += `发布任务失败: 未知错误\n`;
          ElMessage.error('发布任务失败: 未知错误');
        }
      }
    };

    // 取消发布
    const cancelPublish = () => {
      isPublishListVisible.value = false;
      publishSelectionMode.value = 'random'; // 重置为默认随机模式
      publishSelectionCount.value = 1; // 重置为默认选择数量
      selectedMediaIds.value = []; // 清空选择
      // 清空预览链接
      for (const mediaId of publishMediaList.value.map(item => item.mediaId)) {
        previewLinks[mediaId] = '';
      }
    };


    // 打开草稿任务输入框
    const openWriteDart = (wxid: string, secret: string) => {
      isWriteDartVisible.value = true;
      form.wxid = wxid;
      form.secret = secret;
      // 重置相关状态
      form.materialId = '';
      selectedMaterialImgUrl.value = '';
      publishMediaList.value = [];
      publishMediaTotal.value = 0;
      Object.keys(previewLinks).forEach(key => delete previewLinks[key]);
      selectedMediaIds.value = [];
      currentPage.value = 1;
      pageSize.value = 10;

      fetchMaterials(wxid); // 传递 wxid
      fetchTemplates(); // 获取模板列表
      // 手动选择模式下，手动加载标题列表
      if (form.titleSelectionMethod === 'manual') {
        manualTitles.value = [];
        manualCurrentPage.value = 1;
        manualTotal.value = 0;
        fetchManualTitles();
      } else {
        fetchTitles(); // 获取标题列表
      }
    };


    // 提交草稿任务
    const submitWriteDart = async () => {
      const {
        wxid,
        secret,
        materialId,
        draftCount,
        template,
        titleId,
        titleSelectionMethod,
        templateSelectionMethod
      } = form;

      if (!materialId) {
        ElMessage.warning('请填写素材ID');
        return;
      }

      if (!template) {
        ElMessage.warning('请选择模板');
        return;
      }

      // 正确的条件判断
      if (titleSelectionMethod === 'manual' && selectedTitleIds.value.length !== draftCount) {
        ElMessage.error(`草稿数量必须等于选中的标题数量 (${selectedTitleIds.value.length})`);
        return;
      }

      let titlesToSend: string[] = [];

      if (titleSelectionMethod === 'manual') {
        const selectedTitles = manualTitles.value.filter(title => selectedTitleIds.value.includes(title.id));

        if (selectedTitles.length !== draftCount) {
          ElMessage.error('选择的标题数量必须等于草稿数量');
          return;
        }
        titlesToSend = selectedTitles.map(title => title.title);
        publishLog.value += `手动选择标题: ${titlesToSend.join(', ')}\n`;
      }

      try {
        const payload: any = {
          wxid,
          secret,
          materialId,
          draftCount,
          template,
          templateSelectionMethod,
          titleSelectionMethod,
        };

        if (titleSelectionMethod === 'manual') {
          payload.titles = titlesToSend;
        }

        const response = await http.post('/send/writeDart', payload);

        if (response.data.code === 200 && response.data.message.includes("开始")) {
          publishLog.value += `草稿任务已开始！ 数量: ${draftCount}, 模板: ${template}, 模板模式: ${templateSelectionMethod}` +
              (titleSelectionMethod === 'manual' ? `, 标题: ${titlesToSend.join(', ')}\n` : '\n') + `可草稿标题: ${response.data.data.join(', ')}\n` + `服务器返回信息: ${response.data.message}\n`;
          ElMessage.success('草稿任务已开始');
          isWriteDartVisible.value = false; // 关闭输入框
        } else {
          publishLog.value += `草稿任务失败: ${response.data.message}\n`;
          ElMessage.error(`草稿任务失败: ${response.data.message}`);
        }
      } catch (error) {
        publishLog.value += `草稿任务失败: \n`;
        ElMessage.error('草稿任务失败');
      }
    };


    // 关闭草稿任务输入框
    const closeWriteDart = () => {
      isWriteDartVisible.value = false;
    };

    // 停止任务操作
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

    // 状态映射函数
    const statusFormatter = (row: WechatAccount) => {
      switch (row.nowState) {
        case 0:
          return '无任务';
        case 1:
          return '草稿';
        case 2:
          return '发布中';
        default:
          return '未知状态';
      }
    };

    // 分页相关状态
    const wechatCurrentPage = ref<number>(1); // 当前页码
    const wechatPageSize = ref<number>(10); // 每页显示数量

    const paginatedWechatAccounts = computed(() => {
      const start = (wechatCurrentPage.value - 1) * wechatPageSize.value;
      const end = start + wechatPageSize.value;
      return wechatAccounts.value.slice(start, end);
    });


// 分页处理函数：页面大小变化
    const handleWechatPageSizeChange = (newSize: number) => {
      wechatPageSize.value = newSize;
      wechatCurrentPage.value = 1; // 重置到第一页
      fetchWechatAccounts();
    };

// 分页处理函数：页码变化
    const handleWechatCurrentChange = (newPage: number) => {
      wechatCurrentPage.value = newPage;
      fetchWechatAccounts();
    };
    const handleTitleSearch = () => {
      // 清空当前手动标题列表并重置分页
      manualTitles.value = [];
      manualCurrentPage.value = 1;
      manualTotal.value = 0;
      fetchManualTitles();
    };

    // 处理 publishSelectionCount 变化时的逻辑
    const handlePublishSelectionCountChange = (newVal: number) => {
      if (newVal > publishMediaList.value.length) {
        ElMessage.warning(`选择数量不能超过可发布媒体数量（最多 ${publishMediaList.value.length} 个）。`);
        publishSelectionCount.value = publishMediaList.value.length > 0 ? publishMediaList.value.length : 1;
      }
    };

    onMounted(() => {
      fetchWechatAccounts();
      // fetchTitles(); // 获取标题列表
    });

    return {
      wechatAccounts,
      taskCount,
      publishLog,
      isWriteDartVisible,
      loadingManualTitles,
      form,
      materials,
      templates,           // 返回模板列表
      titles,              // 返回标题列表
      loading,
      selectedMaterialImgUrl,
      getTaskCount,
      publishTask,
      openWriteDart,
      submitWriteDart,
      closeWriteDart,
      stopTask,
      draftTaskCountMax,
      statusFormatter,
      onMaterialChange,
      handleUploadError,
      handleBeforeUpload,
      uploadProgress,
      handleUploadProgress,
      copyImageUrl,
      onSelectDropdownChange,
      // 发布媒体列表相关
      fetchManualTitles,
      isPublishListVisible,
      mediaSearchQuery,
      publishMediaList,
      selectedMediaIds,
      handleSelectionChange,
      getPreviewLink, // 新增
      openPreview, // 新增
      copyPreviewLink, // 新增
      confirmPublish,
      cancelPublish,
      previewLinks, // 新增
      deleteMediaId,
      // 选择模式和选择数量
      publishSelectionMode,
      publishSelectionCount,
      getSelectedMediaIds,
      draftCountMax,
      publishSelectionCountMax,
      // 相关方法
      handlePublishSelectionCountChange,
      formRules,
      // 计算属性
      selectedMediaToPublish,
      isConfirmPublishDisabled,
      // 分页相关返回
      paginatedWechatAccounts,
      wechatCurrentPage,
      wechatPageSize,
      wechatTotal,
      handleWechatPageSizeChange,
      handleWechatCurrentChange,
      cleanInvalidDrafts,
      currentPublishWxid,
      currentSecret,
      manualTitles,
      loadingPublishMedia,
      handleTitleSelectionMethodChange,
      handleManualTitlePageSizeChange,
      handleManualTitleCurrentChange,
      manualCurrentPage,
      manualPageSize,
      manualTotal,
      handleTitleSelectionChangeForWriteDart,
      selectedTitleIds,
      isPublishSelectionCountDisabled,
      deleteImg,
      formatDate,
      handleSearch ,
      fetchPublishMediaList,
      publishMediaPageSize,
      publishMediaTotal,
      publishMediaCurrentPage,
      allPublishMediaList,
      totalPages,
      paginatedMediaList,
      handlePageChange,
      handlePageSizeChange,
      currentPage,
      pageSize,
      handlePublishSelectionModeChange,
      resetPublishState,
      titleSearchQuery,
      handleTitleSearch
    };
  },
});
</script>

<style scoped>
.card {
  margin-bottom: 20px;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}

/* 调整选择模式和选择数量的布局 */
.publish-controls {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
}

.publish-controls .el-radio-group {
  margin-right: 20px;
}
</style>
