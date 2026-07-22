package scraper

import "testing"

func TestScrapeBangumiDetail(t *testing.T) {
	// 选用一个经典的番剧 MikanID 进行实网解析验证 (例如：3241)
	targetMikanID := 3937

	t.Logf("🌐 正在对蜜柑计划番剧详情页进行深度解析测试 (ID: %d)...", targetMikanID)
	resources, err := ScrapeBangumiDetail(targetMikanID)
	if err != nil {
		t.Fatalf("❌ 详情页深度解析彻底失败: %v", err)
	}

	if len(resources) == 0 {
		t.Fatal("❌ 未能抓取到任何字幕组资源，请检查 Mikan 详情页布局结构是否发生突变！")
	}

	t.Logf("🎉 成功解析出 %d 个参与本番剧翻译的字幕组分栏：", len(resources))

	for _, res := range resources {
		t.Logf("--------------------------------------------------")
		t.Logf("🏷️ 字幕组: %s (ID: %d)", res.SubgroupName, res.SubgroupID)
		t.Logf("📡 专属 RSS 地址: %s", res.RSSURL)
		t.Logf("📦 详情页当前展示的历史更新文件数: %d 个", len(res.Episodes))

		if len(res.Episodes) > 0 {
			firstEp := res.Episodes[0]
			t.Logf("  📂 首条更新实录: %s", firstEp.Title)
			t.Logf("  ⚖️ 文件大小: %s  |  ⏰ 发布时间: %s", firstEp.Size, firstEp.PublishTime)
			t.Logf("  🧲 磁力链前缀: %.40s...", firstEp.Magnet)

			if firstEp.Magnet == "" {
				t.Errorf("⚠️ 警告：发现有文件未能成功捕获到 Magnet 磁力链接！")
			}
		} else {
			t.Errorf("❌ 错误：字幕组 [%s] 的更新文件表格内容为空！", res.SubgroupName)
		}
	}
}
