package tools

import (
	"strconv"
	"sync/atomic"
	"time"
)

var (
	machineID    int64
	maxMachineID int64 = (1 << 10) - 1
	maxSequence  int64 = (1 << 12) - 1
	sequence     atomic.Int64
	timestamp    atomic.Int64
)

func init() {
	if v, err := strconv.ParseInt(MACHINE_ID, 10, 64); err != nil {
		panic("machine id is invalid")
	} else {
		machineID = v
		if machineID > maxMachineID {
			panic("machine id must fit within 10 bits")
		}
	}
}

// GenerateSnowflake returns an int64-safe unique ID
func GenerateSnowflake() int64 {
	now := time.Now().UnixMilli()

	// Reset sequence if time has advanced
	if now != timestamp.Load() {
		sequence.Store(0)
	} else {
		if sequence.Add(1) > maxSequence {
			// Sequence overflowed: wait for next millisecond
			for now <= timestamp.Load() {
				time.Sleep(time.Millisecond)
				now = time.Now().UnixMilli()
			}
			sequence.Store(0)
		}
	}

	timestamp.Store(now)

	id := ((now - EPOCH_MILLI) << 22) | (machineID << 12) | sequence.Load()
	return id
}
