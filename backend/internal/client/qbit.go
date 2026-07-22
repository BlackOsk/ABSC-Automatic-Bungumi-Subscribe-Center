package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

// TorrentParams 映射 qB RSS 规则中的下载参数
type TorrentParams struct {
	Category                 string   `json:"category"`
	DownloadPath             string   `json:"download_path"`
	InactiveSeedingTimeLimit int      `json:"inactive_seeding_time_limit"`
	OperatingMode            string   `json:"operating_mode"` // 默认 "AutoManaged"
	RatioLimit               int      `json:"ratio_limit"`
	SavePath                 string   `json:"save_path"`
	SeedingTimeLimit         int      `json:"seeding_time_limit"`
	SkipChecking             bool     `json:"skip_checking"`
	Tags                     []string `json:"tags"`
	UploadLimit              int      `json:"upload_limit"`
	Stopped                  bool     `json:"stopped"`
}

// RuleDefinition 映射 qB 完整的 RSS 自动化过滤器规则
type RuleDefinition struct {
	Enabled                   bool          `json:"enabled"`
	MustContain               string        `json:"mustContain"`
	MustNotContain            string        `json:"mustNotContain"`
	UseRegex                  bool          `json:"useRegex"`
	EpisodeFilter             string        `json:"episodeFilter"`
	SmartFilter               bool          `json:"smartFilter"`
	PreviouslyMatchedEpisodes []string      `json:"previouslyMatchedEpisodes"`
	AffectedFeeds             []string      `json:"affectedFeeds"` // 绑定关联的 RSS 链接
	IgnoreDays                int           `json:"ignoreDays"`
	LastMatch                 string        `json:"lastMatch"`
	AssignedCategory          string        `json:"assignedCategory"`
	Priority                  int           `json:"priority"`
	SavePath                  string        `json:"savePath"`
	TorrentContentLayout      interface{}   `json:"torrentContentLayout"` // 可为 nil
	TorrentParams             TorrentParams `json:"torrentParams"`
}

// TorrentInfo 映射 qB /api/v2/torrents/info 返回的种子元数据
type TorrentInfo struct {
	Hash        string  `json:"hash"`
	Name        string  `json:"name"`
	SavePath    string  `json:"save_path"`
	ContentPath string  `json:"content_path"`
	Category    string  `json:"category"`
	Progress    float64 `json:"progress"`
}

// QBitClient 封装 qB API 交互核心句柄
type QBitClient struct {
	BaseURL    string
	Username   string
	Password   string
	HTTPClient *http.Client
}

// NewQBitClient 创建并初始化一个自动管理 Cookie 的 qB 客户端
func NewQBitClient(baseURL, username, password string) (*QBitClient, error) {
	// 创建一个 CookieJar，HTTPClient 随后会自动保存并携带登录后的 SID
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("[NewQBitClient] 初始化 CookieJar 失败: %w", err)
	}

	return &QBitClient{
		BaseURL:  strings.TrimSuffix(baseURL, "/"),
		Username: username,
		Password: password,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
			Jar:     jar,
		},
	}, nil
}

// postForm 内部辅助函数：处理标准的 x-www-form-urlencoded POST 请求
func (c *QBitClient) postForm(apiPath string, data url.Values) (*http.Response, error) {
	apiURL := fmt.Sprintf("%s/api/v2/%s", c.BaseURL, apiPath)

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[postForm] qB API [%s] 返回异常状态码: %d", apiPath, resp.StatusCode)
	}

	return resp, nil
}

// Login 登录并激活会话
func (c *QBitClient) Login() error {
	data := url.Values{}
	data.Set("username", c.Username)
	data.Set("password", c.Password)

	// 调用 auth/login 接口
	_, err := c.postForm("auth/login", data)
	if err != nil {
		return fmt.Errorf("[Login] qB 登录失败: %w", err)
	}
	return nil
}

// CreateCategory 创建以番剧命名的独立分类，并绑定 NAS 物理存储路径
func (c *QBitClient) CreateCategory(category, savePath string) error {
	data := url.Values{}
	data.Set("category", category)
	data.Set("savePath", savePath)

	_, err := c.postForm("torrents/createCategory", data)
	if err != nil {
		return fmt.Errorf("[CreateCategory] 创建分类失败: %w", err)
	}
	return nil
}

// AddRSSFeed RSS 链接添加到 qB
func (c *QBitClient) AddRSSFeed(feedURL, feedName string) error {
	data := url.Values{}
	data.Set("url", feedURL)
	data.Set("path", feedName) // 在 qB 内部展现的 RSS 节点名称

	_, err := c.postForm("rss/addFeed", data)
	if err != nil {
		return fmt.Errorf("[AddRSSFeed] 添加 RSS 订阅源失败 [%s]: %w", feedName, err)
	}
	return nil
}

// SetRSSRule 配置智能下载过滤器规则，关联对应的分类与存储路径
func (c *QBitClient) SetRSSRule(ruleName string, ruleDef RuleDefinition) error {
	// 将强类型结构体直接序列化为标准的 JSON 字符串
	jsonBytes, err := json.Marshal(ruleDef)
	if err != nil {
		return fmt.Errorf("[SetRSSRule] 序列化 RuleDefinition 失败: %w", err)
	}

	data := url.Values{}
	data.Set("ruleName", ruleName)
	data.Set("ruleDef", string(jsonBytes))

	_, err = c.postForm("rss/setRule", data)
	if err != nil {
		return fmt.Errorf("[SetRSSRule] 设置 RSS 规则失败 [%s]: %w", ruleName, err)
	}
	return nil
}

// GetTorrents 获取最近的种子列表
// func (c *QBitClient) GetTorrents(limit int) ([]TorrentInfo, error) {
// 	data := url.Values{}
// 	data.Set("filter", "all")
// 	data.Set("sort", "added_on")
// 	data.Set("reverse", "true")
// 	data.Set("limit", fmt.Sprintf("%d", limit))

// }
