package hcl

import (
	"local-ai-platform/internal/hardware"
)

type ModelRecommendation struct {
	Name           string `json:"name"`
	SizeGB         float64 `json:"sizeGB"`
	Quantization   string `json:"quantization"`
	VRAMRequired  uint64 `json:"vramRequired"`
	RAMRequired   uint64 `json:"ramRequired"`
	Reason        string `json:"reason"`
	Recommended   bool  `json:"recommended"`
}

var recommendedModels = []ModelRecommendation{
	{
		Name:          "llama3.2:3b",
		SizeGB:        2.0,
		Quantization:  "Q4_K_M",
		VRAMRequired: 4 * 1024 * 1024 * 1024,
		RAMRequired:  4 * 1024 * 1024 * 1024,
		Reason:       "Lightweight model for limited hardware",
	},
	{
		Name:          "llama3.1:8b",
		SizeGB:        4.7,
		Quantization:  "Q4_K_M",
		VRAMRequired:  6 * 1024 * 1024 * 1024,
		RAMRequired:  8 * 1024 * 1024 * 1024,
		Reason:       "Fast and efficient, works on most hardware",
	},
	{
		Name:          "llama3.1:13b",
		SizeGB:        8.0,
		Quantization:  "Q5_K_M",
		VRAMRequired:  10 * 1024 * 1024 * 1024,
		RAMRequired:  16 * 1024 * 1024 * 1024,
		Reason:       "Balanced performance for your setup",
	},
	{
		Name:          "llama3.1:70b",
		SizeGB:        40.0,
		Quantization:  "Q4_K_M",
		VRAMRequired: 38 * 1024 * 1024 * 1024,
		RAMRequired:  48 * 1024 * 1024 * 1024,
		Reason:       "Best quality for high-end hardware",
	},
	{
		Name:          "qwen2.5:3b",
		SizeGB:        2.0,
		Quantization:  "Q4_K_M",
		VRAMRequired: 4 * 1024 * 1024 * 1024,
		RAMRequired:  4 * 1024 * 1024 * 1024,
		Reason:       "Fast Chinese-language support",
	},
	{
		Name:          "qwen2.5:14b",
		SizeGB:        9.0,
		Quantization:  "Q5_K_M",
		VRAMRequired:  10 * 1024 * 1024 * 1024,
		RAMRequired:  16 * 1024 * 1024 * 1024,
		Reason:       "Strong Chinese-language performance",
	},
	{
		Name:          "phi4:14b",
		SizeGB:        9.0,
		Quantization:  "Q4_K_M",
		VRAMRequired: 10 * 1024 * 1024 * 1024,
		RAMRequired:  16 * 1024 * 1024 * 1024,
		Reason:       "Microsoft's efficient instruction model",
	},
	{
		Name:          "mistral:7b",
		SizeGB:        4.1,
		Quantization:  "Q4_K_M",
		VRAMRequired:  6 * 1024 * 1024 * 1024,
		RAMRequired:  8 * 1024 * 1024 * 1024,
		Reason:       "Balanced all-purpose model",
	},
}

func RecommendModels(specs *hardware.SystemSpecs) []ModelRecommendation {
	var recs []ModelRecommendation

	totalVRAM := uint64(0)
	for _, g := range specs.GPUs {
		if g.VRAMBytes > totalVRAM {
			totalVRAM = g.VRAMBytes
		}
	}

	if len(specs.GPUs) > 0 && specs.GPUs[0].Backend == "metal" {
		totalVRAM = specs.Memory.TotalBytes
	}

	totalRAM := specs.Memory.TotalBytes

	isAppleSilicon := len(specs.GPUs) > 0 && specs.GPUs[0].Backend == "metal"
	isNvidia := false
	for _, g := range specs.GPUs {
		if g.Backend == "cuda" {
			isNvidia = true
			break
		}
	}

	hasAmpleRAM := totalRAM >= 32*1024*1024*1024
	hasModerateRAM := totalRAM >= 16*1024*1024*1024
	hasLimitedRAM := totalRAM >= 8*1024*1024*1024

	for _, model := range recommendedModels {
		canRun := false

		if isAppleSilicon || isNvidia {
			if model.VRAMRequired <= totalVRAM {
				canRun = true
			} else if totalVRAM == 0 && model.VRAMRequired <= totalRAM {
				canRun = true
			}
		} else {
			if model.RAMRequired <= totalRAM {
				canRun = true
			}
		}

		if canRun {
			rec := model
			rec.Recommended = false

			if isAppleSilicon && model.Name == "llama3.2:3b" {
				rec.Recommended = true
			} else if hasAmpleRAM && model.Name == "llama3.1:13b" {
				rec.Recommended = true
			} else if hasModerateRAM && model.Name == "llama3.1:8b" {
				rec.Recommended = true
			} else if hasLimitedRAM && model.Name == "llama3.2:3b" {
				rec.Recommended = true
			}

			recs = append(recs, rec)
		}
	}

	if len(recs) == 0 {
		recs = append(recs, ModelRecommendation{
			Name:         "llama3.2:3b",
			SizeGB:       2.0,
			Quantization: "Q4_K_M",
			VRAMRequired: 4 * 1024 * 1024 * 1024,
			RAMRequired:  4 * 1024 * 1024 * 1024,
			Reason:       "Minimum viable model for your hardware",
		})
	}

	return recs
}

func GetTopRecommendation(specs *hardware.SystemSpecs) *ModelRecommendation {
	recs := RecommendModels(specs)
	if len(recs) > 0 {
		for i := range recs {
			if recs[i].Recommended {
				return &recs[i]
			}
		}
		return &recs[0]
	}
	return nil
}

func GetCompatibleModels(specs *hardware.SystemSpecs) []ModelRecommendation {
	recs := RecommendModels(specs)
	return recs
}