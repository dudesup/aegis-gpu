package agent

import (
	"fmt"
	"math/rand"
	"time"
)

type MockNVMLProvider struct {
	deviceCount uint
	baseTemps   map[uint]float64
	injectedXID map[uint]uint
	tempDelta   map[uint]float64
}

func NewMockProvider(devices uint) *MockNVMLProvider {
	return &MockNVMLProvider{
		deviceCount: devices,
		baseTemps:   make(map[uint]float64),
		injectedXID: make(map[uint]uint),
		tempDelta:   make(map[uint]float64),
	}
}

func (m *MockNVMLProvider) Init() error {
	for i := uint(0); i < m.deviceCount; i++ {
		m.baseTemps[i] = 45.0 // Idle temperature
		m.injectedXID[i] = 0
		m.tempDelta[i] = 0.0
	}
	return nil
}

func (m *MockNVMLProvider) GetDeviceCount() (uint, error) {
	return m.deviceCount, nil
}

func (m *MockNVMLProvider) QueryMetrics(deviceIdx uint) (GPUMetrics, error) {
	if deviceIdx >= m.deviceCount {
		return GPUMetrics{}, fmt.Errorf("device index out of bounds")
	}

	// Simulate measurement noise (physical sensors always jitter slightly)
	noise := (rand.Float64() - 0.5) * 0.4

	// Calculate temperature considering inertia and injected faults
	m.baseTemps[deviceIdx] += m.tempDelta[deviceIdx]

	// Natural cooling back to base temperature if there is no load/fault
	if m.tempDelta[deviceIdx] == 0 && m.baseTemps[deviceIdx] > 45.0 {
		m.baseTemps[deviceIdx] -= 0.1
	}

	currentTemp := m.baseTemps[deviceIdx] + noise

	// Dynamic power draw tied to temperature
	power := 80.0 + (currentTemp-45.0)*4.5
	if power > 350.0 {
		power = 350.0
	}

	return GPUMetrics{
		ID:             deviceIdx,
		UUID:           fmt.Sprintf("GPU-MOCK-UUID-%08X", deviceIdx),
		Timestamp:      time.Now(),
		Temperature:    currentTemp,
		PowerDraw:      power,
		PowerLimit:     350.0,
		FanSpeed:       uint(currentTemp * 1.1),
		VRAMTotal:      85899345920, // 80 GB (like H100)
		VRAMUsed:       uint64(rand.Float64() * 70000000000),
		GPUUtilization: uint(rand.Intn(30) + 70), // Simulate pre-training workload (70-100%)
		XidError:       m.injectedXID[deviceIdx],
		EccSingleBit:   uint64(rand.Intn(2)),
		EccMultiBit:    0,
	}, nil
}

func (m *MockNVMLProvider) InjectFault(deviceIdx uint, xid uint, tempDelta float64) {
	m.injectedXID[deviceIdx] = xid
	m.tempDelta[deviceIdx] = tempDelta
}

func (m *MockNVMLProvider) Close() error {
	return nil
}
