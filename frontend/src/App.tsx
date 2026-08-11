import { useEffect, useState } from "react";
import type { BangumiItem } from "./types/anime";
import { AnimeAPI } from "./api/client";
import { Navbar } from "./components/Navbar";
import { BangumiCard } from "./components/BangumiCard";
import { SubscribeModal } from "./components/SubscribeModal";
import { Loader2 } from "lucide-react";

export default function App() {
  const [activeTab, setActiveTab] = useState<"all" | "subscribed" | "offsets">(
    "all",
  );
  const [bangumis, setBangumis] = useState<BangumiItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [isSyncing, setIsSyncing] = useState(false);
  const [selectedBangumi, setSelectedBangumi] = useState<BangumiItem | null>(
    null,
  );

  // 1. 组件首次挂载时异步加载数据 (避免同步 setState 触发 Cascading Render)
  useEffect(() => {
    let isMounted = true;

    AnimeAPI.getCurrentBangumis()
      .then((data) => {
        if (isMounted) {
          setBangumis(data);
        }
      })
      .catch((err: unknown) => {
        if (isMounted) {
          const msg = err instanceof Error ? err.message : "获取番剧列表失败";
          console.error(msg);
        }
      })
      .finally(() => {
        if (isMounted) {
          setLoading(false);
        }
      });

    return () => {
      isMounted = false;
    };
  }, []);

  // 2. 供下架、订阅成功或手动同步后调用的静默/无感知刷新函数
  const reloadBangumis = async () => {
    try {
      const data = await AnimeAPI.getCurrentBangumis();
      setBangumis(data);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "刷新番剧列表失败";
      console.error(msg);
    }
  };

  // 3. 手动触发增量同步
  const handleSync = async () => {
    setIsSyncing(true);
    try {
      await AnimeAPI.syncBangumis();
      setTimeout(reloadBangumis, 3000); // 3秒后刷新视图
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : "未知错误";
      alert("同步触发失败: " + errorMessage);
    } finally {
      setIsSyncing(false);
    }
  };

  // 4. 彻底下架处理
  const handlePurge = async (bangumi: BangumiItem) => {
    if (
      confirm(
        `确定要彻底下架《${bangumi.title_cn}》吗？这将清理 qBittorrent 规则及对应的物理文件。`,
      )
    ) {
      try {
        await AnimeAPI.purgeSubscription(bangumi.mikan_id, true);
        reloadBangumis();
      } catch (err: unknown) {
        const errorMessage = err instanceof Error ? err.message : "未知错误";
        alert("下架失败: " + errorMessage);
      }
    }
  };

  // 过滤展示列表
  const displayedBangumis = bangumis.filter((b) => {
    if (activeTab === "subscribed") return b.sub_status === "subscribing";
    return true;
  });

  return (
    <div className="min-h-screen flex flex-col bg-slate-900 text-slate-100 font-sans antialiased">
      {/* 顶部导航 */}
      <Navbar
        activeTab={activeTab}
        setActiveTab={setActiveTab}
        onSync={handleSync}
        isSyncing={isSyncing}
      />

      {/* 主界面海报墙区域 */}
      <main className="flex-1 max-w-7xl w-full mx-auto px-6 py-8">
        {loading ? (
          <div className="py-24 flex flex-col items-center justify-center text-slate-400 gap-3">
            <Loader2 className="w-10 h-10 animate-spin text-indigo-500" />
            <p className="text-sm font-medium">
              从后端 SQLite 获取当季新番数据中...
            </p>
          </div>
        ) : displayedBangumis.length === 0 ? (
          <div className="py-24 text-center text-slate-500">
            <p className="text-lg font-medium">暂无符合条件的番剧数据</p>
          </div>
        ) : (
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-6">
            {displayedBangumis.map((bangumi) => (
              <BangumiCard
                key={bangumi.mikan_id}
                bangumi={bangumi}
                onSubscribeClick={setSelectedBangumi}
                onPurgeClick={handlePurge}
              />
            ))}
          </div>
        )}
      </main>

      {/* 订阅弹窗 Modal */}
      {selectedBangumi && (
        <SubscribeModal
          bangumi={selectedBangumi}
          onClose={() => setSelectedBangumi(null)}
          onSuccess={reloadBangumis}
        />
      )}
    </div>
  );
}
