package service

import (
	"ABSC/internal/database"
	"ABSC/internal/model"
	"ABSC/internal/scraper"
	"errors"
	"log"

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
	log.Println("🔄 开始同步当季新番流...")
	mikanList, err := scraper.ScrapeCurrentQuarter()
	if err != nil {
		return_err := errors.New("从 Mikan 抓取基础数据失败: " + err.Error())
		return return_err
	}
	log.Printf("📥 从 Mikan 成功获取到 %d 部番剧，准备进行元数据比对...", len(mikanList))
	// 2. 对比本地数据库中的信息，判断是否需要更新
	for _, bangumi := range mikanList {
		var localBangumi model.BangumiMetadata

		// 查询本地数据库是否存在MikanId
		err := database.DB.Where("mikan_id = ?", bangumi.MikanID).First(&localBangumi).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {

			}
			// 出现错误

		}

	}

}

// enrichTMDBInfo 辅助函数：负责单个番剧的 TMDB 刮削与信息补充
func (s *BangumiService) enrichTMDBInfo(bangumi *model.BangumiMetadata) error {

}
