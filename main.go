package main

import (
	"aegis-gpu/agent"
	"fmt"
	"log"
	"time"
)

func main() {
	log.Println("=== Start Aegis-GPU Agent ===")

	// Initialize the telemetry provider - using advanced Mock for now
	// Simulating a node equipped with 4 GPUs
	provider := agent.NewMockProvider(4)
	if err := provider.Init(); err != nil {
		log.Fatalf("Telemetry initialization failed: %v", err)
	}
	defer provider.Close()

	// Store previous samples to calculate the derivative: temperature trend
	previousTemps := make(map[uint]float64)

	// Sampling rate: 1 Hz (every 1 second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Fire up the Chaos module in a separate thread (Goroutine) to inject a fault after 5 seconds
	go func() {
		time.Sleep(5 * time.Second)
		log.Println("\n[CHAOS MONKEY] !!! Injecting liquid cooling failure and NVLink fault on GPU 2 !!!")
		// Inject XID 45 error (NVLink Fault) and a sudden temperature spike (0.9°C per second)
		provider.InjectFault(2, 45, 1.8)
	}()

	for range ticker.C {
		count, _ := provider.GetDeviceCount()

		fmt.Print("\033[H\033[2J") // Clear the terminal screen for a clean, real-time data view
		fmt.Println("=== AEGIS-GPU TELEMETRY MONITOR LIVE ===")
		fmt.Printf("Time: %s | Active GPU Nodes: %d\n", time.Now().Format("15:04:05"), count)
		fmt.Println("--------------------------------------------------------------------------------")

		for i := uint(0); i < count; i++ {
			metrics, err := provider.QueryMetrics(i)
			if err != nil {
				log.Printf("Error reading GPU %d: %v", i, err)
				continue
			}

			// Calculate the temperature derivative over time: dT/dt
			prevTemp, exists := previousTemps[i]
			var dTdt float64
			if exists {
				dTdt = metrics.Temperature - prevTemp
			}
			previousTemps[i] = metrics.Temperature

			// State machine and early warning system: Fault Detection
			status := "HEALTHY"
			alertMessage := ""

			if metrics.XidError != 0 {
				status = "CRITICAL_FAULT"
				alertMessage = fmt.Sprintf("HARDWARE ERROR DETECTED: XID %d (Immediate node isolation required!)", metrics.XidError)
			} else if dTdt > 1.2 { // If temperature rises faster than 1.2°C/s
				status = "THERMAL_ANOMALY"
				alertMessage = fmt.Sprintf("ALERT: Rapid temperature spike! dT/dt = +%.2f°C/s. Possible cooling failure.", dTdt)
			} else if metrics.Temperature > 82.0 {
				status = "WARNING_HOT"
				alertMessage = "Warning: Temperature exceeds safe threshold."
			}

			fmt.Printf("GPU [%d] [%s] ID: %s\n", metrics.ID, status, metrics.UUID)
			fmt.Printf("   |-- Metrics: Temp: %.2f°C (dT/dt: %+.2f°C/s) | Power Draw: %.1fW / %.1fW\n",
				metrics.Temperature, dTdt, metrics.PowerDraw, metrics.PowerLimit)
			fmt.Printf("   |-- Resource: Utilization: %d%% | VRAM: %d GB / %d GB\n",
				metrics.GPUUtilization, metrics.VRAMUsed/1024/1024/1024, metrics.VRAMTotal/1024/1024/1024)

			if alertMessage != "" {
				fmt.Printf("   └── \033[1;31m[ACTION REQUIRED] %s\033[0m\n", alertMessage)
			}
			fmt.Println()
		}
	}
}
