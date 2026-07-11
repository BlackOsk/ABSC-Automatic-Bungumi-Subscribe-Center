package scraper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const TMDBBaseURL = "https://api.themoviedb.org/3"

// TMDBClient 封装了请求 TMDB 所需的凭证与网络客户端
type TMDBClient struct {
	APIKey string
	Client *http.Client
}

// TMDBResult 映射 TMDB 搜索返回的单条剧集元数据
type TMDBResult struct {
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	OriginalName  string   `json:"original_name"`
	Overview      string   `json:"overview"`
	PosterPath    string   `json:"poster_path"`
	FirstAirDate  string   `json:"first_air_date"`
	OriginCountry []string `json:"origin_country"`
}

// TMDBSearchResponse 映射 TMDB 搜索接口的完整回包
type TMDBSearchResponse struct {
	Page         int          `json:"page"`
	Results      []TMDBResult `json:"results"`
	TotalPages   int          `json:"total_pages"`
	TotalResults int          `json:"total_results"`
}

// NewTMDBClient 初始化客户端，支持传入可选的 HTTP 代理
func NewTMDBClient(apiKey string, proxyURL string) *TMDBClient {
	transport := &http.Transport{}

	// 如果配置了代理（如软路由内网代理），则让请求走代理
	if proxyURL != "" {
		if pURL, err := url.Parse(proxyURL); err == nil {
			transport.Proxy = http.ProxyURL(pURL)
		}
	}

	return &TMDBClient{
		APIKey: apiKey,
		Client: &http.Client{
			Timeout:   10 * time.Second,
			Transport: transport,
		},
	}
}

// SearchAnime 传入动漫名称，寻找最佳匹配的 TMDB 元数据
func (c *TMDBClient) SearchAnime(title string) (*TMDBResult, error) {
	// 1. 构建请求 URL
	searchURL := fmt.Sprintf("%s/search/tv", TMDBBaseURL)
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	// 2. 携带必备参数 (必须指定 language=zh-CN 以获取中文简介)
	q := req.URL.Query()
	q.Add("api_key", c.APIKey)
	q.Add("query", title)
	q.Add("language", "zh-CN")
	q.Add("include_adult", "true")
	req.URL.RawQuery = q.Encode()

	req.Header.Add("accept", "application/json")

	// 3. 发起请求
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 TMDB 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDB 返回异常状态码: %d", resp.StatusCode)
	}

	// 4. 解析回包
	var searchResp TMDBSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("解析 TMDB JSON 失败: %w", err)
	}

	if len(searchResp.Results) == 0 {
		return nil, fmt.Errorf("未能在 TMDB 找到相关动漫: %s", title)
	}

	// 5. 精准筛选策略：
	// 优先挑选来自日本 (JP) 的剧集，防止被同名国产剧或欧美剧污染
	for _, result := range searchResp.Results {
		for _, country := range result.OriginCountry {
			if country == "JP" {
				return &result, nil
			}
		}
	}

	// 如果没有 JP 标签，默认信任第一个搜索结果
	return &searchResp.Results[0], nil
}
