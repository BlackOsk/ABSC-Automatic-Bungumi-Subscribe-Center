import React, { useEffect, useState } from "react";
import type {
  BangumiItem,
  SubgroupResource,
  MikanEpisode,
} from "../types/anime";
import { AnimeAPI } from "../api/client";
import { X, Loader2, Filter, Layers, Radio } from "lucide-react";

interface SubscribeModalProps {
  bangumi: BangumiItem | null;
  onClose: () => void;
  onSuccess: () => void;
}

// 辅助函数：生成唯一的 Group Key
const getGroupKey = (group: SubgroupResource) => {
  return group.subgroup_id > 0
    ? `id_${group.subgroup_id}`
    : `name_${group.subgroup_name}`;
};

export const SubscribeModal: React.FC<SubscribeModalProps> = ({
  bangumi,
  onClose,
  onSuccess,
}) => {
  const [loading, setLoading] = useState(true);
  const [subgroups, setSubgroups] = useState<SubgroupResource[]>([]);
  const [selectedGroupKey, setSelectedGroupKey] = useState<string>("");
  const [season, setSeason] = useState<number>(bangumi?.current_season || 1);
  const [customOffset, setCustomOffset] = useState<string>("");
  const [mustContain, setMustContain] = useState<string>("");
  const [mustNotContain, setMustNotContain] = useState<string>("");
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!bangumi) return;

    let isMounted = true;

    AnimeAPI.getBangumiDetail(bangumi.mikan_id)
      .then((rawSubgroups) => {
        if (!isMounted) return;

        // 1. 过滤掉误识别的“订阅”字符串节点
        const validGroups = rawSubgroups.filter(
          (g) => g.subgroup_name !== "订阅" && g.subgroup_name !== "",
        );

        // 2. 按 Subgroup Key 合并去重与装载全量 episodes
        const subgroupMap = new Map<string, SubgroupResource>();

        validGroups.forEach((item) => {
          const key = getGroupKey(item);
          if (!subgroupMap.has(key)) {
            subgroupMap.set(key, {
              ...item,
              episodes: [...(item.episodes || [])],
            });
          } else {
            const existing = subgroupMap.get(key)!;
            if (!existing.rss_url && item.rss_url) {
              existing.rss_url = item.rss_url;
            }
            (item.episodes || []).forEach((ep) => {
              if (!existing.episodes.some((e) => e.title === ep.title)) {
                existing.episodes.push(ep);
              }
            });
          }
        });

        const mergedList = Array.from(subgroupMap.values());
        setSubgroups(mergedList);

        if (mergedList.length > 0) {
          setSelectedGroupKey(getGroupKey(mergedList[0]));
        }
      })
      .catch((err: unknown) => {
        if (!isMounted) return;
        const msg = err instanceof Error ? err.message : "获取字幕组详情失败";
        console.error(msg);
      })
      .finally(() => {
        if (isMounted) {
          setLoading(false);
        }
      });

    return () => {
      isMounted = false;
    };
  }, [bangumi]);

  if (!bangumi) return null;

  const currentGroupData = subgroups.find(
    (g) => getGroupKey(g) === selectedGroupKey,
  );

  const handleSubscribe = async () => {
    if (!currentGroupData) return;

    setSubmitting(true);
    try {
      await AnimeAPI.subscribe({
        mikan_id: bangumi.mikan_id,
        subgroup_id: currentGroupData.subgroup_id,
        season,
        rss_url: currentGroupData.rss_url,
        custom_offset:
          customOffset !== "" ? parseInt(customOffset, 10) : undefined,
        must_contain: mustContain.trim() || undefined,
        must_not_contain: mustNotContain.trim() || undefined,
      });
      onSuccess();
      onClose();
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : "未知错误";
      alert("订阅提交失败: " + errorMessage);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 bg-slate-950/80 backdrop-blur-md flex items-center justify-center p-4">
      <div className="bg-slate-900 border border-slate-800 rounded-3xl w-full max-w-2xl max-h-[90vh] overflow-hidden flex flex-col shadow-2xl">
        {/* Modal 头部 */}
        <div className="p-6 border-b border-slate-800 flex items-center justify-between">
          <div>
            <h2 className="text-xl font-bold text-white">
              订阅设置：《{bangumi.title_cn}》
            </h2>
            <p className="text-xs text-slate-400 mt-1">
              选择字幕组并配置下载过滤器与季数偏移
            </p>
          </div>
          <button
            onClick={onClose}
            className="p-2 text-slate-400 hover:text-white rounded-xl hover:bg-slate-800 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Modal Body */}
        <div className="p-6 overflow-y-auto flex-1 space-y-6 text-sm text-slate-300">
          {loading ? (
            <div className="py-12 flex flex-col items-center justify-center text-slate-400 gap-3">
              <Loader2 className="w-8 h-8 animate-spin text-indigo-500" />
              <p>正在懒加载 Mikan 字幕组与 RSS 规则...</p>
            </div>
          ) : (
            <>
              {/* 1. 精准去重后的字幕组选择按钮列表 */}
              <div>
                <label className="block font-semibold text-white mb-2 flex items-center gap-2">
                  <Radio className="w-4 h-4 text-indigo-400" />
                  选择字幕组
                </label>
                <div className="grid grid-cols-2 gap-2 max-h-36 overflow-y-auto p-1">
                  {subgroups.map((group) => {
                    const key = getGroupKey(group);
                    const isSelected = selectedGroupKey === key;
                    return (
                      <button
                        key={key}
                        onClick={() => setSelectedGroupKey(key)}
                        className={`p-3 rounded-xl text-left border transition-all flex items-center justify-between ${
                          isSelected
                            ? "bg-indigo-600/20 border-indigo-500 text-white font-medium"
                            : "bg-slate-800/50 border-slate-700/50 text-slate-300 hover:bg-slate-800"
                        }`}
                      >
                        <span className="truncate">{group.subgroup_name}</span>
                        {isSelected && (
                          <div className="w-2 h-2 rounded-full bg-indigo-400 shadow-sm shadow-indigo-400" />
                        )}
                      </button>
                    );
                  })}
                </div>
              </div>

              {/* 2. 当前选中字幕组的全量文件展示列表（带定制滚动条） */}
              {currentGroupData && (
                <div className="bg-slate-950/50 rounded-xl p-3.5 border border-slate-800/80 text-xs space-y-2">
                  <div className="flex items-center justify-between font-semibold text-slate-300">
                    <span>
                      【{currentGroupData.subgroup_name}】发布的历史文件：
                    </span>
                    <span className="text-indigo-400 text-[11px] bg-indigo-500/10 px-2 py-0.5 rounded-md border border-indigo-500/20">
                      共 {currentGroupData.episodes.length} 个文件
                    </span>
                  </div>

                  {/* 纵向无限滚动容器 */}
                  <div className="max-h-44 overflow-y-auto space-y-1.5 pr-1 text-slate-400 font-mono select-text">
                    {currentGroupData.episodes.length > 0 ? (
                      currentGroupData.episodes.map(
                        (ep: MikanEpisode, index: number) => (
                          <div
                            key={ep.id || index}
                            className="hover:text-slate-100 hover:bg-slate-800/60 p-1.5 rounded transition-colors truncate border border-transparent hover:border-slate-700/50"
                            title={ep.title}
                          >
                            • {ep.title}
                          </div>
                        ),
                      )
                    ) : (
                      <p className="text-slate-500 italic py-2">
                        该字幕组暂未展示具体文件列表
                      </p>
                    )}
                  </div>
                </div>
              )}

              {/* 3. 季数与偏移配置 */}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-semibold text-slate-300 mb-1.5 flex items-center gap-1.5">
                    <Layers className="w-3.5 h-3.5 text-indigo-400" />
                    目标季数 (Season)
                  </label>
                  <input
                    type="number"
                    min={1}
                    value={season}
                    onChange={(e) =>
                      setSeason(parseInt(e.target.value, 10) || 1)
                    }
                    className="w-full bg-slate-800 border border-slate-700 rounded-xl px-3 py-2 text-white focus:outline-none focus:border-indigo-500"
                  />
                </div>

                <div>
                  <label className="block text-xs font-semibold text-slate-300 mb-1.5">
                    自定义偏移集数 (可选)
                  </label>
                  <input
                    type="number"
                    placeholder="留空则自动从 TMDB 推导"
                    value={customOffset}
                    onChange={(e) => setCustomOffset(e.target.value)}
                    className="w-full bg-slate-800 border border-slate-700 rounded-xl px-3 py-2 text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500"
                  />
                </div>
              </div>

              {/* 4. 关键字过滤配置 */}
              <div>
                <label className="block font-semibold text-white mb-2 flex items-center gap-2">
                  <Filter className="w-4 h-4 text-indigo-400" />
                  qBittorrent RSS 过滤器关键词
                </label>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <span className="block text-xs text-slate-400 mb-1">
                      必须包含 (Must Contain):
                    </span>
                    <input
                      type="text"
                      placeholder="如: CHS 或 1080p"
                      value={mustContain}
                      onChange={(e) => setMustContain(e.target.value)}
                      className="w-full bg-slate-800 border border-slate-700 rounded-xl px-3 py-2 text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500"
                    />
                  </div>
                  <div>
                    <span className="block text-xs text-slate-400 mb-1">
                      排除包含 (Must Not Contain):
                    </span>
                    <input
                      type="text"
                      placeholder="如: 720p 或 繁中"
                      value={mustNotContain}
                      onChange={(e) => setMustNotContain(e.target.value)}
                      className="w-full bg-slate-800 border border-slate-700 rounded-xl px-3 py-2 text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500"
                    />
                  </div>
                </div>
              </div>
            </>
          )}
        </div>

        {/* Modal 底部按钮 */}
        <div className="p-6 border-t border-slate-800 flex items-center justify-end gap-3 bg-slate-900/50">
          <button
            onClick={onClose}
            className="px-5 py-2.5 rounded-xl border border-slate-700 text-slate-300 hover:bg-slate-800 text-sm font-medium transition-colors"
          >
            取消
          </button>
          <button
            onClick={handleSubscribe}
            disabled={submitting || loading || !currentGroupData}
            className="px-5 py-2.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 active:bg-indigo-700 text-white text-sm font-medium shadow-lg shadow-indigo-600/30 flex items-center gap-2 transition-all disabled:opacity-50"
          >
            {submitting && <Loader2 className="w-4 h-4 animate-spin" />}
            确认一键订阅并下发 qB
          </button>
        </div>
      </div>
    </div>
  );
};
