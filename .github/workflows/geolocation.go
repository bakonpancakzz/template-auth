package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"io/fs"
	"log"
	"math/big"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
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
	OUTPUT_LOCATION = os.Getenv("OUTPUT_LOCATION")
	REQUEST_KEY     = os.Getenv("API_KEY_IP2LOCATION")
	ENTRIES_IPV4    = []IPV4Entry{}
	ENTRIES_IPV6    = []IPV6Entry{}
	ENTRIES_TEXT    = []string{}
	INDEX_TEXT      = map[string]uint32{}
	INDEX_RENAME    = map[string]string{
		"United States of America": "United States",
		"-":                        "Unknown",
	}
)

func TrimString(givenString string) string {
	return strings.Trim(givenString, "\"")
}

func FindString(givenString string) uint32 {
	if replace, ok := INDEX_RENAME[givenString]; ok {
		givenString = replace
	}
	if index, ok := INDEX_TEXT[givenString]; ok {
		// Duplicate String, return index
		return index
	} else {
		// Unique String Given, return new index
		ENTRIES_TEXT = append(ENTRIES_TEXT, givenString)
		index := uint32(len(ENTRIES_TEXT) - 1)
		INDEX_TEXT[givenString] = index
		return index
	}
}

func ParseOffset(offset string) int32 {
	if offset == "-" {
		return 0
	}
	sign := 1
	if offset[0] == '-' {
		sign = -1
	}
	parts := strings.Split(offset[1:], ":")
	hours, _ := strconv.Atoi(parts[0])
	mins, _ := strconv.Atoi(parts[1])
	return int32(sign * ((hours * 3600) + (mins * 60)))
}

func Download(code, filename string) fs.File {

	// Fetch Archive
	var data []byte
	if _, err := os.Stat(filename); err == nil {
		// Check Disk for Cached File
		body, err := os.ReadFile(filename)
		if err != nil {
			log.Fatalln("Cannot Open File:", err)
		}
		data = body
	} else {
		// Download Archive from API
		url := fmt.Sprintf("https://www.ip2location.com/download/?token=%s&file=%s", REQUEST_KEY, code)
		res, err := http.Get(url)
		if err != nil {
			log.Fatalln("Request Error", err)
		}
		defer res.Body.Close()

		// Download Server is flawed and returns a 200 OK even if an error was thrown
		// Instead we can check to see if Received a Reasonable Amount of Bytes
		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatalln("Response Error:", err)
		}
		if len(body) < 1024 {
			log.Fatalln("Server Error:", string(body))
		}
		data = body
	}

	// Unzip Archive and Read CSV
	unzip, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		log.Fatalln("Unzip Error:", err)
	}
	csv, err := unzip.Open(filename)
	if err != nil {
		log.Fatalln("Unzip Open Error", err)
	}

	return csv
}

