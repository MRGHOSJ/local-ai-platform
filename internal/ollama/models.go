package ollama

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type ModelInfo struct {
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	ModifiedAt string `json:"modified_at"`
	Digest     string `json:"digest"`
}

type PullProgress struct {
	Status    string `json:"status"`
	Completed int64  `json:"completed"`
	Total     int64  `json:"total"`
}

func ListModels() ([]ModelInfo, error) {
	resp, err := http.Get(GetAPIURL() + "/api/tags")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to list models: %s", resp.Status)
	}

	var result struct {
		Models []ModelInfo `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Models, nil
}

func PullModel(name string, onProgress func(PullProgress)) error {
	payload := map[string]string{"name": name}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", GetAPIURL()+"/api/pull", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to pull model: %s", resp.Status)
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		var progress PullProgress
		if err := json.Unmarshal(line, &progress); err == nil {
			if progress.Status != "" && onProgress != nil {
				onProgress(progress)
			}
		}
	}
	return nil
}

func DeleteModel(name string) error {
	payload := map[string]string{"name": name}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("DELETE", GetAPIURL()+"/api/delete", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to delete model: %s", resp.Status)
	}
	return nil
}

func GetModelInfo(name string) (*ModelInfo, error) {
	models, err := ListModels()
	if err != nil {
		return nil, err
	}

	for _, m := range models {
		modelName := strings.Split(m.Name, ":")[0]
		reqName := strings.Split(name, ":")[0]
		if modelName == reqName {
			return &m, nil
		}
	}
	return nil, fmt.Errorf("model %s not found", name)
}

func GetCurrentModel() (string, error) {
	resp, err := http.Get(GetAPIURL() + "/api/tags")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Models) > 0 {
		return result.Models[0].Name, nil
	}
	return "", nil
}

func GetCacheSize() (int64, error) {
	var size int64
	err := filepath.Walk(getModelsDir(), func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

func GetCacheSizeFormatted() string {
	size, err := GetCacheSize()
	if err != nil {
		return "Unknown"
	}
	return formatBytes(size)
}

func formatBytes(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}
	const k = 1024
	sizes := []string{"B", "KB", "MB", "GB", "TB"}
	i := 0
	f := float64(bytes)
	for f >= k && i < len(sizes)-1 {
		f /= k
		i++
	}
	return fmt.Sprintf("%.1f %s", f, sizes[i])
}

type GenerateRequest struct {
	Model    string `json:"model"`
	Prompt   string `json:"prompt"`
	Stream   bool   `json:"stream"`
}

type GenerateResponse struct {
	Model     string `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func Generate(prompt string, onChunk func(string)) error {
	req := GenerateRequest{
		Model:  "llama3.2:3b",
		Prompt: prompt,
		Stream: true,
	}

	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequest("POST", GetAPIURL()+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		var genResp GenerateResponse
		if err := json.Unmarshal(line, &genResp); err == nil {
			if genResp.Response != "" && onChunk != nil {
				onChunk(genResp.Response)
			}
			if genResp.Done {
				break
			}
		}
	}
	return nil
}

type ModelDetails struct {
	ParentModel      string `json:"parent_model"`
	Format         string `json:"format"`
	Family         string `json:"family"`
	Families       []string `json:"families"`
	ParameterSize  string `json:"parameter_size"`
	QuantizationLevel string `json:"quantization_level"`
}

func GetModelDetails(name string) (*ModelDetails, error) {
	resp, err := http.Get(GetAPIURL() + "/api/show/" + name)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get model details: %s", resp.Status)
	}

	var details ModelDetails
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, err
	}
	return &details, nil
}

func GetInstalledModelsCount() (int, error) {
	models, err := ListModels()
	if err != nil {
		return 0, err
	}
	return len(models), nil
}

func IsModelDownloaded(name string) (bool, error) {
	models, err := ListModels()
	if err != nil {
		return false, err
	}

	for _, m := range models {
		modelPart := extractModelName(m.Name)
		reqPart := extractModelName(name)
		if modelPart == reqPart {
			return true, nil
		}
	}
	return false, nil
}

func extractModelName(fullName string) string {
	re := regexp.MustCompile(`^([a-zA-Z0-9._-]+)`)
	matches := re.FindStringSubmatch(fullName)
	if len(matches) > 1 {
		return matches[1]
	}
	return strings.Split(fullName, ":")[0]
}