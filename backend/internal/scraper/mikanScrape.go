package scraper

import (
	"ABSC/internal/model"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const MikanBaseURL = "https://mikanani.me"

func ScrapeCurrentQuarter() ([]model.BangumiMetadata, error) {
	// 1. 发送 HTTP 请求 (预留 UA 以免被CF拦截)
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", MikanBaseURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 Mikan 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Mikan 返回非200状态码: %d", resp.StatusCode)
	}

	// 2. 加载 DOM 文档
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("解析 HTML 失败: %w", err)
	}

	var bangumis []model.BangumiMetadata
	idRegex := regexp.MustCompile(`/Home/Bangumi/(\d+)`)

	// 3. 核心逆向解析：根据 Mikan 的网格结构进行 CSS 选择
	// Mikan 首页按天排列，`.sk-bangumi` 或 `.mikan-main` 下包含每天的列
	doc.Find(".sk-bangumi").Each(func(dayIdx int, dayList *goquery.Selection) {
		// dayIdx 0-6 刚好代表周日到周六（需根据 Mikan 实际展示位置微调对应映射）

		dayList.Find(".an-text").Each(func(_ int, s *goquery.Selection) {
			title := strings.TrimSpace(s.Text())
			href, exists := s.Attr("href")
			if !exists {
				return
			}

			// 从 href="/Home/Bangumi/3241" 中提取 Mikan ID
			matches := idRegex.FindStringSubmatch(href)
			if len(matches) < 2 {
				return
			}
			mikanID, _ := strconv.Atoi(matches[1])

			bangumi := model.BangumiMetadata{
				MikanID:      mikanID,
				TitleCN:      title,
				BroadcastDay: dayIdx, // 记录更新周期
			}
			bangumis = append(bangumis, bangumi)
		})
	})

	return bangumis, nil
}
