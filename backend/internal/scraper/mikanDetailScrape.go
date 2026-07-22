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

// ScrapeBangumiDetail 深度爬取并解析单部番剧详情页中的【字幕组RSS】与【文件列表】
func ScrapeBangumiDetail(mikanID int) ([]model.MikanSubgroupResource, error) {
	detailURL := fmt.Sprintf("%s/Home/Bangumi/%d", MikanBaseURL, mikanID)

	// 1. 发送 HTTP 请求 (预留 UA 以免被CF拦截)
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", detailURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[ScrapeBangumiDetail] 请求蜜柑详情页失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[ScrapeBangumiDetail] 蜜柑详情页返回异常状态码: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("[ScrapeBangumiDetail] 解析详情页 HTML 失败: %w", err)
	}

	var resources []model.MikanSubgroupResource
	subgroupIDRegex := regexp.MustCompile(`(?i)[?&]subgroupid=(\d+)`)

	doc.Find(".central-container > .subgroup-text").Each(func(_ int, s *goquery.Selection) {
		// 获取字幕组的ID
		idStr, exist := s.Attr("id")
		subgroupID := 0
		if exist {
			subgroupID, _ = strconv.Atoi(idStr)
		}

		// 获取字幕组的名字
		subgroupName := ""
		s.Find("a").Each(func(_ int, a *goquery.Selection) {
			href, _ := a.Attr("href")
			if strings.Contains(href, "/Home/PublishGroup/") {
				subgroupName = strings.TrimSpace(a.Text())
			}
		})
		// 名字未获取到时
		if subgroupName == "" {
			subgroupName = strings.TrimSpace(s.Find("a").Not(".mikan-rss").First().Text())
		}
		if subgroupName == "" {
			subgroupName = "未知字幕组"
		}
		// 提取该字幕组针对该番剧专用的 RSS 订阅链接
		rssBtn := s.Find("a.mikan-rss")
		rssHref, rssExist := rssBtn.Attr("href")
		rssURL := ""
		if rssExist {
			rssURL = MikanBaseURL + rssHref
		}
		// 如果由于特殊情况 id 属性没抓到，利用正则从 RSS url 重新提取 SubgroupID 容错
		if subgroupID == 0 {
			if matches := subgroupIDRegex.FindStringSubmatch(rssHref); len(matches) >= 2 {
				subgroupID, _ = strconv.Atoi(matches[1])
			}
		}
		resource := model.MikanSubgroupResource{
			MikanID:      mikanID,
			SubgroupID:   subgroupID,
			SubgroupName: subgroupName,
			RSSURL:       rssURL,
			Episodes:     []model.MikanEpisode{}, // 先初始化为空，后续可通过 RSS 订阅源获取
		}

		// 利用平级遍历，寻找当前 .subgroup-text 后面紧跟的平级 .episode-table
		episodeTable := s.NextAll().Filter(".episode-table").First()

		// 解析表格内部的具体单集资源
		episodeTable.Find("table tbody tr").Each(func(_ int, tr *goquery.Selection) {
			tds := tr.Find("td")
			if tds.Length() < 4 {
				return // 确保有足够的列
			}
			// 提取动漫长文件名 (在第二列 td[1] 内部的 .magnet-link-wrap 标签中)
			titleLink := tds.Eq(1).Find("a.magnet-link-wrap")
			title := strings.TrimSpace(titleLink.Text())
			if title == "" {
				return
			}

			// 提取文件大小 (在第三列 td[2])
			size := strings.TrimSpace(tds.Eq(2).Text())

			// 提取发布时间 (在第四列 td[3])
			publishTime := strings.TrimSpace(tds.Eq(3).Text())

			// 精准提取磁力链接
			magnet := ""
			// 优先从第一列 td[0] 的复选框绑定的 data-magnet 属性读取，100% 完整纯净
			if checkbox := tds.Eq(0).Find("input.js-episode-select"); checkbox.Length() > 0 {
				magnet, _ = checkbox.Attr("data-magnet")
			}
			// 如果没有，则从第二列 td[1] 的 a.magnet-link-wrap 标签的 href 属性中提取
			if magnet == "" {
				if copuBtn := tds.Eq(1).Find("a.magnet-link-wrap"); copuBtn.Length() > 0 {
					magnet, _ = copuBtn.Attr("href")
				}

			}

			episode := model.MikanEpisode{
				MikanID:     mikanID,
				SubgroupID:  subgroupID,
				Title:       title,
				Size:        size,
				PublishTime: publishTime,
				Magnet:      magnet,
			}
			resource.Episodes = append(resource.Episodes, episode)

			// 过滤无意义噪音数据：只有当拥有有效 RSS 或提取到了更新文件时，才计入最终列表
			if resource.RSSURL != "" || len(resource.Episodes) > 0 {
				resources = append(resources, resource)
			}
		})

	})
	return resources, nil

}
