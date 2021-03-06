// Copyright 2020 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build !nonetdev

package collector

import (
	"encoding/json"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/shirou/gopsutil/net"
)

func getNetDevStats(filter *netDevFilter, logger log.Logger) (netDevStats, error) {
	netInterfaces, err := net.IOCounters(true)
	if err != nil {
		return nil, err
	}

	return parseNetDevStats(netInterfaces, filter, logger)
}

func parseNetDevStats(ni []net.IOCountersStat, filter *netDevFilter, logger log.Logger) (netDevStats, error) {

	netDev := netDevStats{}

	for _, net := range ni {
		dev := net.Name
		if filter.ignored(dev) {
			level.Debug(logger).Log("msg", "Ignoring device", "device", dev)
			continue
		}

		statistic, err := parseToString(net)
		if err != nil {
			return nil, err
		}
		netDev[dev] = statistic
	}
	return netDev, nil
}

func parseToString(data net.IOCountersStat) (map[string]uint64, error) {
	statistic := make(map[string]uint64)

	statsBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(statsBytes, &statistic)

	// Ignore field name in statistic map
	delete(statistic, "name")

	return statistic, nil
}
