package tools

import (
	"bufio"
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
	entriesIPV4 []IPV4Entry = nil
	entriesIPV6 []IPV6Entry = nil
	entriesText []string    = nil
)

func SetupGeolocation(stop context.Context, await *sync.WaitGroup) {
	t := time.Now()

	// Decompress Archive
	archive := bytes.NewReader(include.ArchiveGeolocation)
	gunzip, err := gzip.NewReader(archive)
	if err != nil {
		LoggerGeolocation.Fatal("Invalid Archive Header", err.Error())
	}
	defer gunzip.Close()

	// Preallocate Tables
	buffer := make([]byte, 256)
	reader := bufio.NewReaderSize(gunzip, 1<<20)
	read := func(n int) {
		if _, err := io.ReadAtLeast(reader, buffer[:n], n); err != nil {
			LoggerGeolocation.Fatal("Failed to Read Archive", err.Error())
		}
	}
	read(12)
	num4 := int(binary.LittleEndian.Uint32(buffer[0:4]))
	num6 := int(binary.LittleEndian.Uint32(buffer[4:8]))
	numT := int(binary.LittleEndian.Uint32(buffer[8:12]))
	entriesIPV4 = make([]IPV4Entry, num4)
	entriesIPV6 = make([]IPV6Entry, num6)
	entriesText = make([]string, numT)

	// Decode Entries
	for i := 0; i < num4; i++ {
		read(20)
		entriesIPV4[i] = IPV4Entry{
			RangeStart:     binary.LittleEndian.Uint32(buffer[0:4]),
			CountryIndex:   binary.LittleEndian.Uint32(buffer[4:8]),
			RegionIndex:    binary.LittleEndian.Uint32(buffer[8:12]),
			CityIndex:      binary.LittleEndian.Uint32(buffer[12:16]),
			TimezoneOffset: int32(binary.LittleEndian.Uint32(buffer[16:20])),
		}
	}
	for i := 0; i < num6; i++ {
		read(32)
		entriesIPV6[i] = IPV6Entry{
			RangeStart:     [16]byte(buffer[:16]),
			CountryIndex:   binary.LittleEndian.Uint32(buffer[16:20]),
			RegionIndex:    binary.LittleEndian.Uint32(buffer[20:24]),
			CityIndex:      binary.LittleEndian.Uint32(buffer[24:28]),
			TimezoneOffset: int32(binary.LittleEndian.Uint32(buffer[28:32])),
		}
	}
	for i := 0; i < numT; i++ {
		read(1)
		l := int(buffer[0])
		read(l)
		entriesText[i] = string(buffer[:l])
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