func main() {
	t := time.Now()
	if REQUEST_KEY == "" {
		log.Fatalln("Variable 'API_KEY_IP2LOCATION' is not set.")
	}
	if OUTPUT_LOCATION == "" {
		OUTPUT_LOCATION = "geolocation.kani.gz"
	}

	// Process IPV4 CSV
	{
		log.Println("Downloading IPV4 Archive")
		f := Download("DB11LITECSV", "IP2LOCATION-LITE-DB11.CSV")
		defer f.Close()

		log.Println("Processing IPV4 Archive")
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			splits := strings.SplitN(scanner.Text(), ",", 10)
			if len(splits) != 10 {
				log.Fatalf("Error Decoding Line %d: Bad Split\n", len(ENTRIES_IPV4)+1)
			}

			rangeStart, err := strconv.Atoi(TrimString(splits[0]))
			if err != nil {
				log.Fatalf("Error Decoding Range on Line %d: %s\n", len(ENTRIES_IPV4)+1, err)
			}

			ENTRIES_IPV4 = append(ENTRIES_IPV4, IPV4Entry{
				RangeStart:     uint32(rangeStart),
				CountryIndex:   FindString(TrimString(splits[3])),
				RegionIndex:    FindString(TrimString(splits[4])),
				CityIndex:      FindString(TrimString(splits[5])),
				TimezoneOffset: ParseOffset(TrimString(splits[9])),
			})
		}
	}

	// Process IPV6 CSV
	{
		log.Println("Downloading IPV6 Archive")
		f := Download("DB11LITECSVIPV6", "IP2LOCATION-LITE-DB11.IPV6.CSV")
		defer f.Close()

		log.Println("Processing IPV6 Archive")
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			splits := strings.SplitN(scanner.Text(), ",", 10)
			if len(splits) != 10 {
				log.Fatalf("Error Decoding Line %d: Bad Split\n", len(ENTRIES_IPV6)+1)
			}
			rangeStr := TrimString(splits[0])
			bigInt := new(big.Int)
			if _, ok := bigInt.SetString(rangeStr, 10); !ok {
				log.Fatalf("Error Decoding Range on Line %d\n", len(ENTRIES_IPV6)+1)
			}
			rangeBytes := bigInt.FillBytes(make([]byte, 16)) // big-endian by default
			rangeStart := [16]byte{}
			copy(rangeStart[:], rangeBytes)

			ENTRIES_IPV6 = append(ENTRIES_IPV6, IPV6Entry{
				RangeStart:     rangeStart,
				CountryIndex:   FindString(TrimString(splits[3])),
				RegionIndex:    FindString(TrimString(splits[4])),
				CityIndex:      FindString(TrimString(splits[5])),
				TimezoneOffset: ParseOffset(TrimString(splits[9])),
			})
		}
	}

	// Setup Compression
	log.Printf(
		"Compressing: %d IPV4 Ranges, %d IPV6 Ranges, %d Strings",
		len(ENTRIES_IPV4), len(ENTRIES_IPV6), len(ENTRIES_TEXT),
	)
	if err := os.MkdirAll(path.Dir(OUTPUT_LOCATION), 0755); err != nil {
		log.Fatalln("Error Creating Archive Folder:", err)
	}

	output, err := os.Create(OUTPUT_LOCATION)
	if err != nil {
		log.Fatalln("Error Creating Archive File:", err)
	}
	defer output.Close()

	writer, err := gzip.NewWriterLevel(output, gzip.BestCompression)
	if err != nil {
		log.Fatalln("Error Creating Compressor:", err)
	}
	defer writer.Close()

	// Repackage Contents
	binary.Write(writer, binary.LittleEndian, uint32(len(ENTRIES_IPV4)))
	binary.Write(writer, binary.LittleEndian, uint32(len(ENTRIES_IPV6)))
	binary.Write(writer, binary.LittleEndian, uint32(len(ENTRIES_TEXT)))
	for _, e := range ENTRIES_IPV4 {
		binary.Write(writer, binary.LittleEndian, e.RangeStart)
		binary.Write(writer, binary.LittleEndian, e.CountryIndex)
		binary.Write(writer, binary.LittleEndian, e.RegionIndex)
		binary.Write(writer, binary.LittleEndian, e.CityIndex)
		binary.Write(writer, binary.LittleEndian, e.TimezoneOffset)
	}
	for _, e := range ENTRIES_IPV6 {
		binary.Write(writer, binary.LittleEndian, e.RangeStart)
		binary.Write(writer, binary.LittleEndian, e.CountryIndex)
		binary.Write(writer, binary.LittleEndian, e.RegionIndex)
		binary.Write(writer, binary.LittleEndian, e.CityIndex)
		binary.Write(writer, binary.LittleEndian, e.TimezoneOffset)
	}
	for _, s := range ENTRIES_TEXT {
		binary.Write(writer, binary.LittleEndian, uint8(len(s)))
		io.WriteString(writer, s)
	}
	log.Printf("Complete in %s", time.Since(t))
}
