package goreporter

import (
	"github.com/CapillarySoftware/gostat/protoStat"
	nano "github.com/op/go-nanomsg"
	// "strings"
	"fmt"
	"sync"
	"time"
)

var asyncQ chan *protoStat.ProtoStat
var wg sync.WaitGroup

func init() {
	wg.Add(1)
	asyncQ = make(chan *protoStat.ProtoStat, 1000)
	go asyncProcess(asyncQ)
}

type Reporter struct {
	socket *nano.PushSocket
	async  chan *protoStat.ProtoStat
}

//async background thread that processing all stats
func asyncProcess(q <-chan *protoStat.ProtoStat) {
	stats := make(map[string]*protoStat.ProtoStat)
	reportInterval := make(chan bool, 1)
	go func() {
		for {
			time.Sleep(5 * time.Second)
			reportInterval <- true
		}
	}()
main:
	for {
		select {
		case m := <-q:
			if nil == m {
				break main
			}
			updateMap(stats, m)
			fmt.Println(stats)
		case report := <-reportInterval:
			fmt.Println("Time to report :", report)
		}
	}

	//cleanup anything still on the queue
	for m := range q {
		if nil == m {
			break
		}
		updateMap(stats, m)
	}

	fmt.Println("Finished bg thread")
	wg.Done()
}

//update map with new data
func updateMap(stats map[string]*protoStat.ProtoStat, stat *protoStat.ProtoStat) {
	// fmt.Println("map: ", stats)
	ck := stat.GetKey() + stat.GetIndexKey()
	oldStat, ok := stats[ck]
	if !ok {
		stats[ck] = stat
	} else {
		v := oldStat.GetValue() + stat.GetValue()
		oldStat.Value = &v
	}
}

//New reporter that reports when flushed
func NewReporter(url *string) (r Reporter, err error) {

	r = Reporter{async: asyncQ}
	err = r.connect(url)
	r.SetTimeout(0)
	return
}

//Connect to the remote server
func (this *Reporter) connect(url *string) (err error) {
	this.socket, err = nano.NewPushSocket()
	if nil != err {
		return
	}
	_, err = this.socket.Connect(*url)
	return
}

//Set the timeout for the send socket
func (this *Reporter) SetTimeout(millis time.Duration) {
	this.socket.SetSendTimeout(millis * time.Millisecond)
}

func (this *Reporter) AddStat(key string, value float64) {
	stat := protoStat.ProtoStat{Key: &key, Value: &value}
	this.async <- &stat
}

//Add multiple stats into the same graph
func (this *Reporter) AddStatWIndex(key string, value float64, indexKey string) {
	stat := protoStat.ProtoStat{Key: &key, Value: &value, IndexKey: &indexKey}
	this.async <- &stat
}

func (this *Reporter) Close() {
	close(this.async)
	fmt.Println("Closed queue, waiting for cleanup")
	wg.Wait()
}

// //Send data to queue
// func (this *Reporter) Flush() (err error) {
// 	var s []*protoStat.ProtoStat
// 	stats := new(protoStat.ProtoStats)
// 	for k, v := range this.stats {
// 		stat := protoStat.ProtoStat{Key: &k, Value: &v}
// 		s = append(s, &stat)
// 	}
// 	stats.Stats = s
// 	bytes, err := stats.Marshal()
// 	_, err = this.socket.Send(bytes, 0) //blocking
// 	return
// }
