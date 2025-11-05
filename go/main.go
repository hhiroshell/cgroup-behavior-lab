package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type CgroupVersion int

const (
	CgroupV1 CgroupVersion = iota
	CgroupV2
)

func main() {
	printHeader()

	// Detect cgroup version
	cgroupVersion := detectCgroupVersion()
	fmt.Printf("Detected cgroup version: %s\n", cgroupVersion)
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	// Monitor resources continuously
	for {
		printResourceInfo(cgroupVersion)
		time.Sleep(5 * time.Second)
	}
}

func printHeader() {
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("Resource Monitor Started (Go)")
	fmt.Println(strings.Repeat("=", 80))
}

func printResourceInfo(version CgroupVersion) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("[%s]\n", timestamp)
	fmt.Println(strings.Repeat("-", 80))

	// CPU Information
	printCPUInfo(version)
	fmt.Println()

	// Memory Information
	printMemoryInfo(version)
	fmt.Println()

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()
}

func printCPUInfo(version CgroupVersion) {
	fmt.Println("CPU Resources:")

	// Number of CPUs available to the Go runtime
	numCPU := runtime.NumCPU()
	fmt.Printf("  Available CPUs (Go runtime): %d\n", numCPU)

	// GOMAXPROCS
	maxProcs := runtime.GOMAXPROCS(0)
	fmt.Printf("  GOMAXPROCS: %d\n", maxProcs)

	if version == CgroupV2 {
		printCPUInfoV2()
	} else {
		printCPUInfoV1()
	}
}

func printCPUInfoV1() {
	// Read CPU quota and period for cgroup v1
	quotaPath := "/sys/fs/cgroup/cpu/cpu.cfs_quota_us"
	periodPath := "/sys/fs/cgroup/cpu/cpu.cfs_period_us"

	quota, err := readInt64FromFile(quotaPath)
	if err == nil {
		fmt.Printf("  cgroup v1 cpu.cfs_quota_us: %d\n", quota)
	}

	period, err := readInt64FromFile(periodPath)
	if err == nil {
		fmt.Printf("  cgroup v1 cpu.cfs_period_us: %d\n", period)
	}

	if quota > 0 && period > 0 {
		cpuLimit := float64(quota) / float64(period)
		fmt.Printf("  CPU Limit: %.2f cores\n", cpuLimit)
	} else if quota == -1 {
		fmt.Println("  CPU Limit: unlimited")
	}

	// Read CPU usage
	usagePath := "/sys/fs/cgroup/cpu,cpuacct/cpuacct.usage"
	usage, err := readInt64FromFile(usagePath)
	if err == nil {
		fmt.Printf("  Total CPU usage (nanoseconds): %s\n", formatNumber(usage))
	}
}

func printCPUInfoV2() {
	// Read CPU max for cgroup v2
	cpuMaxPath := "/sys/fs/cgroup/cpu.max"
	content, err := readStringFromFile(cpuMaxPath)
	if err == nil {
		fmt.Printf("  cgroup v2 cpu.max: %s\n", strings.TrimSpace(content))

		parts := strings.Fields(content)
		if len(parts) == 2 {
			if parts[0] == "max" {
				fmt.Println("  CPU Limit: unlimited")
			} else {
				quota, err1 := strconv.ParseInt(parts[0], 10, 64)
				period, err2 := strconv.ParseInt(parts[1], 10, 64)
				if err1 == nil && err2 == nil && period > 0 {
					cpuLimit := float64(quota) / float64(period)
					fmt.Printf("  CPU Limit: %.2f cores\n", cpuLimit)
				}
			}
		}
	}

	// Read CPU stats
	cpuStatPath := "/sys/fs/cgroup/cpu.stat"
	content, err = readStringFromFile(cpuStatPath)
	if err == nil {
		fmt.Println("  cgroup v2 cpu.stat:")
		for _, line := range strings.Split(content, "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				fmt.Printf("    %s\n", line)
			}
		}
	}
}

func printMemoryInfo(version CgroupVersion) {
	fmt.Println("Memory Resources:")

	// Go runtime memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Printf("  Go Alloc: %s\n", formatBytes(m.Alloc))
	fmt.Printf("  Go TotalAlloc: %s\n", formatBytes(m.TotalAlloc))
	fmt.Printf("  Go Sys: %s\n", formatBytes(m.Sys))
	fmt.Printf("  Go NumGC: %d\n", m.NumGC)
	fmt.Printf("  Go HeapAlloc: %s\n", formatBytes(m.HeapAlloc))
	fmt.Printf("  Go HeapSys: %s\n", formatBytes(m.HeapSys))
	fmt.Printf("  Go HeapInuse: %s\n", formatBytes(m.HeapInuse))

	if version == CgroupV2 {
		printMemoryInfoV2()
	} else {
		printMemoryInfoV1()
	}
}

