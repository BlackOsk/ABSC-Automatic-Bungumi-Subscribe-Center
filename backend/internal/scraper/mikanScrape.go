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

const MikanBaseURL = "https://mikanani.kas.pub"

// parseNumericSeason 将中文或阿拉伯数字转为标准 int
func parseNumericSeason(s string) int {
	mapping := map[string]int{
		"一": 1, "二": 2, "三": 3, "四": 4, "五": 5, "六": 6, "七": 7, "八": 8, "九": 9, "十": 10,
	}
	if val, ok := mapping[s]; ok {
		return val
	}
	if val, err := strconv.Atoi(s); err == nil {
		return val
	}
	return 1
}

// parseRomanSeason 将各种符号的罗马数字转为标准 int
func parseRomanSeason(s string) int {
	mapping := map[string]int{
		"I": 1, "II": 2, "III": 3, "IV": 4, "V": 5, "VI": 6, "VII": 7, "VIII": 8, "IX": 9, "X": 10,
		"Ⅰ": 1, "Ⅱ": 2, "Ⅲ": 3, "Ⅳ": 4, "Ⅴ": 5, "Ⅵ": 6, "Ⅶ": 7, "Ⅷ": 8, "Ⅸ": 9, "Ⅹ": 10,
	}
	if val, ok := mapping[strings.ToUpper(s)]; ok {
		return val
	}
	return 1
}

// cleanTitleAndExtractSeason 提取季数并返回剔除了季数字符串的干净标题
func cleanTitleAndExtractSeason(rawTitle string) (string, int) {
	// 1. 边界防御：将所有奇形怪状的特殊空格（全角、非换行空格）全部归一化为标准英文空格
	t := rawTitle
	t = strings.ReplaceAll(t, "\u00a0", " ")
	t = strings.ReplaceAll(t, "\u3000", " ")

	currentSeason := 1

	// 2. 策略A：匹配 "第X季"、"第X期"、"第X部分"
	reChinese := regexp.MustCompile(`第\s*([一二三四五六七八九十\d]+)\s*[季期部分]`)
	if loc := reChinese.FindStringSubmatchIndex(t); len(loc) > 0 {
		numStr := t[loc[2]:loc[3]]
		currentSeason = parseNumericSeason(numStr)
		// 挖掉这部分季数噪音
		t = t[:loc[0]] + " " + t[loc[1]:]
	}

	// 3. 策略B：匹配末尾或独立的罗马数字 (如 Clevatess II-魔兽之王 / 汪！II)
	reRoman := regexp.MustCompile(`(?i)\s+([ⅠⅡⅢⅣⅤⅥⅦⅧⅨⅩ]|[IVXLCDM]+)([\s\-]|$)`)
	if loc := reRoman.FindStringSubmatchIndex(t); len(loc) > 0 {
		romanStr := t[loc[2]:loc[3]]
		rNum := parseRomanSeason(romanStr)
		// 确认为有效罗马数字则提取并挖掉噪音
		if rNum > 1 || strings.ToUpper(romanStr) == "I" {
			currentSeason = rNum
			t = t[:loc[0]] + " " + t[loc[1]:]
		}
	}

	// 4. 收尾清理：干掉连续的多余空格及无用连接符（如 ～ -）
	t = regexp.MustCompile(`[-～~〜—]`).ReplaceAllString(t, " ")
	t = regexp.MustCompile(`\s+`).ReplaceAllString(t, " ")

	return t, currentSeason
}

func ScrapeCurrentQuarter() ([]model.BangumiMetadata, error) {
	// 1. 发送 HTTP 请求 (预留 UA 以免被CF拦截)
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", MikanBaseURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[scraper.ScrapeCurrentQuarter] 请求 Mikan 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[scraper.ScrapeCurrentQuarter] Mikan 返回非200状态码: %d", resp.StatusCode)
	}

	// 2. 加载 DOM 文档
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("[scraper.ScrapeCurrentQuarter] 解析 HTML 失败: %w", err)
	}

	var bangumis []model.BangumiMetadata
	idRegex := regexp.MustCompile(`/Home/Bangumi/(\d+)`)

	// 3. 核心逆向解析：根据 Mikan 的网格结构进行 CSS 选择
	// Mikan 首页按天排列，`.sk-bangumi` 或 `.mikan-main` 下包含每天的列
	doc.Find(".sk-bangumi").Each(func(_ int, dayList *goquery.Selection) {
		// 从数据源的html中的 data-dayofweek 提取星期
		dayStr, exists := dayList.Attr("data-dayofweek")
		var broadcastDay int
		if exists {
			var err error
			broadcastDay, err = strconv.Atoi(dayStr)
			if err != nil {
				broadcastDay = -1
			}
		} else {
			broadcastDay = -1
		}

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
				BroadcastDay: broadcastDay, // 记录更新日
			}
			bangumis = append(bangumis, bangumi)
		})
	})

	return bangumis, nil
}
