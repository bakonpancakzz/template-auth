package tools

import (
	"strconv"
	"sync/atomic"
	"time"
)

var (
	machineID    uint64
	maxMachineID uint64 = (1 << 10) - 1
	maxSequence  uint64 = (1 << 12) - 1
	sequence     atomic.Uint64
	timestamp    atomic.Int64
)

func init() {
	// Parse and Validate Machine ID from Environment
	if v, err := strconv.ParseUint(MACHINE_ID, 10, 64); err != nil {
		panic("machine id is invalid")
	} else {
		machineID = v
		if machineID > maxMachineID {
			panic("machine id must fit within ten bits")
		}
	}
}

// Generate a Snowflake like integer, not really standard :P
func GenerateSnowflake() uint64 {

	// Restart Sequence if time has advanced enough
	now := time.Now().UnixMilli()
	if now != timestamp.Load() {
		sequence.Store(0)
	} else {
		if sequence.Add(1) > maxSequence {
			// Sequence Overflowed! Wait a millisecond...
			for now <= timestamp.Load() {
				time.Sleep(time.Millisecond)
				now = time.Now().UnixMilli()
			}
		}
	}
	timestamp.Store(now)

	// Shift and combine bits
	return (uint64(now-EPOCH_MILLI) << 22) | (machineID << 12) | sequence.Load()
}
