package service

import (
	"ABSC/internal/database"
	"ABSC/internal/model"
	"ABSC/internal/scraper"
	"errors"
	"log"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
)

type BangumiService struct {
	TMDBClient *scraper.TMDBClient
}

func NewBanbumiService(tmdbClient *scraper.TMDBClient) *BangumiService {
	return &BangumiService{TMDBClient: tmdbClient}
}

// SyncCurrentQuarterBangumi 核心编排函数：结合 Mikan 和 TMDB 同步当季新番
func (s *BangumiService) SyncCurrentQuarterBangumi() error {
	// 1. 从Mikan上获取当季新番的全部信息
	log.Println("[SyncCurrentQuarterBangumi] 🔄 开始同步当季新番流...")
	mikanList, err := scraper.ScrapeCurrentQuarter()
	if err != nil {
		return_err := errors.New("从 Mikan 抓取基础数据失败: " + err.Error())
		return return_err
	}
	log.Printf("[SyncCurrentQuarterBangumi] 📥 从 Mikan 成功获取到 %d 部番剧，准备进行元数据比对...", len(mikanList))

	// 2. 对比本地数据库中的信息，判断是否需要更新
	for _, bangumi := range mikanList {
		var localBangumi model.BangumiMetadata

		// 查询本地数据库是否存在MikanId(1. 不存在：新增；2. 存在但TMDBID为空：尝试刮削TMDB信息；3. 存在且TMDBID不为空：无需操作	)
		err := database.DB.Where("mikan_id = ?", bangumi.MikanID).First(&localBangumi).Error
		if err != nil {
			// 如果本地数据库中不存在该条目，则新增
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("[SyncCurrentQuarterBangumi] 🆕 新增番剧: %s (Mikan ID: %d)", bangumi.TitleCN, bangumi.MikanID)
				s.enrichTMDBInfo(&bangumi) // 尝试刮削TMDB信息并落盘
			} else {
				log.Printf("[SyncCurrentQuarterBangumi] ❌ 查询本地数据库失败: %v", err)
			}

			// 出现错误

		} else {
			// 本地数据库中存在该条目，补充TMDB信息
			if localBangumi.TMDBID == nil {
				log.Printf("[SyncCurrentQuarterBangumi] 🔄 补充番剧信息: %s (Mikan ID: %d) - 尝试刮削TMDB信息", bangumi.TitleCN, bangumi.MikanID)
				localBangumi.BroadcastDay = bangumi.BroadcastDay // 更新播出日
				s.enrichTMDBInfo(&localBangumi)                  // 尝试刮削TMDB信息并落盘
			} else {
				// 本地数据库完整，仅更新 Mikan 侧可能变化的字段（如同播时间）
				database.DB.Model(&localBangumi).Updates(map[string]interface{}{
					"broadcast_day": bangumi.BroadcastDay,
					"title_cn":      bangumi.TitleCN,
				})
			}
		}
	}
	log.Print("[SyncCurrentQuarterBangumi] ✅ 当季新番同步完成！")
	return nil

}

// enrichTMDBInfo 辅助函数：负责单个番剧的 TMDB 刮削与信息补充，补充完后实现落盘
func (s *BangumiService) enrichTMDBInfo(b *model.BangumiMetadata) {

	time.Sleep(1 * time.Second)

	var tmdbResult *scraper.TMDBResult
	var err error

	cleanTitle := regexp.MustCompile(`\s+`).ReplaceAllString(b.TitleCN, " ")
	words := strings.Fields(cleanTitle)

	successQueryTitle := ""

	// 从全长数组开始，如果搜不到，就砍掉最后一个元素，直到只剩第一个单词
	for i := len(words); i > 0; i-- {
		queryTitle := strings.Join(words[:i], " ")

		// 再次清洗边缘可能因裁剪暴露出的特殊标点
		queryTitle = strings.Trim(queryTitle, " -～~〜—")
		if queryTitle == "" {
			continue
		}

		log.Printf("[enrichTMDBInfo] 尝试使用检索词 [%s] 请求 TMDB...", queryTitle)
		tmdbResult, err = s.TMDBClient.SearchAnime(queryTitle)

		if err == nil && tmdbResult != nil {
			// 匹配成功！记录标题
			successQueryTitle = queryTitle
			break
		}
		log.Printf("[enrichTMDBInfo] 检索词 [%s] 未命中 TMDB，准备剥离末尾副标题并退级重试...", queryTitle)
	}

	// 最终降级兜底：如果一轮失败了
	if successQueryTitle == "" || tmdbResult == nil {
		log.Printf("[enrichTMDBInfo] [%s] 在经历全轮次副标题剥离后依然无法在 TMDB 匹配，仅保留 Mikan 基础数据", b.TitleCN)
		if err := database.DB.Save(b).Error; err != nil {
			log.Printf("[enrichTMDBInfo] [%s] 的Mikan信息落盘失败: %v", b.TitleCN, err)
		}
		return
	}

	// // 获取 TMDB 信息
	// tmdbResult, err = s.TMDBClient.SearchAnime(b.TitleCN)
	// // 如果 TMDB 刮削失败，返回错误
	// if err != nil {
	// 	log.Printf("[enrichTMDBInfo]%s 条目的信息在TMDB上刮削失败（跳过）：%v", b.TitleCN, err)
	// 	// 在没有TMDB信息的情况下，将Mikan的基础信息落盘
	// 	if err := database.DB.Save(b).Error; err != nil {
	// 		log.Printf("[enrichTMDBInfo] ❌  %s 的Mikan信息落盘失败: %v", b.TitleCN, err)
	// 	}

	// 	return

	// }

	// 用检索成功的标题覆盖重写 TitleCN
	b.TitleCN = successQueryTitle

	// 成功获取到TMDB信息，进行信息补充
	b.TMDBID = &tmdbResult.ID
	b.PosterPath = tmdbResult.PosterPath
	b.Overview = tmdbResult.Overview
	b.AirDate = tmdbResult.FirstAirDate

	// 将补充后的信息落盘
	if err := database.DB.Save(b).Error; err != nil {
		log.Printf("[enrichTMDBInfo] ❌  %s 的混合信息落盘失败: %v", b.TitleCN, err)
		return
	} else {
		log.Printf("[enrichTMDBInfo] ✅  %s (TMDB ID: %d ) 的混合信息落盘成功", b.TitleCN, *b.TMDBID)

	}

}

func (s *BangumiService) CalculateAutoOffset(tmdbID int, targetSeason int) int {
	if targetSeason <= 1 {
		return 0
	}

	// 1. 获取该番剧的所有季信息
	tmdbDetails, err := s.TMDBClient.GetTVDetails(tmdbID)
	if err != nil {
		log.Printf("[CalculateAutoOffset] TMDB ID %d 获取 TMDB 详情失败: %v", tmdbID, err)
		return 0
	}

	autoOffset := 0

	// 2. 遍历所有季，计算目标季之前的总集数
	for _, season := range tmdbDetails.Seasons {
		if season.SeasonNumber > 0 && season.SeasonNumber < targetSeason {
			autoOffset += season.EpisodeCount
			log.Printf("动漫 TMDB_ID: %d; 自动推导累加：第 %d 季有 %d 集", tmdbID, season.SeasonNumber, season.EpisodeCount)
		}

	}
	log.Printf("动漫 TMDB_ID: %d，当前订阅第 %d 季，推导出的集数偏移量为: 减去 %d 集", tmdbID, targetSeason, autoOffset)

	return autoOffset

}
