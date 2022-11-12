package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/eu-evops/edulink/pkg/edulink"
	"github.com/go-redis/cache/v8"
)

type CacheableRequest struct {
	ApiMethod string        `json:"api_method"`
	TTL       time.Duration `json:"ttl"`
}

var (
	CACHEABLE_REQUESTS = []CacheableRequest{
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
	for _, cacheableRequest := range CACHEABLE_REQUESTS {
		if cacheableRequest.ApiMethod == apiMethod {
			return true
		}
	}
	return false
}

func call(body edulink.Request, response edulink.Result) error {
	apiMethod := body.GetBaseRequest().Method

	if isCacheableRequest(apiMethod) {
		fmt.Printf("Request cachable: '%s', checking cache\n", apiMethod)
		if Cache.Exists(context.Background(), apiMethod) {
			fmt.Printf("Found in cache: '%s', returning\n", apiMethod)
			return Cache.Get(context.Background(), apiMethod, response)
		}

		fmt.Printf("Request not cached, calling API: '%s'\n", apiMethod)
	}

	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", edulink.API_ENDPOINT, bytes.NewBuffer(bodyBytes))

	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-api-method", apiMethod)

	if body.GetBaseRequest().AuthToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", body.GetBaseRequest().AuthToken))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	respBody, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(respBody, response)

	if !response.GetBaseResult().Success {
		parsedJSON, _ := json.MarshalIndent(response, "", "  ")
		fmt.Printf("Response body: %s\n", respBody)
		fmt.Printf("Parsed JSON: %s\n", parsedJSON)
		fmt.Println()
		return fmt.Errorf("API call failed: %s", apiMethod)
	}

	if isCacheableRequest(apiMethod) {
		fmt.Printf("Caching response: '%s'\n", apiMethod)
		Cache.Set(&cache.Item{
			Ctx:   context.Background(),
			Key:   apiMethod,
			Value: response,
			TTL:   24 * time.Hour,
		})
	}

	return nil
}
