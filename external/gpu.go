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
	Index          int
	UUID           string
	UtilizationGPU int
	MemoryTotal    int
	MemoryUsed     int
	MemoryFree     int
	DriverVersion  string
	Name           string
	GPUSerial      string
	DisplayActive  bool
	DisplayMode    bool
	TemperatureGPU int
}

func NvidiaGPUInfoList() ([]NvidiaGPUInfo, error) {
	GPUInfos := []NvidiaGPUInfo{}

	output, err := exec.Command("nvidia-smi", "--query-gpu=index,uuid,utilization.gpu,memory.total,memory.used,memory.free,driver_version,name,gpu_serial,display_active,display_mode,temperature.gpu", "--format=csv,noheader,nounits").Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		value := strings.Split(line, ", ")
		if len(value) == 12 {
			index, _ := strconv.Atoi(value[0])
			utilizationGPU, _ := strconv.Atoi(value[2])
			memoryTotal, _ := strconv.Atoi(value[3])
			memoryUsed, _ := strconv.Atoi(value[4])
			memoryFree, _ := strconv.Atoi(value[5])
			temperatureGPU, _ := strconv.Atoi(value[11])
			GPUInfos = append(GPUInfos, NvidiaGPUInfo{
				Index:          index,
				UUID:           value[1],
				UtilizationGPU: utilizationGPU,
				MemoryTotal:    memoryTotal,
				MemoryUsed:     memoryUsed,
				MemoryFree:     memoryFree,
				DriverVersion:  value[6],
				Name:           value[7],
				GPUSerial:      value[8],
				DisplayActive:  value[9] == "Enable",
				DisplayMode:    value[10] == "Enabled",
				TemperatureGPU: temperatureGPU,
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
