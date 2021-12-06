package balancer

import (
	"math/rand"
	"proj3/utils"
	"time"
)

type BalanceTask struct {
	Filepath string
}

func (bt *BalanceTask) Execute(arg interface{}) {
	ctx := arg.(*utils.SharedContext)

	records := make(map[string]utils.ZipcodeInfo)

	utils.ReadData(bt.Filepath, records, ctx.Zipcode, ctx.Month, ctx.Year)

	ctx.RWmutex.Lock()
	for key, value := range records {
		if _, present := ctx.Records[key]; !present {
			ctx.TotalCases += value.Cases
			ctx.TotalTests += value.Tests
			ctx.TotalDeaths += value.Deaths
			ctx.Records[key] = true
		}
	}

	// update filesCounter
	ctx.FilesCounter--
	ctx.RWmutex.Unlock()

}

type SharingWorker struct {
	queues    *[]chan interface{}
	myGid     int
	ctx       *utils.SharedContext
	done      bool
	threshold int
}

func NewSharingWorker(assignedQueue int, ctx interface{}, queues *[]chan interface{}, thd int) *SharingWorker {
	return &SharingWorker{
		queues:    queues,
		myGid:     assignedQueue,
		ctx:       ctx.(*utils.SharedContext),
		done:      false,
		threshold: thd,
	}
}

func (worker *SharingWorker) Run() {

	myRandSrc := rand.NewSource(time.Now().UnixNano())
	myRand := rand.New(myRandSrc)

	var curSize int
	var victim int
	var max, min int
	var task interface{}

	for {

		if worker.ctx.FilesCounter == 0 {
			worker.done = true
		}

		if worker.done {
			break
		}

		select {
		case task = <-(*worker.queues)[worker.myGid]:
		default:
			task = nil
		}

		if task != nil {
			curTask := task.(BalanceTask)
			curTask.Execute(worker.ctx)
		}

		curSize = len((*worker.queues)[worker.myGid])

		// Rebalance
		if myRand.Intn(curSize+1) == curSize {
			victim = myRand.Intn(len(*worker.queues))

			if victim <= worker.myGid {
				min, max = victim, worker.myGid
			} else {
				min, max = worker.myGid, victim
			}
			worker.balance(&(*worker.queues)[min], &(*worker.queues)[max])

		}

	}

	worker.ctx.WgContext.Done()

}

func (worker *SharingWorker) balance(q0 *chan interface{}, q1 *chan interface{}) {

	var qMin, qMax *chan interface{}
	if len(*q0) < len(*q1) {
		qMin, qMax = q0, q1
	} else {
		qMin, qMax = q1, q0
	}

	if len(*qMax)-len(*qMin) > worker.threshold {

		for len(*qMax) > len(*qMin) {
			(*qMin) <- (<-(*qMax))

		}
	}

}
