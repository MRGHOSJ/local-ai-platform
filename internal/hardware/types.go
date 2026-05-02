package hardware

import "time"

type SystemSpecs struct {
	CPU        CPUInfo        `json:"cpu"`
	GPUs       []GPUInfo      `json:"gpus"`
	Memory     MemoryInfo     `json:"memory"`
	Storage    StorageInfo    `json:"storage"`
	OS         OSInfo         `json:"os"`
	Score      int            `json:"score"`
	DetectedAt time.Time      `json:"detectedAt"`
}

type CPUInfo struct {
	Model        string   `json:"model"`
	Cores        int      `json:"cores"`
	Threads      int      `json:"threads"`
	Architecture string   `json:"architecture"`
	Features     []string `json:"features"`
	FrequencyMHz float64  `json:"frequencyMHz"`
}

type GPUInfo struct {
	Index       int    `json:"index"`
	Vendor      string `json:"vendor"`
	Model       string `json:"model"`
	VRAMBytes   uint64 `json:"vramBytes"`
	Driver      string `json:"driver"`
	Backend     string `json:"backend"`
	IsSupported bool   `json:"isSupported"`
}

type MemoryInfo struct {
	TotalBytes   uint64  `json:"totalBytes"`
	UsedBytes   uint64  `json:"usedBytes"`
	FreeBytes   uint64  `json:"freeBytes"`
	UsedPercent float64 `json:"usedPercent"`
}

type StorageInfo struct {
	TotalBytes   uint64  `json:"totalBytes"`
	FreeBytes   uint64  `json:"freeBytes"`
	UsedPercent float64 `json:"usedPercent"`
	Path        string  `json:"path"`
}

type OSInfo struct {
	Platform string `json:"platform"`
	Arch     string `json:"arch"`
	Version  string `json:"version"`
}