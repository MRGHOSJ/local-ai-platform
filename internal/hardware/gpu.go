package hardware

import (
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

func detectGPUs(specs *SystemSpecs) error {
	switch runtime.GOOS {
	case "darwin":
		return detectAppleGPUs(specs)
	case "windows", "linux":
		return detectPCGPUs(specs)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func detectAppleGPUs(specs *SystemSpecs) error {
	out, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output()
	if err != nil {
		return err
	}
	brand := strings.TrimSpace(string(out))
	if !strings.Contains(brand, "Apple") {
		return detectIntelMacGPU(specs)
	}

	out, err = exec.Command("sysctl", "-n", "hw.memsize").Output()
	if err != nil {
		return err
	}
	memBytes, _ := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64)

	out, err = exec.Command("system_profiler", "SPDisplaysDataType", "-json").Output()
	if err != nil {
		specs.GPUs = append(specs.GPUs, GPUInfo{
			Index:       0,
			Vendor:      "Apple",
			Model:       brand,
			VRAMBytes:   memBytes,
			Backend:     "metal",
			Driver:      "Metal",
			IsSupported: memBytes >= 8*1024*1024*1024,
		})
		return nil
	}

	model := extractJSONField(string(out), "chipset_model")
	if model == "" {
		model = brand
	}

	specs.GPUs = append(specs.GPUs, GPUInfo{
		Index:       0,
		Vendor:      "Apple",
		Model:       model,
		VRAMBytes:   memBytes,
		Backend:     "metal",
		Driver:      "Metal",
		IsSupported: memBytes >= 8*1024*1024*1024,
	})
	return nil
}

func detectIntelMacGPU(specs *SystemSpecs) error {
	out, err := exec.Command("system_profiler", "SPDisplaysDataType", "-json").Output()
	if err != nil {
		return err
	}
	model := extractJSONField(string(out), "chipset_model")
	vendor := extractJSONField(string(out), "vendor")

	if model != "" {
		vramStr := extractJSONField(string(out), "vram")
		vramBytes := parseMemoryString(vramStr)

		specs.GPUs = append(specs.GPUs, GPUInfo{
			Index:       0,
			Vendor:      normalizeVendor(vendor),
			Model:       model,
			VRAMBytes:   vramBytes,
			Backend:     "none",
			Driver:      "",
			IsSupported: false,
		})
	}
	return nil
}

func detectPCGPUs(specs *SystemSpecs) error {
	if nvidiaGPUs := detectNvidia(); len(nvidiaGPUs) > 0 {
		specs.GPUs = append(specs.GPUs, nvidiaGPUs...)
	}

	if amdGPUs := detectAMD(); len(amdGPUs) > 0 {
		specs.GPUs = append(specs.GPUs, amdGPUs...)
	}

	if len(specs.GPUs) == 0 {
		if fallback := detectGPUFallback(); len(fallback) > 0 {
			specs.GPUs = append(specs.GPUs, fallback...)
		}
	}

	return nil
}

func detectNvidia() []GPUInfo {
	cmd := exec.Command("nvidia-smi",
		"--query-gpu=index,name,memory.total,driver_version",
		"--format=csv,noheader,nounits")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var gpus []GPUInfo
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		parts := strings.Split(line, ", ")
		if len(parts) < 4 {
			continue
		}
		idx, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
		name := strings.TrimSpace(parts[1])
		vramMB, _ := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
		driver := strings.TrimSpace(parts[3])

		gpus = append(gpus, GPUInfo{
			Index:       idx,
			Vendor:      "NVIDIA",
			Model:       name,
			VRAMBytes:   uint64(vramMB * 1024 * 1024),
			Driver:      driver,
			Backend:     "cuda",
			IsSupported: vramMB >= 4096,
		})
	}
	return gpus
}

func detectAMD() []GPUInfo {
	cmd := exec.Command("rocm-smi", "--showproductname", "--showmeminfo", "vram", "--csv")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var gpus []GPUInfo
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for i, line := range lines {
		if i == 0 || strings.HasPrefix(line, "GPU") {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 2 {
			name := strings.TrimSpace(parts[1])
			gpus = append(gpus, GPUInfo{
				Index:       i - 1,
				Vendor:      "AMD",
				Model:       name,
				VRAMBytes:   0,
				Backend:     "rocm",
				Driver:      "ROCm",
				IsSupported: true,
			})
		}
	}
	return gpus
}

func detectGPUFallback() []GPUInfo {
	if runtime.GOOS == "linux" {
		out, err := exec.Command("lspci").Output()
		if err == nil {
			return parseLspci(string(out))
		}
	}
	if runtime.GOOS == "windows" {
		out, err := exec.Command("wmic", "path", "win32_VideoController", "get", "Name,AdapterRAM", "/format:csv").Output()
		if err == nil {
			return parseWmic(string(out))
		}
	}
	return nil
}

func parseLspci(output string) []GPUInfo {
	var gpus []GPUInfo
	lines := strings.Split(output, "\n")
	idx := 0
	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "vga") || strings.Contains(lower, "3d controller") {
			if strings.Contains(lower, "nvidia") {
				gpus = append(gpus, GPUInfo{
					Index: idx, Vendor: "NVIDIA", Model: extractLspciName(line),
					Backend: "cuda", IsSupported: false,
				})
				idx++
			} else if strings.Contains(lower, "amd") || strings.Contains(lower, "advanced micro") {
				gpus = append(gpus, GPUInfo{
					Index: idx, Vendor: "AMD", Model: extractLspciName(line),
					Backend: "rocm", IsSupported: false,
				})
				idx++
			} else if strings.Contains(lower, "intel") {
				gpus = append(gpus, GPUInfo{
					Index: idx, Vendor: "Intel", Model: extractLspciName(line),
					Backend: "none", IsSupported: false,
				})
				idx++
			}
		}
	}
	return gpus
}

