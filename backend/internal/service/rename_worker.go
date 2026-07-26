package service

import (
	"ABSC/internal/client"
	"ABSC/internal/model"
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type RenameService struct {
	QBitClient      *client.QBitClient
	SeriesDirectory string // 如 /downloads/Series/
	IncompleteDir   string // 如 /downloads/incomplete
}

func NewRenameService(qbClient *client.QBitClient, seriesDir, incompleteDir string) *RenameService {
	return &RenameService{
		QBitClient:      qbClient,
		SeriesDirectory: strings.TrimSuffix(seriesDir, "/") + "/",
		IncompleteDir:   incompleteDir,
	}
}

// 预编译正则模式列表
var episodePatterns = []*regexp.Regexp{
	regexp.MustCompile(`\s([0-9]{2})\s`),
	regexp.MustCompile(`-([0-9]{2})-`),
	regexp.MustCompile(`\[([0-9]{2})\]`),
	regexp.MustCompile(`\[([0-9]{2})v[0-9]\]`),
	regexp.MustCompile(`\s([0-9]{2})v[0-9]\s`),
	regexp.MustCompile(`(?i)E([0-9]{2})`), // 容错支持 E01 格式
}

// ExtractAbsoluteEpisode 从原始长文件名中提取绝对集数 (如从 "... - 13 ..." 中提取出 13)
func ExtractAbsoluteEpisode(fileName string) (bool, int) {
	for _, pattern := range episodePatterns {
		matches := pattern.FindStringSubmatch(fileName)
		if len(matches) >= 2 {
			absEp, err := strconv.Atoi(matches[1])
			if err == nil {
				return true, absEp
			}
		}
	}
	return false, 0
}

// // ExtractAndCalculateEpisode 从原始长文件名中提取集数，并根据 offset 计算相对集数
// func ExtractAndCalculateEpisode(fileName string, offset int) (bool, string) {
// 	for _, pattern := range episodePatterns {
// 		matches := pattern.FindStringSubmatch(fileName)
// 		if len(matches) >= 2 {
// 			// matches[1] 抓取到的是纯数字集数（如 "13"）
// 			absEp, err := strconv.Atoi(matches[1])
// 			if err != nil {
// 				continue
// 			}

// 			// 计算相对集数：绝对集数 - 偏移量
// 			relEp := absEp - offset
// 			if relEp <= 0 {
// 				relEp = absEp // 保护机制：如果减完<=0，退回到绝对集数
// 			}

// 			// 格式化输出为 E01, E02 等标准格式
// 			return true, fmt.Sprintf("E%02d", relEp)
// 		}
// 	}
// 	return false, ""
// }

// DetermineTargetSeasonAndOffset 根据 SQLite 中的多季偏移配置，判断绝对集数归属于哪一季，并计算相对集数和目标路径
func DetermineTargetSeasonAndOffset(absEp int, offsets []model.EpisodeOffset) (int, int) {
	// 如果没有配置偏移，默认归属为 Season 1，相对集数 = 绝对集数
	if len(offsets) == 0 {
		return 1, absEp
	}

	// 按 Season 从大到小降序排列偏移配置，做降级匹配
	sort.Slice(offsets, func(i, j int) bool {
		return offsets[i].Season > offsets[j].Season
	})

	for _, off := range offsets {
		// 只要绝对集数大于该季度的偏移基准值，就说明命中该季度
		// 例如：绝对集数 13，Season 2 的 offset 为 11 (13 > 11)，命中 Season 2！
		if absEp > off.OffsetValue {
			relEp := absEp - off.OffsetValue
			return off.Season, relEp
		}
	}

	return 1, absEp
}

// ExecuteRenameTask 扫描最近下载完成的种子并实施规范化重命名
func (s *RenameService) ExecuteRenameTask(checkLimit int) error {
	log.Println("启动重命名")

	// 1. 调用 qb API 获取最近的种子列表
	torrents, err := s.QBitClient.GetTorrents(checkLimit)
	if err != nil {
		return fmt.Errorf("获取种子列表失败: %w", err)
	}

	for _, t := range torrents {
		cleanSavePath := filepath.Clean(t.SavePath)

		// 判断未完成的torrents，未完成不改名
		if strings.Contains(t.SavePath, s.IncompleteDir) {
			continue

		}
		// 只处理存放在剧集根目录下的torrents
		if !strings.Contains(t.SavePath, s.SeriesDirectory) {
			continue
		}

		// 从路径中解析 relative 相对文件名
		oldFileName := strings.TrimPrefix(t.SavePath, s.SeriesDirectory)

		// 提取文件名中的绝对集数
		matched, absEp := ExtractAbsoluteEpisode(oldFileName)
		if !matched {
			log.Printf("[ExtractAbsoluteEpisode] 未能在文件名中匹配到标准集数(跳过): %s", oldFileName)
			continue
		}

		// 从存储路径中解析当前分类对应的剧名
		relSavePath := strings.TrimPrefix(cleanSavePath, s.SeriesDirectory)
		relSavePath = strings.TrimSuffix(relSavePath, "/")
		parts := strings.Split(relSavePath, "/")
		if len(parts) == 0 || parts[0] == "" {
			continue
		}
		titleCN := parts[0]

	}

}
