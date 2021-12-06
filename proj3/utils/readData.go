package utils

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func ReadData(filePath string, records map[string]ZipcodeInfo, kZipcode, kMonth, kYear int) {

	// Open the file
	file, _ := os.Open(filePath)
	r := csv.NewReader(file)
	var idx int
	var err error
	var zipcode, month, year, cases, tests, deaths int

	for {
		record, errRead := r.Read()
		idx++
		if idx == 1 { // skip header
			continue
		}
		if errRead == io.EOF { // reach the last line
			break
		}
		if errRead != nil {
			fmt.Printf("file:%s, Problem:%v\n", filePath, errRead)
		}

		key := strings.Join(record, ",")

		if _, present := records[key]; !present {
			if zipcode, err = strconv.Atoi(record[ZipcodeCol]); err != nil || zipcode != kZipcode {
				continue
			}
			startStrs := strings.Split(record[WeekStart], "/")
			if len(startStrs) != 3 {
				continue
			}
			month, _ = strconv.Atoi(startStrs[0])
			if month != kMonth {
				continue
			}
			year, _ = strconv.Atoi(startStrs[2])
			if year != kYear {
				continue
			}
			if cases, err = strconv.Atoi(record[CasesWeek]); err != nil {
				continue
			}
			if tests, err = strconv.Atoi(record[TestsWeek]); err != nil {
				continue
			}
			if deaths, err = strconv.Atoi(record[DeathsWeek]); err != nil {
				continue
			}
			records[key] = ZipcodeInfo{cases, tests, deaths}
		}

	}
	file.Close()
}
