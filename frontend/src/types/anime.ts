// 单部番剧的元数据结构 (对应 Go 的 BangumiMetadata + 订阅状态)
export interface BangumiItem {
  mikan_id: number;
  title_cn: string;
  title_raw: string;
  tmdb_id?: number;
  overview?: string;
  poster_path?: string;
  air_date?: string;
  broadcast_day: number; // 0: 周日, 1: 周一 ... 6: 周六
  current_season: number;
  sub_status: 'none' | 'subscribing'; // 本地联表得出的订阅状态
}

// 详情页单集资源结构
export interface MikanEpisode {
  id: number;
  mikan_id: number;
  subgroup_id: number;
  title: string;
  size: string;
  publish_time: string;
  magnet: string;
}

// 详情页字幕组资源结构
export interface SubgroupResource {
  mikan_id: number;
  subgroup_id: number;
  subgroup_name: string;
  rss_url: string;
  episodes: MikanEpisode[];
}

// 订阅请求 Payload
export interface SubscribePayload {
  mikan_id: number;
  subgroup_id: number;
  season: number;
  rss_url: string;
  custom_offset?: number;
  must_contain?: string;
  must_not_contain?: string;
}

// 偏移量更新 Payload
export interface UpdateOffsetPayload {
  mikan_id: number;
  season: number;
  offset_value: number;
}

// 实时改名预览 Payload & Response
export interface PreviewRenameRequest {
  file_name: string;
  offset: number;
}

export interface PreviewRenameResponse {
  matched: boolean;
  old_name: string;
  new_name: string;
  relative_ep: string;
}

// 统一 API 响应格式
export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}