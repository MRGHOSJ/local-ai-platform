package models

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const hfAPIBase = "https://huggingface.co/api/models"

type HFSearchResult struct {
	ID        string   `json:"id"`
	Downloads int      `json:"downloads"`
	Likes     int      `json:"likes"`
	Tags      []string `json:"tags"`
}

type HFFile struct {
	RFilename string `json:"rfilename"`
	Size     int64  `json:"size"`
}

type ModelTag struct {
	Tag            string  `json:"tag"`
	Quantization  string  `json:"quantization"`
	SizeGB        float64 `json:"sizeGB"`
	VRAMRequiredGB float64 `json:"vramRequiredGB"`
	RAMRequiredGB  float64 `json:"ramRequiredGB"`
	SourceFile    string  `json:"sourceFile"`
}

type CatalogModel struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	Description string      `json:"description"`
	Provider     string      `json:"provider"`
	TotalPulls   int         `json:"totalPulls"`
	Tags         []ModelTag  `json:"tags"`
}

var (
	hfCache     []CatalogModel
	hfCacheTTL  time.Time
)

const cacheTTL = 24 * time.Hour

func SearchHFModels(query string, limit int) ([]CatalogModel, error) {
	if limit == 0 {
		limit = 100
	}

	if time.Since(hfCacheTTL) < cacheTTL && len(hfCache) > 0 {
		if query == "" {
			return hfCache, nil
		}
		var filtered []CatalogModel
		q := strings.ToLower(query)
		for _, m := range hfCache {
			if strings.Contains(strings.ToLower(m.Name), q) || strings.Contains(strings.ToLower(m.ID), q) {
				filtered = append(filtered, m)
			}
		}
		return filtered, nil
	}

	searchTerm := query
	if searchTerm == "" {
		searchTerm = "gguf"
	}

	url := fmt.Sprintf("%s?search=%s&sort=downloads&direction=-1&limit=%d&full=true",
		hfAPIBase, searchTerm, limit)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HF API returned %d", resp.StatusCode)
	}

	var results []HFSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}

	var models []CatalogModel
	for _, r := range results {
		tags, err := fetchHFFiles(r.ID)
		if err != nil {
			// Keep models even if we can't find GGUF files
			tags = []ModelTag{{
				Tag:            "unknown",
				Quantization:  "unknown",
				SizeGB:         0,
				VRAMRequiredGB: 8,
				RAMRequiredGB:  10,
				SourceFile:     "unknown",
			}}
		}

		if len(tags) == 0 {
			continue
		}

		parts := strings.Split(r.ID, "/")
		name := parts[len(parts)-1]
		name = strings.ReplaceAll(name, "-", " ")

		models = append(models, CatalogModel{
			ID:          sanitizeHFID(r.ID),
			Name:        name,
			Description: fmt.Sprintf("%s • %d downloads", parts[0], r.Downloads),
			Provider:    "huggingface",
			TotalPulls:  r.Downloads,
			Tags:        tags,
		})
	}

	hfCache = models
	hfCacheTTL = time.Now()
	return models, nil
}

func GetPopularHFModels(limit int) ([]CatalogModel, error) {
	return SearchHFModels("", limit)
}

func GetAllPopularHFModels() ([]CatalogModel, error) {
	var allModels []CatalogModel
	seen := make(map[string]bool)

	searches := []string{"", "llama", "mistral", "qwen", "phi", "gemma", "deepseek", "code", "instruct"}

	for _, search := range searches {
		models, err := SearchHFModels(search, 100)
		if err != nil {
			continue
		}
		for _, m := range models {
			if !seen[m.ID] {
				seen[m.ID] = true
				allModels = append(allModels, m)
			}
		}
	}

	sort.Slice(allModels, func(i, j int) bool {
		return allModels[i].TotalPulls > allModels[j].TotalPulls
	})

	return allModels, nil
}

func fetchHFFiles(modelID string) ([]ModelTag, error) {
	branches := []string{"main", "master", "gguf"}

	for _, branch := range branches {
		tags, err := fetchHFFilesFromBranch(modelID, branch)
		if err == nil && len(tags) > 0 {
			return tags, nil
		}
	}
	return nil, fmt.Errorf("no GGUF files found in any branch")
}

