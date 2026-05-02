package main

import (
	"context"
	"local-ai-platform/internal/hardware"
	"local-ai-platform/internal/hcl"
	"local-ai-platform/internal/models"
	"local-ai-platform/internal/ollama"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// HardwareDetect runs a fresh full hardware scan
func (a *App) HardwareDetect() (*hardware.SystemSpecs, error) {
	return hardware.Detect()
}

// HardwareGetCached returns cached specs (or detects if none exists)
func (a *App) HardwareGetCached() (*hardware.SystemSpecs, error) {
	return hardware.DetectIfNeeded()
}

// HardwareClearCache removes the cached hardware file
func (a *App) HardwareClearCache() error {
	return hardware.ClearCache()
}

// ============ Ollama API ============

// EnsureOllama ensures Ollama is available, downloading if necessary
func (a *App) EnsureOllama() error {
	ollama.SetDownloadProgressCallback(func(progress ollama.DownloadProgress) {
		runtime.EventsEmit(a.ctx, "ollama:download:progress", progress)
	})

	_, err := ollama.EnsureOllama(nil)
	return err
}

// OllamaStart starts the Ollama runtime
func (a *App) OllamaStart() error {
	return ollama.GetManager().Start()
}

// OllamaStop stops the Ollama runtime
func (a *App) OllamaStop() error {
	return ollama.GetManager().Stop()
}

// OllamaStatus returns true if Ollama is running
func (a *App) OllamaStatus() bool {
	return ollama.GetManager().IsRunning()
}

// ListModels returns installed models
func (a *App) ListModels() ([]ollama.ModelInfo, error) {
	return ollama.ListModels()
}

// PullModel downloads a model
func (a *App) PullModel(name string) error {
	return ollama.PullModel(name, nil)
}

// DeleteModel removes a model
func (a *App) DeleteModel(name string) error {
	return ollama.DeleteModel(name)
}

// GetCacheSize returns the model cache size
func (a *App) GetCacheSize() (int64, error) {
	return ollama.GetCacheSize()
}

// GetModelRecommendations returns hardware-based recommendations
func (a *App) GetModelRecommendations() ([]hcl.ModelRecommendation, error) {
	specs, err := hardware.DetectIfNeeded()
	if err != nil {
		return nil, err
	}
	return hcl.RecommendModels(specs), nil
}

// GetCurrentModel returns the currently loaded model
func (a *App) GetCurrentModel() (string, error) {
	return ollama.GetCurrentModel()
}

// GetOllamaVersion returns the pinned Ollama version
func (a *App) GetOllamaVersion() string {
	return ollama.GetPinnedVersion()
}

// ============ Hugging Face API ============

// SearchHFModels searches Hugging Face for GGUF models
func (a *App) SearchHFModels(query string) ([]models.CatalogModel, error) {
	return models.SearchHFModels(query, 50)
}

// GetPopularHFModels returns popular GGUF models from Hugging Face
func (a *App) GetPopularHFModels() ([]models.CatalogModel, error) {
	return models.GetPopularHFModels(30)
}

// GetAllPopularHFModels returns all popular models from multiple searches
func (a *App) GetAllPopularHFModels() ([]models.CatalogModel, error) {
	return models.GetAllPopularHFModels()
}
