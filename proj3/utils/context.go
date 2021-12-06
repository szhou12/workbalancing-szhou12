package utils

import "sync"

const ZipcodeCol = 0
const WeekStart = 2
const CasesWeek = 4
const TestsWeek = 8
const DeathsWeek = 14
const NumOfFiles = 500

const THRESHOLD = 4 // rebalancing threshold

type SharedContext struct {
	WgContext    *sync.WaitGroup
	RWmutex      *sync.RWMutex
	Records      map[string]bool // work as deduplication
	Zipcode      int
	Month        int
	Year         int
	TotalCases   int
	TotalTests   int
	TotalDeaths  int
	FilesCounter int32
}

type ZipcodeInfo struct {
	Cases  int
	Tests  int
	Deaths int
}
