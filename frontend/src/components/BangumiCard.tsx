import React from "react";
import type { BangumiItem } from "../types/anime";
import { Plus, Trash2, Calendar } from "lucide-react";

interface BangumiCardProps {
  bangumi: BangumiItem;
  onSubscribeClick: (bangumi: BangumiItem) => void;
  onPurgeClick: (bangumi: BangumiItem) => void;
}

const WEEKDAYS = ["周日", "周一", "周二", "周三", "周四", "周五", "周六"];

// 🎯 补全 TMDB 图片 CDN 前缀
const getPosterUrl = (path?: string) => {
  if (!path) return "";
  if (path.startsWith("http")) return path;
  return `https://image.tmdb.org/t/p/w600_and_h900_face${path}`;
};

export const BangumiCard: React.FC<BangumiCardProps> = ({
  bangumi,
  onSubscribeClick,
  onPurgeClick,
}) => {
  const isSubscribed = bangumi.sub_status === "subscribing";
  const posterUrl = getPosterUrl(bangumi.poster_path);

  return (
    <div className="group relative bg-slate-800/60 rounded-2xl overflow-hidden border border-slate-700/50 hover:border-indigo-500/50 transition-all duration-300 hover:shadow-2xl hover:shadow-indigo-500/10 flex flex-col">
      {/* 海报与悬停剧情遮罩 */}
      <div className="relative aspect-[2/3] w-full bg-slate-900 overflow-hidden">
        {posterUrl ? (
          <img
            src={posterUrl}
            alt={bangumi.title_cn || bangumi.title_raw}
            className="w-full h-full object-cover transition-transform duration-500 group-hover:scale-105"
            loading="lazy"
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center text-slate-600 font-medium">
            暂无海报
          </div>
        )}

        {/* 顶部徽章（放送周期 & 订阅状态） */}
        <div className="absolute top-3 left-3 right-3 flex items-center justify-between pointer-events-none z-10">
          <span className="flex items-center gap-1 text-xs font-semibold px-2.5 py-1 rounded-lg bg-slate-900/80 backdrop-blur-md text-slate-300 border border-slate-700/50">
            <Calendar className="w-3 h-3 text-indigo-400" />
            {WEEKDAYS[bangumi.broadcast_day] ?? "未知"}
          </span>

          <span
            className={`text-xs font-semibold px-2.5 py-1 rounded-lg backdrop-blur-md border ${
              isSubscribed
                ? "bg-emerald-500/20 text-emerald-400 border-emerald-500/30"
                : "bg-slate-900/80 text-slate-400 border-slate-700/50"
            }`}
          >
            {isSubscribed ? "订阅中" : "未订阅"}
          </span>
        </div>

        {/* 悬停平滑展示剧情简介 */}
        <div className="absolute inset-0 bg-slate-950/85 backdrop-blur-sm p-4 opacity-0 group-hover:opacity-100 transition-opacity duration-300 flex flex-col justify-end text-xs text-slate-300 leading-relaxed overflow-y-auto z-20">
          <p className="font-semibold text-white mb-1.5 text-sm">剧情大纲：</p>
          <p>{bangumi.overview || "暂无详细大纲信息..."}</p>
        </div>
      </div>

      {/* 卡片底部标题与交互按钮 */}
      <div className="p-4 flex-1 flex flex-col justify-between gap-3">
        <div>
          <h3 className="font-bold text-white text-base line-clamp-1 group-hover:text-indigo-300 transition-colors">
            {bangumi.title_cn || bangumi.title_raw || "未命名番剧"}
          </h3>
          <p className="text-xs text-slate-400 line-clamp-1 mt-0.5">
            {bangumi.title_raw}
          </p>
        </div>

        {/* 操作按钮 */}
        {isSubscribed ? (
          <button
            onClick={() => onPurgeClick(bangumi)}
            className="w-full py-2 px-3 rounded-xl bg-slate-700/50 hover:bg-rose-500/20 text-slate-300 hover:text-rose-400 border border-slate-600/50 hover:border-rose-500/30 text-xs font-medium flex items-center justify-center gap-1.5 transition-all"
          >
            <Trash2 className="w-3.5 h-3.5" />
            彻底下架/删除
          </button>
        ) : (
          <button
            onClick={() => onSubscribeClick(bangumi)}
            className="w-full py-2 px-3 rounded-xl bg-indigo-600 hover:bg-indigo-500 active:bg-indigo-700 text-white text-xs font-medium flex items-center justify-center gap-1.5 shadow-lg shadow-indigo-600/20 transition-all"
          >
            <Plus className="w-3.5 h-3.5" />
            一键订阅
          </button>
        )}
      </div>
    </div>
  );
};
