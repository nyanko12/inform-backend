package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

const rakutenSearchURL = "https://app.rakuten.co.jp/services/api/IchibaItem/Search/20170706"

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
	applicationID string
	client        *http.Client
}

func NewRakutenService(applicationID string) *RakutenService {
	return &RakutenService{
		applicationID: applicationID,
		client:        &http.Client{},
	}
}

func (s *RakutenService) Search(keyword string, page, hits int) (*RakutenSearchResult, error) {
	if s.applicationID == "" {
		return s.mockResult(keyword, page, hits), nil
	}

	params := url.Values{}
	params.Set("applicationId", s.applicationID)
	params.Set("keyword", keyword)
	params.Set("page", strconv.Itoa(page))
	params.Set("hits", strconv.Itoa(hits))
	params.Set("format", "json")
	params.Set("formatVersion", "2")

	reqURL := rakutenSearchURL + "?" + params.Encode()
	log.Printf("[rakuten] GET %s", reqURL)

	resp, err := s.client.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("rakuten search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[rakuten] error status %d: %s — falling back to mock", resp.StatusCode, string(body))
		return s.mockResult(keyword, page, hits), nil
	}

	var raw struct {
		Count int `json:"count"`
		Page  int `json:"page"`
		PageCount int `json:"pageCount"`
		Items []struct {
			ItemName     string `json:"itemName"`
			ItemCode     string `json:"itemCode"`
			GenreName    string `json:"genreName"`
			MediumImageUrls []struct {
				ImageUrl string `json:"imageUrl"`
			} `json:"mediumImageUrls"`
		} `json:"Items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode rakuten: %w", err)
	}

	result := &RakutenSearchResult{
		TotalCount: raw.Count,
		Page:       raw.Page,
		PageCount:  raw.PageCount,
	}
	for _, it := range raw.Items {
		item := &RakutenItem{
			ItemCode: it.ItemCode,
			Name:     it.ItemName,
			Genre:    it.GenreName,
		}
		if len(it.MediumImageUrls) > 0 {
			item.ImageURL = it.MediumImageUrls[0].ImageUrl
		}
		result.Items = append(result.Items, item)
	}
	return result, nil
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
