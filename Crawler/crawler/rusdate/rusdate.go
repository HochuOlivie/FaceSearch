package rusdate

import (
	"Crawler/crawler"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const rusDateStrategyId = "RusDate"

type rusDateCrawlerStrategy struct {
	client *http.Client
}

func CreateRusDateStrategy() crawler.Strategy {
	return rusDateCrawlerStrategy{client: http.DefaultClient}
}

type searchResponse struct {
	AlertCode string `json:"alert_code"`
	Members   []struct {
		ProfileTemplate struct {
			Mode string `json:"mode"`
			Type string `json:"type"`
		} `json:"profile_template"`
		MemberId int    `json:"member_id"`
		Status   int    `json:"status"`
		Username string `json:"username"`
		Name     string `json:"name"`
		Age      int    `json:"age"`
		Gender   struct {
			Id    int    `json:"id"`
			Title string `json:"title"`
		} `json:"gender"`
		TotalPhotos int `json:"total_photos"`
		TotalVideos int `json:"total_videos"`
		MainPhoto   struct {
			PhotoId int    `json:"photo_id"`
			Photo   string `json:"photo"`
		} `json:"main_photo"`
	} `json:"members"`
	NextPage       bool `json:"next_page"`
	AdditionalData struct {
		ButtonNextPage bool   `json:"button_next_page"`
		Srchqs         string `json:"srchqs"`
	} `json:"additional_data"`
	SearchMarks struct {
		Gender   string `json:"gender"`
		AgeRange string `json:"age_range"`
		GeoTitle string `json:"geo_title"`
		Photo    string `json:"photo"`
	} `json:"search_marks"`
}

// Crawl extracts profiles using search results
// Example of the payload : action=search&op=s&pt=&genre=1&look_genre=0&age_from=20&age_to=26&geo_select=30&look_photo=1&look_online=0&position=2&portion=12&service=Search&task=GetSearchResult
// Position in the example is the page number
func (r rusDateCrawlerStrategy) Crawl(req crawler.Request) (err error, results []crawler.ImageWithMetadata, nextRequests []crawler.Request) {
	var bodyInserter io.Reader
	if req.Method == crawler.POST {
		bodyInserter = strings.NewReader(req.Body)
	} else {
		bodyInserter = nil
	}
	request, err := http.NewRequest(string(req.Method), req.Url, bodyInserter)

	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36")

	response, err := r.client.Do(request)
	if err != nil {
		return err, nil, nil
	}
	if response.StatusCode != http.StatusOK {
		return errors.New(
			fmt.Sprint("HTTP call to ", req.Url, " resulted in HTTP code ", response.StatusCode)), nil, nil
	}

	bodyContent, err := io.ReadAll(response.Body)
	if err != nil {
		return err, nil, nil
	}

	return parseRusDateSearch(req.Body, bodyContent)
}

func (r rusDateCrawlerStrategy) Id() string {
	return "RusDate"
}

func parseRusDateSearch(currentSearchBody string, responseBody []byte) (err error, results []crawler.ImageWithMetadata, nextRequests []crawler.Request) {
	var response searchResponse
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return err, nil, nil
	}
	results = make([]crawler.ImageWithMetadata, len(response.Members))
	for i, member := range response.Members {
		// validate URLs
		imageUrl, err := url.Parse(strings.ReplaceAll(member.MainPhoto.Photo, "\\/", "/"))
		if err != nil {
			return err, nil, nil
		}
		pageUrl := "https://rusdate.de/u/" + member.Username
		results[i] = crawler.ImageWithMetadata{
			ImageURL: imageUrl.String(),
			PageURL:  pageUrl,
			Title:    fmt.Sprint("RusDate ", member.Name),
		}
	}

	formParams, err := url.ParseQuery(currentSearchBody)
	if err != nil {
		return err, nil, nil
	}
	pageNumber, err := strconv.Atoi(formParams.Get("position"))
	if err != nil {
		return err, nil, nil
	}
	formParams.Set("position", strconv.Itoa(pageNumber+1))

	if response.NextPage {
		nextRequests = []crawler.Request{{
			Method:      crawler.POST,
			Url:         "https://rusdate.de/api/get_rest.php",
			StrategyId:  rusDateStrategyId,
			ContentType: "application/x-www-form-urlencoded; charset=UTF-8",
			Body:        formParams.Encode(),
		}}
	} else {
		nextRequests = []crawler.Request{}
	}

	return nil, results, nextRequests
}
