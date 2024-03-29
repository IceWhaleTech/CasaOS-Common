package external

import (
	"os/exec"
	"strconv"
	"strings"

	"github.com/samber/lo"
)

type GPUInfo struct {
	MemoryTotal    int
	MemoryUsed     int
	MemoryFree     int
	Name           string
	TemperatureGPU int
}

type NvidiaGPUInfo struct {
	Index             int
	UUID              string
	UtilizationGPU    int
	MemoryTotal       int
	MemoryUsed        int
	MemoryFree        int
	DriverVersion     string
	Name              string
	GPUSerial         string
	DisplayActive     bool
	DisplayMode       bool
	TemperatureGPU    int
	PowerDraw         float32 `json:"power_draw"`
	PowerLimit        float32 `json:"power_limit"`
	MemoryUtilization float32 `json:"memory_utilization"`
	Utilization       float32 `json:"utilization"`
}

func NvidiaGPUInfoList() ([]NvidiaGPUInfo, error) {
	GPUInfos := []NvidiaGPUInfo{}

	output, err := exec.Command("nvidia-smi", "--query-gpu=index,uuid,utilization.gpu,memory.total,memory.used,memory.free,driver_version,name,gpu_serial,display_active,display_mode,temperature.gpu,utilization.gpu,utilization.memory,power.draw,power.limit", "--format=csv,noheader,nounits").Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		value := strings.Split(line, ", ")
		if len(value) == 16 {
			index, _ := strconv.Atoi(value[0])
			utilizationGPU, _ := strconv.Atoi(value[2])
			memoryTotal, _ := strconv.Atoi(value[3])
			memoryUsed, _ := strconv.Atoi(value[4])
			memoryFree, _ := strconv.Atoi(value[5])
			temperatureGPU, _ := strconv.Atoi(value[11])
			utilization, _ := strconv.ParseFloat(value[12], 32)
			memoryUtilization, _ := strconv.ParseFloat(value[13], 32)
			powerDraw, _ := strconv.ParseFloat(value[14], 32)
			powerLimit, _ := strconv.ParseFloat(value[15], 32)

			GPUInfos = append(GPUInfos, NvidiaGPUInfo{
				Index:             index,
				UUID:              value[1],
				UtilizationGPU:    utilizationGPU,
				MemoryTotal:       memoryTotal << 20,
				MemoryUsed:        memoryUsed << 20,
				MemoryFree:        memoryFree << 20,
				DriverVersion:     value[6],
				Name:              value[7],
				GPUSerial:         value[8],
				DisplayActive:     value[9] == "Enable",
				DisplayMode:       value[10] == "Enabled",
				TemperatureGPU:    temperatureGPU,
				PowerDraw:         float32(powerDraw),
				PowerLimit:        float32(powerLimit),
				MemoryUtilization: float32(memoryUtilization),
				Utilization:       float32(utilization),
			})
		} else {
			continue
		}
	}
	return GPUInfos, nil
}

func GPUInfoList() ([]GPUInfo, error) {
	GPUInfos := []GPUInfo{}
	nvidiaGPUInfoList, err := NvidiaGPUInfoList()
	if err != nil {
		return nil, err
	}
	GPUInfos = append(GPUInfos, lo.Map(
		nvidiaGPUInfoList, func(gpuInfo NvidiaGPUInfo, index int) GPUInfo {
			return GPUInfo{
				MemoryTotal:    gpuInfo.MemoryTotal,
				MemoryUsed:     gpuInfo.MemoryUsed,
				MemoryFree:     gpuInfo.MemoryFree,
				Name:           gpuInfo.Name,
				TemperatureGPU: gpuInfo.TemperatureGPU,
			}
		})...,
	)
	return GPUInfos, nil
}
