package agent

import "time"

// GPUMetrics represents the complete physical and digital state of a single GPU
type GPUMetrics struct {
	ID        uint      `json:"id"`
	UUID      string    `json:"uuid"`
	Timestamp time.Time `json:"timestamp"`

	// Physical Telemetry (Mechatronics)
	Temperature float64 `json:"temperature_c"`
	PowerDraw   float64 `json:"power_draw_w"`
	PowerLimit  float64 `json:"power_limit_w"`
	FanSpeed    uint    `json:"fan_speed_pct"`

	// Performance Telemetry
	VRAMUsed       uint64 `json:"vram_used_bytes"`
	VRAMTotal      uint64 `json:"vram_total_bytes"`
	GPUUtilization uint   `json:"gpu_utilization_pct"`

	// Error Telemetry (Digital)
	XidError     uint   `json:"xid_error"`
	EccSingleBit uint64 `json:"ecc_single_bit"` // Soft errors counter
	EccMultiBit  uint64 `json:"ecc_multi_bit"`  // Critical error counter
}

// TelemetryProvider isolates the hardware access layer from the agent's business logic
type TelemetryProvider interface {
	Init() error
	GetDeviceCount() (uint, error)
	QueryMetrics(deviceIdx uint) (GPUMetrics, error)
	InjectFault(deviceIdx uint, xid uint, tempDelta float64) // For Chaos Engineering
	Close() error
}
