import axios from 'axios';
import type {
    ApiResponse,
    BangumiItem,
    SubgroupResource,
    SubscribePayload,
    UpdateOffsetPayload,
    PreviewRenameRequest,
    PreviewRenameResponse,
} from '../types/anime';

const api = axios.create({
  baseURL: '/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
});

export const AnimeAPI = {
  // 1. 获取当季全量新番列表
  getCurrentBangumis: async (): Promise<BangumiItem[]> => {
    const res = await api.get<ApiResponse<BangumiItem[]>>('/bangumi/current');
    return res.data.data;
  },

  // 2. 获取单部番剧的字幕组懒加载详情 (包含 RSS 链接与最新更新文件)
  getBangumiDetail: async (mikanId: number): Promise<SubgroupResource[]> => {
    const res = await api.get<ApiResponse<SubgroupResource[]>>(`/bangumi/${mikanId}/detail`);
    return res.data.data;
  },

  // 3. 手动触发增量同步
  syncBangumis: async (): Promise<string> => {
    const res = await api.post<ApiResponse<string>>('/bangumi/sync');
    return res.data.data;
  },

  // 4. 发起一键订阅
  subscribe: async (payload: SubscribePayload): Promise<string> => {
    const res = await api.post<ApiResponse<string>>('/subscriptions', payload);
    return res.data.data;
  },

  // 5. 彻底下架/安全清理整季
  purgeSubscription: async (mikanId: number, deleteFiles = true): Promise<string> => {
    const res = await api.delete<ApiResponse<string>>(`/subscriptions/${mikanId}?delete_files=${deleteFiles}`);
    return res.data.data;
  },

  // 6. 更新集数偏移量
  updateOffset: async (payload: UpdateOffsetPayload): Promise<string> => {
    const res = await api.put<ApiResponse<string>>('/offsets', payload);
    return res.data.data;
  },

  // 7. 实时改名效果预览
  previewRename: async (payload: PreviewRenameRequest): Promise<PreviewRenameResponse> => {
    const res = await api.post<ApiResponse<PreviewRenameResponse>>('/offsets/preview', payload);
    return res.data.data;
  },
};