package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
)

const yahooSearchURL = "https://shopping.yahooapis.jp/ShoppingWebService/V3/itemSearch"

type RakutenItem struct {
	ItemCode      string  `json:"item_code"`
	Name          string  `json:"name"`
	Genre         string  `json:"genre"`
	ImageURL      string  `json:"image_url"`
	ContentVolume float64 `json:"content_volume"`
	ContentUnit   string  `json:"content_unit"`
}

type RakutenSearchResult struct {
	Items      []*RakutenItem `json:"items"`
	TotalCount int            `json:"total_count"`
	Page       int            `json:"page"`
	PageCount  int            `json:"page_count"`
}

type RakutenService struct {
	appID  string
	client *http.Client
}

func NewRakutenService(appID string) *RakutenService {
	return &RakutenService{
		appID:  appID,
		client: &http.Client{},
	}
}

func (s *RakutenService) Search(keyword string, page, hits int) (*RakutenSearchResult, error) {
	if s.appID == "" {
		return s.mockResult(keyword, page, hits), nil
	}

	params := url.Values{}
	params.Set("appid", s.appID)
	params.Set("query", keyword)
	params.Set("results", strconv.Itoa(hits))
	params.Set("page", strconv.Itoa(page))
	params.Set("sort", "-score")

	reqURL := yahooSearchURL + "?" + params.Encode()
	log.Printf("[yahoo] GET %s", reqURL)

	resp, err := s.client.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("yahoo search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[yahoo] error status %d: %s — falling back to mock", resp.StatusCode, string(body))
		return s.mockResult(keyword, page, hits), nil
	}

	var raw struct {
		TotalResultsAvailable int `json:"totalResultsAvailable"`
		TotalResultsReturned  int `json:"totalResultsReturned"`
		FirstResultsPosition  int `json:"firstResultsPosition"`
		Hits                  []struct {
			Name  string `json:"name"`
			Code  string `json:"code"`
			Image struct {
				Medium string `json:"medium"`
			} `json:"image"`
			GenreCategory struct {
				Name string `json:"name"`
			} `json:"genreCategory"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode yahoo: %w", err)
	}

	pageCount := 1
	if hits > 0 {
		pageCount = (raw.TotalResultsAvailable + hits - 1) / hits
	}

	result := &RakutenSearchResult{
		TotalCount: raw.TotalResultsAvailable,
		Page:       page,
		PageCount:  pageCount,
	}
	for _, it := range raw.Hits {
		vol, unit := extractVolume(it.Name)
		result.Items = append(result.Items, &RakutenItem{
			ItemCode:      it.Code,
			Name:          it.Name,
			Genre:         it.GenreCategory.Name,
			ImageURL:      it.Image.Medium,
			ContentVolume: vol,
			ContentUnit:   unit,
		})
	}
	return result, nil
}

var volumeRe = regexp.MustCompile(`(\d+(?:\.\d+)?)\s*(ml|mL|ｍｌ|ℓ|L|l|g|ｇ|kg|ｋｇ|枚|ロール|個|本|袋|包|sheets)`)

func extractVolume(name string) (float64, string) {
	unitMap := map[string]string{
		"ml": "ml", "mL": "ml", "ｍｌ": "ml",
		"ℓ": "L", "L": "L", "l": "L",
		"g": "g", "ｇ": "g",
		"kg": "kg", "ｋｇ": "kg",
		"枚": "枚", "ロール": "ロール", "個": "個",
		"本": "本", "袋": "袋", "包": "包", "sheets": "枚",
	}
	m := volumeRe.FindStringSubmatch(name)
	if m == nil {
		return 0, ""
	}
	vol, _ := strconv.ParseFloat(m[1], 64)
	unit := unitMap[m[2]]
	if unit == "" {
		unit = m[2]
	}
	return vol, unit
}

func (s *RakutenService) SearchByJAN(janCode string) (*RakutenSearchResult, error) {
	return s.Search(janCode, 1, 10)
}

func (s *RakutenService) mockResult(keyword string, page, hits int) *RakutenSearchResult {
	var items []*RakutenItem
	for i := 1; i <= hits; i++ {
		items = append(items, &RakutenItem{
			ItemCode:      fmt.Sprintf("MOCK-%s-%d", keyword, i),
			Name:          fmt.Sprintf("%s サンプル商品 %d", keyword, i),
			Genre:         "日用品",
			ImageURL:      "",
			ContentVolume: 500,
			ContentUnit:   "ml",
		})
	}
	return &RakutenSearchResult{
		Items:      items,
		TotalCount: 100,
		Page:       page,
		PageCount:  10,
	}
}
