package tools

import (
	"bytes"
	"compress/gzip"
	"context"
	_ "embed"
	"encoding/binary"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bakonpancakzz/template-auth/include"
)

type IPV4Entry struct {
	RangeStart, CountryIndex, RegionIndex, CityIndex uint32
	TimezoneOffset                                   int32
}
type IPV6Entry struct {
	RangeStart                           [16]byte
	CountryIndex, RegionIndex, CityIndex uint32
	TimezoneOffset                       int32
}

var (
	entriesIPV4 []IPV4Entry
	entriesIPV6 []IPV6Entry
	entriesText []string
)

func SetupGeolocation(stop context.Context, await *sync.WaitGroup) {
	// Arguments are unused, they just exist for async startup :L
	t := time.Now()

	// Decompress Archive
	reader := bytes.NewReader(include.ArchiveGeolocation)
	gunzip, err := gzip.NewReader(reader)
	if err != nil {
		LoggerGeolocation.Fatal("Invalid archive header", err)
	}
	defer gunzip.Close()

	// Buffers
	i32 := int32(0)
	b1 := make([]byte, 1)     // string length
	b4 := make([]byte, 4)     // ipv4 range or uint32
	b16 := make([]byte, 16)   // ipv6 range
	b255 := make([]byte, 255) // string
	var load = func(buf []byte) []byte {
		if _, err := io.ReadFull(gunzip, buf); err != nil {
			LoggerGeolocation.Fatal("Failed to read archive", err)
		}
		return buf
	}
	var loadTime = func() int32 {
		binary.Read(gunzip, binary.LittleEndian, &i32)
		return i32
	}

	// Decode Entries
	entriesIPV4 = make([]IPV4Entry, binary.LittleEndian.Uint32(load(b4)))
	entriesIPV6 = make([]IPV6Entry, binary.LittleEndian.Uint32(load(b4)))
	entriesText = make([]string, binary.LittleEndian.Uint32(load(b4)))
	for i := 0; i < len(entriesIPV4); i++ {
		entriesIPV4[i] = IPV4Entry{
			RangeStart:     binary.LittleEndian.Uint32(load(b4)),
			CountryIndex:   binary.LittleEndian.Uint32(load(b4)),
			RegionIndex:    binary.LittleEndian.Uint32(load(b4)),
			CityIndex:      binary.LittleEndian.Uint32(load(b4)),
			TimezoneOffset: loadTime(),
		}
	}
	for i := 0; i < len(entriesIPV6); i++ {
		entriesIPV6[i] = IPV6Entry{
			RangeStart:     [16]byte(load(b16)),
			CountryIndex:   binary.LittleEndian.Uint32(load(b4)),
			RegionIndex:    binary.LittleEndian.Uint32(load(b4)),
			CityIndex:      binary.LittleEndian.Uint32(load(b4)),
			TimezoneOffset: loadTime(),
		}
	}
	for i := 0; i < len(entriesText); i++ {
		entriesText[i] = string(load(b255[:load(b1)[0]]))
	}

	LoggerGeolocation.Info("Ready", map[string]any{
		"time":        time.Since(t).String(),
		"loc_strings": len(entriesText),
		"ipv4_ranges": len(entriesIPV4),
		"ipv6_ranges": len(entriesIPV6),
	})
}

func GeolocateIPV4(address string) (bool, IPV4Entry) {
	// Parse IP address
	var ipValue uint32
	var octets = strings.SplitN(address, ".", 4)
	if len(octets) != 4 {
		return false, IPV4Entry{}
	}
	for i, segment := range octets {
		value, err := strconv.Atoi(segment)
		if err != nil || value < 0 || value > 255 {
			return false, IPV4Entry{}
		}
		ipValue += uint32(value << ((3 - i) * 8))
	}

	// Perform a Binary Search for IP Address Range
	low, high := 0, len(entriesIPV4)-1
	for low <= high {
		mid := (low + high) / 2
		entry := entriesIPV4[mid]
		if ipValue < entry.RangeStart {
			high = mid - 1
		} else if mid == len(entriesIPV4)-1 || ipValue < entriesIPV4[mid+1].RangeStart {
			return true, entry
		} else {
			low = mid + 1
		}
	}

	// Theoretically Impossible
	return false, IPV4Entry{}
}

func GeolocateIPV6(address string) (bool, IPV6Entry) {
	// Convert net.IP to [16]byte
	ip := net.ParseIP(address)
	if ip == nil {
		return false, IPV6Entry{}
	}
	ip16 := ip.To16()
	if ip16 == nil || len(ip16) != 16 {
		return false, IPV6Entry{}
	}
	var ipBytes [16]byte
	copy(ipBytes[:], ip16)

	// Compare two [16]byte IPs
	compareIP := func(a, b [16]byte) int {
		for i := 0; i < 16; i++ {
			if a[i] < b[i] {
				return -1
			}
			if a[i] > b[i] {
				return 1
			}
		}
		return 0
	}

	// Binary search in sorted EntriesIPV6
	low, high := 0, len(entriesIPV6)-1
	for low <= high {
		mid := (low + high) / 2
		entry := entriesIPV6[mid]
		comp := compareIP(ipBytes, entry.RangeStart)

		if comp < 0 {
			high = mid - 1
		} else if mid == len(entriesIPV6)-1 || compareIP(ipBytes, entriesIPV6[mid+1].RangeStart) < 0 {
			return true, entry
		} else {
			low = mid + 1
		}
	}

	// Theoretically Impossible
	return false, IPV6Entry{}
}

func GeolocateString(index uint32) string {
	return entriesText[index]
}
