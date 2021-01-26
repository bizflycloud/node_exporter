// This file is part of bizfly-agent
//
// Copyright (C) 2020  BizFly Cloud
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>

// +build !nogpu

package collector

import (
	"strconv"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/mindprince/gonvml"
)

var (
	averageDuration = 10 * time.Second
)

type gpuCollector struct {
	gpuMinorNumber           *prometheus.Desc
	gpuName                  *prometheus.Desc
	gpuUUID                  *prometheus.Desc
	gpuTemperature           *prometheus.Desc
	gpuPowerUsage            *prometheus.Desc
	gpuFanSpeed              *prometheus.Desc
	gpuMemoryTotal           *prometheus.Desc
	gpuMemoryUsed            *prometheus.Desc
	gpuUtilizationMemory     *prometheus.Desc
	gpuUtilizationGPU        *prometheus.Desc
	gpuUtilizationGPUAverage *prometheus.Desc
	logger                   log.Logger
}

type gpuDevice struct {
	Index                 string
	MinorNumber           string
	Name                  string
	UUID                  string
	Temperature           float64
	PowerUsage            float64
	FanSpeed              float64
	MemoryTotal           float64
	MemoryUsed            float64
	UtilizationMemory     float64
	UtilizationGPU        float64
	UtilizationGPUAverage float64
}

type gpuMetrics struct {
	Version string
	Devices []gpuDevice
}

func init() {
	registerCollector("gpu", defaultEnabled, NewGPUCollector)
}

// NewGPUCollector returns a new Collector exposing CPU stats.
func NewGPUCollector(logger log.Logger) (Collector, error) {
	subsystem := "gpu"
	return &gpuCollector{
		gpuTemperature: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "Temperature"),
			"Temperature of GPU device in system",
			[]string{"minornumber", "name", "uuid", "system_driver_version"}, nil,
		),
		gpuPowerUsage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "PowerUsage"),
			"Power Usage of GPU device in system",
			[]string{"minornumber", "name", "uuid", "system_driver_version"}, nil,
		),
		gpuFanSpeed: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "FanSpeed"),
			"Fan Speed of GPU device in system",
			[]string{"minornumber", "name", "uuid", "system_driver_version"}, nil,
		),
		gpuMemoryTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "MemoryTotal_Bytes"),
			"Memory Total of GPU device in system",
			[]string{"minornumber", "name", "uuid", "system_driver_version"}, nil,
		),
		gpuMemoryUsed: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "MemoryUsed_Bytes"),
			"Memory Used of GPU device in system",
			[]string{"minornumber", "name", "uuid", "system_driver_version"}, nil,
		),
		gpuUtilizationMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "UtilizationMemory"),
			"Utilization Memory of GPU device in system",
			[]string{"minornumber", "name", "uuid", "system_driver_version"}, nil,
		),
		gpuUtilizationGPU: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "UtilizationGPU"),
			"Utilization of GPU device in system",
			[]string{"minornumber", "name", "uuid", "system_driver_version"}, nil,
		),
		gpuUtilizationGPUAverage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "UtilizationGPUAverage"),
			"Utilization Average of GPU device in system",
			[]string{"minornumber", "name", "uuid", "system_driver_version"}, nil,
		),
		logger: logger,
	}, nil
}

func (g *gpuCollector) Update(ch chan<- prometheus.Metric) error {
	gpu, err := collectMetricDevice()
	if err != nil {
		level.Debug(g.logger).Log("msg", "gpu information is unavailable to collect")
		return nil
	}

	for _, metrics := range gpu.Devices {
		ch <- prometheus.MustNewConstMetric(
			g.gpuTemperature, prometheus.GaugeValue, metrics.Temperature, metrics.MinorNumber, metrics.Name, metrics.UUID, gpu.Version)
		ch <- prometheus.MustNewConstMetric(
			g.gpuPowerUsage, prometheus.GaugeValue, metrics.PowerUsage, metrics.MinorNumber, metrics.Name, metrics.UUID, gpu.Version)
		ch <- prometheus.MustNewConstMetric(
			g.gpuFanSpeed, prometheus.GaugeValue, metrics.FanSpeed, metrics.MinorNumber, metrics.Name, metrics.UUID, gpu.Version)
		ch <- prometheus.MustNewConstMetric(
			g.gpuMemoryTotal, prometheus.CounterValue, metrics.MemoryTotal, metrics.MinorNumber, metrics.Name, metrics.UUID, gpu.Version)
		ch <- prometheus.MustNewConstMetric(
			g.gpuMemoryUsed, prometheus.GaugeValue, metrics.MemoryUsed, metrics.MinorNumber, metrics.Name, metrics.UUID, gpu.Version)
		ch <- prometheus.MustNewConstMetric(
			g.gpuUtilizationMemory, prometheus.GaugeValue, metrics.UtilizationMemory, metrics.MinorNumber, metrics.Name, metrics.UUID, gpu.Version)
		ch <- prometheus.MustNewConstMetric(
			g.gpuUtilizationGPU, prometheus.GaugeValue, metrics.UtilizationGPU, metrics.MinorNumber, metrics.Name, metrics.UUID, gpu.Version)
		ch <- prometheus.MustNewConstMetric(
			g.gpuUtilizationGPUAverage, prometheus.GaugeValue, metrics.UtilizationGPUAverage, metrics.MinorNumber, metrics.Name, metrics.UUID, gpu.Version)
	}

	return nil
}

func collectMetricDevice() (*gpuMetrics, error) {
	if err := gonvml.Initialize(); err != nil {
		return nil, err
	}
	defer gonvml.Shutdown()

	version, err := gonvml.SystemDriverVersion()
	if err != nil {
		return nil, err
	}

	metrics := &gpuMetrics{
		Version: version,
	}

	numDevices, err := gonvml.DeviceCount()
	if err != nil {
		return nil, err
	}

	for index := 0; index < int(numDevices); index++ {
		device, err := gonvml.DeviceHandleByIndex(uint(index))
		if err != nil {
			return nil, err
		}

		uuid, err := device.UUID()
		if err != nil {
			return nil, err
		}

		name, err := device.Name()
		if err != nil {
			return nil, err
		}

		minorNumber, err := device.MinorNumber()
		if err != nil {
			return nil, err
		}

		temperature, err := device.Temperature()
		if err != nil {
			return nil, err
		}

		powerUsage, err := device.PowerUsage()
		if err != nil {
			return nil, err
		}

		fanSpeed, err := device.FanSpeed()
		if err != nil {
			return nil, err
		}

		memoryTotal, memoryUsed, err := device.MemoryInfo()
		if err != nil {
			return nil, err
		}

		utilizationGPU, utilizationMemory, err := device.UtilizationRates()
		if err != nil {
			return nil, err
		}

		utilizationGPUAverage, err := device.AverageGPUUtilization(averageDuration)
		if err != nil {
			return nil, err
		}

		metrics.Devices = append(metrics.Devices,
			gpuDevice{
				Index:                 strconv.Itoa(index),
				MinorNumber:           strconv.Itoa(int(minorNumber)),
				Name:                  name,
				UUID:                  uuid,
				Temperature:           float64(temperature),
				PowerUsage:            float64(powerUsage),
				FanSpeed:              float64(fanSpeed),
				MemoryTotal:           float64(memoryTotal),
				MemoryUsed:            float64(memoryUsed),
				UtilizationMemory:     float64(utilizationMemory),
				UtilizationGPU:        float64(utilizationGPU),
				UtilizationGPUAverage: float64(utilizationGPUAverage),
			})
	}

	return metrics, nil
}