func printMemoryInfoV1() {
	// Read memory limit for cgroup v1
	limitPath := "/sys/fs/cgroup/memory/memory.limit_in_bytes"
	limit, err := readInt64FromFile(limitPath)
	if err == nil {
		// Very large values indicate no limit (usually > 2^60)
		if limit > (1 << 60) {
			fmt.Println("  cgroup v1 memory.limit_in_bytes: unlimited")
		} else {
			fmt.Printf("  cgroup v1 memory.limit_in_bytes: %s\n", formatBytes(uint64(limit)))
		}
	}

	// Read memory usage
	usagePath := "/sys/fs/cgroup/memory/memory.usage_in_bytes"
	usage, err := readInt64FromFile(usagePath)
	if err == nil {
		fmt.Printf("  cgroup v1 memory.usage_in_bytes: %s\n", formatBytes(uint64(usage)))
	}

	// Read memory stats
	statPath := "/sys/fs/cgroup/memory/memory.stat"
	content, err := readStringFromFile(statPath)
	if err == nil {
		fmt.Println("  cgroup v1 memory.stat (selected):")
		for _, line := range strings.Split(content, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "cache ") || strings.HasPrefix(line, "rss ") ||
				strings.HasPrefix(line, "mapped_file ") || strings.HasPrefix(line, "inactive_anon ") {
				parts := strings.Fields(line)
				if len(parts) == 2 {
					value, err := strconv.ParseUint(parts[1], 10, 64)
					if err == nil {
						fmt.Printf("    %s: %s\n", parts[0], formatBytes(value))
					}
				}
			}
		}
	}
}

func printMemoryInfoV2() {
	// Read memory max for cgroup v2
	memMaxPath := "/sys/fs/cgroup/memory.max"
	content, err := readStringFromFile(memMaxPath)
	if err == nil {
		content = strings.TrimSpace(content)
		if content == "max" {
			fmt.Println("  cgroup v2 memory.max: unlimited")
		} else {
			value, err := strconv.ParseUint(content, 10, 64)
			if err == nil {
				fmt.Printf("  cgroup v2 memory.max: %s\n", formatBytes(value))
			}
		}
	}

	// Read current memory usage
	memCurrentPath := "/sys/fs/cgroup/memory.current"
	current, err := readInt64FromFile(memCurrentPath)
	if err == nil {
		fmt.Printf("  cgroup v2 memory.current: %s\n", formatBytes(uint64(current)))
	}

	// Read memory stats
	memStatPath := "/sys/fs/cgroup/memory.stat"
	content, err = readStringFromFile(memStatPath)
	if err == nil {
		fmt.Println("  cgroup v2 memory.stat (selected):")
		for _, line := range strings.Split(content, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "anon ") || strings.HasPrefix(line, "file ") ||
				strings.HasPrefix(line, "kernel_stack ") || strings.HasPrefix(line, "slab ") {
				parts := strings.Fields(line)
				if len(parts) == 2 {
					value, err := strconv.ParseUint(parts[1], 10, 64)
					if err == nil {
						fmt.Printf("    %s: %s\n", parts[0], formatBytes(value))
					}
				}
			}
		}
	}
}

func detectCgroupVersion() CgroupVersion {
	// Check if cgroup v2 is mounted
	cgroupV2Path := "/sys/fs/cgroup/cgroup.controllers"
	if _, err := os.Stat(cgroupV2Path); err == nil {
		return CgroupV2
	}
	return CgroupV1
}

func (v CgroupVersion) String() string {
	if v == CgroupV2 {
		return "V2"
	}
	return "V1"
}

func readStringFromFile(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func readInt64FromFile(path string) (int64, error) {
	content, err := readStringFromFile(path)
	if err != nil {
		return 0, err
	}
	content = strings.TrimSpace(content)
	value, err := strconv.ParseInt(content, 10, 64)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatNumber(n int64) string {
	in := strconv.FormatInt(n, 10)
	numOfDigits := len(in)
	if n < 0 {
		numOfDigits--
	}
	numOfCommas := (numOfDigits - 1) / 3

	out := make([]byte, len(in)+numOfCommas)
	if n < 0 {
		in, out[0] = in[1:], '-'
	}

	for i, j, k := len(in)-1, len(out)-1, 0; ; i, j = i-1, j-1 {
		out[j] = in[i]
		if i == 0 {
			return string(out)
		}
		if k++; k == 3 {
			j, k = j-1, 0
			out[j] = ','
		}
	}
}
