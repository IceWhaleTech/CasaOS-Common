package external

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
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

func NvidiaGPUInfoListWithNVML() (info []NvidiaGPUInfo, err error) {
	// defer recover
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
			err = fmt.Errorf("error getting GPU info: %v", r)
		}
	}()
	var GPUInfos []NvidiaGPUInfo

	// Initialize NVML
	if result := nvml.Init(); result != nvml.SUCCESS {
		return nil, fmt.Errorf("error initializing NVML: %w", err)
	}
	defer nvml.Shutdown()

	// Get device count
	deviceCount, result := nvml.DeviceGetCount()
	if result != nvml.SUCCESS {
		return nil, fmt.Errorf("error getting device count: %w", err)
	}

	for i := 0; i < deviceCount; i++ {
		device, result := nvml.DeviceGetHandleByIndex(i)
		if result != nvml.SUCCESS {
			return nil, fmt.Errorf("error getting device handle: %w", err)
		}

		info := NvidiaGPUInfo{}
		info.Index = int(i)

		info.UUID, result = device.GetUUID()
		if result != nvml.SUCCESS {
			return nil, fmt.Errorf("error getting UUID: %w", err)
		}

		utilization, result := device.GetUtilizationRates()
		if result != nvml.SUCCESS {
			return nil, fmt.Errorf("error getting utilization rates: %w", err)
		}
		info.UtilizationGPU = int(utilization.Gpu)
		info.MemoryUtilization = float32(utilization.Memory)

		memInfo, result := device.GetMemoryInfo()
		if result != nvml.SUCCESS {
			return nil, fmt.Errorf("error getting memory info: %w", err)
		}
		info.MemoryTotal = int(memInfo.Total)
		info.MemoryUsed = int(memInfo.Used)
		info.MemoryFree = int(memInfo.Free)

		info.Name, result = device.GetName()
		if result != nvml.SUCCESS {
			return nil, fmt.Errorf("error getting name: %w", err)
		}

		driverVersion, result := nvml.SystemGetDriverVersion()
		if result != nvml.SUCCESS {
			return nil, fmt.Errorf("error getting driver version: %w", err)
		}
		info.DriverVersion = driverVersion

		temp, result := device.GetTemperature(nvml.TEMPERATURE_GPU)
		if result != nvml.SUCCESS {
			return nil, fmt.Errorf("error getting temperature: %w", err)
		}
		info.TemperatureGPU = int(temp)

		powerDraw, result := device.GetPowerUsage()
		if result != nvml.SUCCESS {
			return nil, fmt.Errorf("error getting power usage: %w", err)
		}
		info.PowerDraw = float32(powerDraw) / 1000.0

		powerLimit, result := device.GetEnforcedPowerLimit()
		if result != nvml.SUCCESS {
			return nil, fmt.Errorf("error getting power limit: %w", err)
		}
		info.PowerLimit = float32(powerLimit) / 1000.0
		info.GPUSerial = "[N/A]"
		GPUInfos = append(GPUInfos, info)

	}

	return GPUInfos, nil
}

func NvidiaGPUInfoListWithSMI() ([]NvidiaGPUInfo, error) {
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

func NvidiaGPUInfoList() ([]NvidiaGPUInfo, error) {
	gpusInfo, err := NvidiaGPUInfoListWithSMI()
	if err != nil {
		return nil, err
	}
	return gpusInfo, nil
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
