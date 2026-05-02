package hardware

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

const cacheFileName = "hardware-cache.json"
const cacheMaxAge = 24 * time.Hour

func Detect() (*SystemSpecs, error) {
	specs := &SystemSpecs{
		DetectedAt: time.Now(),
		OS: OSInfo{
			Platform: runtime.GOOS,
			Arch:     runtime.GOARCH,
		},
	}

	if hi, err := host.Info(); err == nil {
		specs.OS.Version = hi.PlatformVersion
	}

	if err := detectCPU(specs); err != nil {
		return nil, fmt.Errorf("cpu detection: %w", err)
	}

	if err := detectMemory(specs); err != nil {
		return nil, fmt.Errorf("memory detection: %w", err)
	}

	if err := detectStorage(specs); err != nil {
		return nil, fmt.Errorf("storage detection: %w", err)
	}

	if err := detectGPUs(specs); err != nil {
		fmt.Printf("gpu detection warning: %v\n", err)
	}

	specs.Score = computeScore(specs)

	_ = SaveCached(specs)

	return specs, nil
}

func DetectIfNeeded() (*SystemSpecs, error) {
	if cached, err := LoadCached(); err == nil && cached != nil {
		if time.Since(cached.DetectedAt) < cacheMaxAge {
			return cached, nil
		}
	}
	return Detect()
}

func SaveCached(specs *SystemSpecs) error {
	dir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	appDir := filepath.Join(dir, "local-ai-platform")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(specs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(appDir, cacheFileName), data, 0644)
}

func LoadCached() (*SystemSpecs, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "local-ai-platform", cacheFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var specs SystemSpecs
	if err := json.Unmarshal(data, &specs); err != nil {
		return nil, err
	}
	return &specs, nil
}

func ClearCache() error {
	dir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, "local-ai-platform", cacheFileName)
	return os.Remove(path)
}

func detectCPU(specs *SystemSpecs) error {
	info, err := cpu.Info()
	if err != nil {
		return err
	}

	threads, _ := cpu.Counts(true)
	cores, _ := cpu.Counts(false)
	if cores == 0 {
		cores = threads
	}

	var model string
	var mhz float64
	var flags []string

	if len(info) > 0 {
		model = info[0].ModelName
		mhz = info[0].Mhz
		flags = filterFeatures(info[0].Flags)
	}

	specs.CPU = CPUInfo{
		Model:        model,
		Cores:        cores,
		Threads:      threads,
		Architecture: runtime.GOARCH,
		Features:     flags,
		FrequencyMHz: mhz,
	}
	return nil
}

func filterFeatures(flags []string) []string {
	relevant := map[string]bool{"avx": true, "avx2": true, "avx512f": true, "sse4_1": true, "sse4_2": true, "fma": true}
	var out []string
	for _, f := range flags {
		if relevant[f] {
			out = append(out, f)
		}
	}
	return out
}

func detectMemory(specs *SystemSpecs) error {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return err
	}
	specs.Memory = MemoryInfo{
		TotalBytes:   vm.Total,
		UsedBytes:   vm.Used,
		FreeBytes:   vm.Available,
		UsedPercent: vm.UsedPercent,
	}
	return nil
}

func detectStorage(specs *SystemSpecs) error {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	usage, err := disk.Usage(home)
	if err != nil {
		return err
	}
	specs.Storage = StorageInfo{
		TotalBytes:   usage.Total,
		FreeBytes:   usage.Free,
		UsedPercent: usage.UsedPercent,
		Path:        home,
	}
	return nil
}

func computeScore(s *SystemSpecs) int {
	score := 0

	if s.CPU.Cores >= 8 {
		score += 20
	} else if s.CPU.Cores >= 4 {
		score += 10
	}
	for _, f := range s.CPU.Features {
		if f == "avx2" || f == "avx512f" {
			score += 10
			break
		}
	}

	ramGB := s.Memory.TotalBytes / (1024 * 1024 * 1024)
	if ramGB >= 32 {
		score += 30
	} else if ramGB >= 16 {
		score += 20
	} else if ramGB >= 8 {
		score += 10
	}

	for _, g := range s.GPUs {
		if g.IsSupported {
			vramGB := g.VRAMBytes / (1024 * 1024 * 1024)
			if g.Backend == "metal" {
				vramGB = s.Memory.TotalBytes / (1024 * 1024 * 1024)
			}
			if vramGB >= 24 {
				score += 40
			} else if vramGB >= 12 {
				score += 30
			} else if vramGB >= 8 {
				score += 25
			} else if vramGB >= 4 {
				score += 15
			}
			break
		}
	}

	if score > 100 {
		return 100
	}
	return score
}