func fetchHFFilesFromBranch(modelID, branch string) ([]ModelTag, error) {
	url := fmt.Sprintf("https://huggingface.co/api/models/%s/tree/%s", modelID, branch)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("branch %s returned %d", branch, resp.StatusCode)
	}

	var files []HFFile
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, err
	}

	ggufRegex := regexp.MustCompile(`(?i)(?:^|/)([^/]+?)(?:-(\d+[bm]))?(?:-([A-Z0-9_]+))?\.gguf$`)

	var tags []ModelTag
	for _, f := range files {
		if !strings.HasSuffix(strings.ToLower(f.RFilename), ".gguf") {
			continue
		}

		// ALWAYS create a tag for every GGUF file
		tag := ModelTag{
			SourceFile: f.RFilename,
			SizeGB:     float64(f.Size) / (1024 * 1024 * 1024),
		}

		// Try to parse filename for params/quantization
		matches := ggufRegex.FindStringSubmatch(f.RFilename)
		if len(matches) >= 4 {
			if matches[2] != "" {
				tag.Tag = matches[2]
			}
			if matches[3] != "" {
				tag.Quantization = matches[3]
			}
		}

		// Fallback: extract params from path if regex failed
		if tag.Tag == "" {
			tag.Tag = extractParamsFromPath(f.RFilename)
		}
		if tag.Quantization == "" {
			tag.Quantization = "Q4_K_M"
		}

		// Estimate VRAM
		tag.VRAMRequiredGB = estimateVRAM(tag.Tag, tag.Quantization)
		tag.RAMRequiredGB = tag.VRAMRequiredGB * 1.2

		tags = append(tags, tag)
	}

	if len(tags) == 0 {
		return nil, fmt.Errorf("no GGUF files in branch %s", branch)
	}

	// Sort by VRAM (smallest first)
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].VRAMRequiredGB < tags[j].VRAMRequiredGB
	})

	return tags, nil
}

func extractParamsFromPath(path string) string {
	lower := strings.ToLower(path)
	patterns := []string{"1b", "3b", "4b", "7b", "8b", "13b", "14b", "27b", "30b", "34b", "40b", "70b", "72b", "236b", "405b"}
	for _, p := range patterns {
		if strings.Contains(lower, p) {
			return p
		}
	}
	return "7b"
}

func estimateVRAM(paramCount, quantization string) float64 {
	billion := parseParamCount(paramCount)

	var bits float64
	switch {
	case strings.HasPrefix(quantization, "Q4"):
		bits = 4.5
	case strings.HasPrefix(quantization, "Q5"):
		bits = 5.5
	case strings.HasPrefix(quantization, "Q6"):
		bits = 6.5
	case strings.HasPrefix(quantization, "Q8"):
		bits = 8.5
	case strings.Contains(quantization, "FP16"):
		bits = 16
	default:
		bits = 4.5
	}

	vram := (billion * 1e9 * bits) / 8 / (1024 * 1024 * 1024) * 1.15
	return vram
}

func parseParamCount(s string) float64 {
	s = strings.ToLower(s)
	if strings.HasSuffix(s, "b") {
		val, _ := strconv.ParseFloat(strings.TrimSuffix(s, "b"), 64)
		return val
	}
	if strings.HasSuffix(s, "m") {
		val, _ := strconv.ParseFloat(strings.TrimSuffix(s, "m"), 64)
		return val / 1000
	}
	return 7
}

func sanitizeHFID(id string) string {
	return "hf--" + strings.ReplaceAll(id, "/", "--")
}

func estimateFromModelID(id string) []ModelTag {
	lower := strings.ToLower(id)

	var params string
	if strings.Contains(lower, "70b") {
		params = "70b"
	} else if strings.Contains(lower, "13b") {
		params = "13b"
	} else if strings.Contains(lower, "8b") {
		params = "8b"
	} else if strings.Contains(lower, "3b") {
		params = "3b"
	} else if strings.Contains(lower, "1b") {
		params = "1b"
	} else {
		params = "7b"
	}

	vram := estimateVRAM(params, "Q4_K_M")
	return []ModelTag{
		{
			Tag:            params,
			Quantization:  "Q4_K_M",
			SizeGB:         0,
			VRAMRequiredGB: vram,
			RAMRequiredGB:  vram * 1.2,
		},
	}
}

func ClearHFCache() {
	hfCache = nil
	hfCacheTTL = time.Time{}
}

func getEmbeddedOllamaModels() []ModelTag {
	models := []string{
		"llama3.1", "llama3.2", "llama3.3", "mistral", "mixtral",
		"qwen2.5", "deepseek-coder-v2", "phi4", "gemma2",
		"codellama", "neural-chat", "llava", "nomic-embed-text",
		"mxbai-embed-large", "starcoder2", "dolphin-mixtral",
		"wizardlm2", "yi", "falcon3", "granite3",
	}

	// Return as ModelTag with placeholder values
	// These will be used when Ollama is not running
	var tags []ModelTag
	for _, m := range models {
		tags = append(tags, ModelTag{
			Tag:            "7b",
			Quantization:  "Q4_K_M",
			SizeGB:         4.7,
			VRAMRequiredGB: 6,
			RAMRequiredGB:  8,
			SourceFile:     m,
		})
	}
	return tags
}