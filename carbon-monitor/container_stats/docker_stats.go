package container_stats

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"sync"
)

// Energy Profile taken from https://github.com/marmelab/greenframe-cli/blob/main/src/model/README.md
var energyProfile = EnergyProfile{
	CPU:     45,
	MEM:     0.078125,
	DISK:    0.00152,
	NETWORK: 11,
	PUE:     1.4,
}

type EnergyProfile struct {
	CPU     float64 `json:"CPU"`
	MEM     float64 `json:"MEM"`
	DISK    float64 `json:"DISK"`
	NETWORK float64 `json:"NETWORK"`
	PUE     float64 `json:"PUE"`
}

func GetCpuPower(stats *types.StatsJSON) float64 {
	usageDelta := stats.Stats.CPUStats.CPUUsage.TotalUsage - stats.Stats.PreCPUStats.CPUUsage.TotalUsage
	systemDelta := stats.Stats.CPUStats.SystemUsage - stats.Stats.PreCPUStats.SystemUsage
	cpuCount := stats.Stats.CPUStats.OnlineCPUs
	percentageUtil := (float64(usageDelta) / float64(systemDelta)) * float64(cpuCount) * 100
	cpuPower := percentageUtil * energyProfile.PUE * energyProfile.CPU / 3600
	return cpuPower
}

func GetMemoryPower(stats *types.StatsJSON) float64 {
	memoryUsage := float64(stats.Stats.MemoryStats.Usage) / 1073741824 // Convert to GB
	memoryPower := memoryUsage * energyProfile.PUE * energyProfile.MEM / 3600
	return memoryPower
}

func GetNetworkPower(stats *types.StatsJSON) float64 {
	totalRx := 0.0
	totalTx := 0.0
	for _, network := range stats.Networks {
		totalRx += float64(network.RxBytes)
		totalTx += float64(network.TxBytes)
	}
	networkPower := (totalRx + totalTx) / 1073741824 * energyProfile.NETWORK / 7200
	return networkPower
}

func GetSingleContainerStat(cli *client.Client, containerID string, containerName string, containerPower map[string]float64, wg *sync.WaitGroup) (bool, error) {
	defer wg.Done()

	stats, err := cli.ContainerStats(context.Background(), containerID, false)

	if err != nil {
		return false, err
	}

	defer stats.Body.Close()

	data := types.StatsJSON{}
	err = json.NewDecoder(stats.Body).Decode(&data)
	if err != nil {
		return false, err
	}

	totalPower := GetCpuPower(&data) + GetMemoryPower(&data) + GetNetworkPower(&data)
	log.Debug(fmt.Sprintf("Container: %s Power: %f", containerName, totalPower))
	containerPower[containerName] = totalPower
	return true, nil
}

func GetDockerStats(cli *client.Client, containerPower map[string]float64) {

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
		All: false,
	})
	if err != nil {
		log.Fatal(err)
	}

	containerLen := len(containers)
	var wg sync.WaitGroup
	wg.Add(containerLen)

	for _, container := range containers {
		go GetSingleContainerStat(
			cli,
			container.ID,
			container.Names[0],
			containerPower,
			&wg,
		)
	}
	wg.Wait()
}