func parseWmic(output string) []GPUInfo {
	var gpus []GPUInfo
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for i, line := range lines {
		if i == 0 {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) >= 3 {
			name := strings.TrimSpace(parts[1])
			ramStr := strings.TrimSpace(parts[2])
			ramBytes, _ := strconv.ParseUint(ramStr, 10, 64)
			vendor := "Unknown"
			if strings.Contains(strings.ToLower(name), "nvidia") {
				vendor = "NVIDIA"
			} else if strings.Contains(strings.ToLower(name), "amd") {
				vendor = "AMD"
			}
			gpus = append(gpus, GPUInfo{
				Index:       i - 1,
				Vendor:      vendor,
				Model:       name,
				VRAMBytes:   ramBytes,
				Backend:     "none",
				IsSupported: ramBytes >= 4*1024*1024*1024,
			})
		}
	}
	return gpus
}

func normalizeVendor(v string) string {
	l := strings.ToLower(v)
	switch {
	case strings.Contains(l, "nvidia"):
		return "NVIDIA"
	case strings.Contains(l, "amd") || strings.Contains(l, "advanced micro"):
		return "AMD"
	case strings.Contains(l, "intel"):
		return "Intel"
	case strings.Contains(l, "apple"):
		return "Apple"
	default:
		return strings.Title(v)
	}
}

func extractJSONField(json, field string) string {
	pattern := `"` + field + `"\s*:\s*"([^"]*)"`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(json)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func extractLspciName(line string) string {
	idx := strings.Index(line, ": ")
	if idx != -1 {
		return strings.TrimSpace(line[idx+2:])
	}
	return line
}

func parseMemoryString(s string) uint64 {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, ",", "")

	var multiplier uint64 = 1
	if strings.HasSuffix(s, "gb") {
		multiplier = 1024 * 1024 * 1024
		s = strings.TrimSuffix(s, "gb")
	} else if strings.HasSuffix(s, "mb") {
		multiplier = 1024 * 1024
		s = strings.TrimSuffix(s, "mb")
	}

	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return uint64(val * float64(multiplier))
}