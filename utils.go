// Utility functions and helpers.
package monigo

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"
)

func GetGoroutineCount() int {
	return runtime.NumGoroutine()
}

func MeasureExecutionTime(name string, f func()) {

	initialGoroutines := runtime.NumGoroutine() // Capturing the initial number of goroutines
	var memStatsBefore, memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	log.Printf("memStatsBefore = %v\n", memStatsBefore.Alloc)

	folder := "profiles"
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		if err := os.Mkdir(folder, 0755); err != nil {
			fmt.Printf("Error creating profiles folder: %v\n", err)
		}
	}
	cpuProfileName := fmt.Sprintf("%s_cpu.prof", name)
	cpuProfFilePath := fmt.Sprintf("%s/%s", folder, cpuProfileName)

	cpuProfileFile, err := StartCPUProfile(cpuProfFilePath)
	if err != nil {
		fmt.Printf("Error starting CPU profile for %s: %v\n", name, err)
		return
	}
	defer StopCPUProfile(cpuProfileFile)

	memProfName := fmt.Sprintf("%s_mem.prof", name)
	memProfFilePath := fmt.Sprintf("%s/%s", folder, memProfName)

	start := time.Now()
	f()
	elapsed := time.Since(start)

	if err := WriteHeapProfile(memProfFilePath); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}

	// Capture final metrics
	runtime.ReadMemStats(&memStatsAfter)
	finalGoroutines := runtime.NumGoroutine() - initialGoroutines
	if finalGoroutines < 0 {
		finalGoroutines = 0
	}

	// Calculate memory usage
	var memoryUsage uint64
	if memStatsAfter.Alloc >= memStatsBefore.Alloc {
		memoryUsage = memStatsAfter.Alloc - memStatsBefore.Alloc
	}

	log.Printf("memStatsAfter = %v\n", memStatsAfter.Alloc)
	log.Printf("memoryUsage = %v: %s\n", memoryUsage, name)

	mu.Lock()
	defer mu.Unlock()

	// Recording the metrics
	functionMetrics[name] = &FunctionMetrics{
		FunctionLastRanAt: start,
		CPUProfile:        cpuProfileFile.Name(),
		MemoryUsage:       memoryUsage,
		GoroutineCount:    finalGoroutines,
		ExecutionTime:     elapsed,
	}
}
