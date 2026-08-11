import React from "react";
import { Tv, BookmarkCheck, RefreshCw } from "lucide-react";

interface NavbarProps {
  activeTab: "all" | "subscribed" | "offsets";
  setActiveTab: (tab: "all" | "subscribed" | "offsets") => void;
  onSync: () => void;
  isSyncing: boolean;
}

export const Navbar: React.FC<NavbarProps> = ({
  activeTab,
  setActiveTab,
  onSync,
  isSyncing,
}) => {
  return (
    <header className="sticky top-0 z-30 bg-slate-900/80 backdrop-blur-md border-b border-slate-800 px-6 py-4">
      <div className="max-w-7xl mx-auto flex items-center justify-between">
        {/* LOGO */}
        <div className="flex items-center gap-3">
          <div className="p-2 bg-indigo-600 rounded-xl shadow-lg shadow-indigo-500/30">
            <Tv className="w-6 h-6 text-white" />
          </div>
          <div>
            <h1 className="text-xl font-bold bg-gradient-to-r from-white via-slate-200 to-indigo-300 bg-clip-text text-transparent">
              Anime Manager
            </h1>
            <p className="text-xs text-slate-400">自动化追番控制中心</p>
          </div>
        </div>

        {/* 标签页切换 */}
        <nav className="flex items-center bg-slate-800/80 p-1 rounded-xl border border-slate-700/50">
          <button
            onClick={() => setActiveTab("all")}
            className={`flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-lg transition-all ${
              activeTab === "all"
                ? "bg-indigo-600 text-white shadow-md"
                : "text-slate-400 hover:text-white hover:bg-slate-700/50"
            }`}
          >
            <Tv className="w-4 h-4" />
            当季新番
          </button>
          <button
            onClick={() => setActiveTab("subscribed")}
            className={`flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-lg transition-all ${
              activeTab === "subscribed"
                ? "bg-indigo-600 text-white shadow-md"
                : "text-slate-400 hover:text-white hover:bg-slate-700/50"
            }`}
          >
            <BookmarkCheck className="w-4 h-4" />
            我的订阅
          </button>
        </nav>

        {/* 右侧动作按钮 */}
        <button
          onClick={onSync}
          disabled={isSyncing}
          className="flex items-center gap-2 px-4 py-2 text-sm font-medium bg-slate-800 hover:bg-slate-700 active:bg-slate-600 text-slate-200 border border-slate-700 rounded-xl transition-all disabled:opacity-50"
        >
          <RefreshCw
            className={`w-4 h-4 ${isSyncing ? "animate-spin text-indigo-400" : ""}`}
          />
          {isSyncing ? "同步中..." : "刷新新番数据"}
        </button>
      </div>
    </header>
  );
};
