package edulink

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/eu-evops/edulink/pkg/cache"
	"github.com/eu-evops/edulink/pkg/cache/common"
)

type CacheableRequest struct {
	ApiMethod string        `json:"api_method"`
	TTL       time.Duration `json:"ttl"`
}

var (
	Cache             *cache.Cache
	CacheableRequests = []CacheableRequest{
		{
			ApiMethod: "EduLink.SchoolDetails",
			TTL:       24 * time.Hour,
		},
		{
			ApiMethod: "EduLink.AchievementBehaviourLookups",
			TTL:       24 * time.Hour,
		},
	}
)

func isCacheableRequest(apiMethod string) bool {
	for _, cacheableRequest := range CacheableRequests {
		if cacheableRequest.ApiMethod == apiMethod {
			return true
		}
	}
	return false
}

func Call(ctx context.Context, body Request, response Result) error {
	apiMethod := body.GetBaseRequest().Method

	if isCacheableRequest(apiMethod) {
		log.Printf("Request cachable: '%s', checking cache\n", apiMethod)
		if Cache.Exists(ctx, apiMethod) {
			log.Printf("Found in cache: '%s', returning\n", apiMethod)
			return Cache.Get(ctx, apiMethod, response)
		}

		log.Printf("Request not cached, calling API: '%s'\n", apiMethod)
	}

	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequestWithContext(ctx, "POST", API_ENDPOINT, bytes.NewBuffer(bodyBytes))

	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-api-method", apiMethod)

	if body.GetBaseRequest().AuthToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", body.GetBaseRequest().AuthToken))
	}

	client := &http.Client{
		Transport: &http.Transport{},
		Timeout:   10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, response)

	if !response.GetBaseResult().Success {
		parsedJSON, _ := json.MarshalIndent(response, "", "  ")
		log.Printf("Response body: %s\n", respBody)
		log.Printf("Parsed JSON: %s\n", parsedJSON)
		log.Println()
		return fmt.Errorf("API call failed: %s", apiMethod)
	}

	if isCacheableRequest(apiMethod) {
		log.Printf("Caching response: '%s'\n", apiMethod)
		Cache.Set(&common.Item{
			Ctx:   ctx,
			Key:   apiMethod,
			Value: response,
			TTL:   24 * time.Hour,
		})
	}

	return nil
}
