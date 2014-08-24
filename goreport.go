package goreport

import (
	"github.com/CapillarySoftware/gostat/protoStat"
	nano "github.com/op/go-nanomsg"
	// "strings"
	"fmt"
	"sync"
	"time"
)

var asyncQ chan *protoStat.ProtoStat
var confQ chan config
var wg sync.WaitGroup

func init() {
	wg.Add(1)
	asyncQ = make(chan *protoStat.ProtoStat, 1000)
	confQ = make(chan config, 1) //block on config
	go asyncProcess(asyncQ, confQ)
}

// Reporter used by clients
type Reporter struct {
	async chan *protoStat.ProtoStat
	conf  chan config
}

//Simple push socket to send external
type push struct {
	socket *nano.PushSocket
}

// Configuration for the background thread
type config struct {
	timeout int
	url     string
}

//Connect to the remote server
func (this *push) connect(url *string) (err error) {
	if nil != this.socket {
		this.socket.Close()
	}
	this.socket, err = nano.NewPushSocket()
	if nil != err {
		return
	}
	_, err = this.socket.Connect(*url)
	return
}

//Close the push socket
func (this *push) Close() {
	if nil != this.socket {
		this.socket.Close()
	}
}

//Set the timeout for the send socket
func (this *push) SetTimeout(millis time.Duration) {
	this.socket.SetSendTimeout(millis * time.Millisecond)
}

//Create a new push socket, only for internal use
func newPush(url *string, timeout time.Duration) (p push, err error) {
	err = p.connect(url)
	p.SetTimeout(timeout)
	return
}

//async background thread that processing all stats
func asyncProcess(q <-chan *protoStat.ProtoStat, conf <-chan config) {
	c := <-conf
	if c.url == "" {
		fmt.Println("Failed to get valid configuration maybe ReporterConfig not called?")
		wg.Done()
		return
	}
	sendQ, err := newPush(&c.url, time.Duration(c.timeout))
	if nil != err {
		wg.Done()
		return
	}
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

		case c = <-conf:
			// fmt.Println(c)
			sendQ.Close()
			sendQ, err = newPush(&c.url, time.Duration(c.timeout))
			if nil != err {
				fmt.Println("Failed to reconfigure queue")
				break main
			}
		case m := <-q:
			if nil == m {
				break main
			}
			updateMap(stats, m)
		case _ = <-reportInterval:
			// fmt.Println("Time to report :", report)
			// fmt.Println(stats)
			if len(stats) > 0 {
				err = sendQ.sendStats(stats)
				resetStats(stats)
				if nil != err {
					fmt.Println("Failed to send stats: ", err)
				}
			}
		}
	}

	//cleanup anything still on the queue
	for m := range q {
		if nil == m {
			break
		}
		updateMap(stats, m)
	}
	err = sendQ.sendStats(stats)
	if nil != err {
		fmt.Println(err)
	}

	fmt.Println("Finished bg thread")
	wg.Done()
}

//reset all stats to 0
func resetStats(stats map[string]*protoStat.ProtoStat) {
	for _, v := range stats {
		if v.GetRepeat() {
			zero := float64(0)
			v.Value = &zero
		}
	}
}

//update map with new data
func updateMap(stats map[string]*protoStat.ProtoStat, stat *protoStat.ProtoStat) {
	// fmt.Println("map: ", stats)
	ck := stat.GetKey() + stat.GetIndexKey()
	oldStat, ok := stats[ck]
	if !ok {
		if stat.GetRepeat() {
			stats[ck] = stat
		} else {
			fmt.Println("Stat has not been registered! ", stat)
		}
	} else {
		v := oldStat.GetValue() + stat.GetValue()
		oldStat.Value = &v
	}
}

func (this *push) sendStats(stats map[string]*protoStat.ProtoStat) (err error) {
	var s []*protoStat.ProtoStat
	// fmt.Println(stats)
	pStats := new(protoStat.ProtoStats)
	for _, v := range stats {
		s = append(s, v)
	}
	// fmt.Println(s)
	pStats.Stats = s
	now := time.Now().UTC().UnixNano()
	pStats.TimeNano = &now

	fmt.Println(pStats)
	bytes, err := pStats.Marshal()
	if nil != err {
		return
	}
	_, err = this.socket.Send(bytes, 0) //blocking
	return
}

//New reporter that reports at 5 second intervals
func ReporterConfig(url string, timeout int) {
	c := config{timeout: timeout, url: url}
	confQ <- c
	return
}

func NewReporter() (r Reporter) {
	r = Reporter{async: asyncQ, conf: confQ}
	return
}

//Add a stat that should be repeated with 0 when not seen
func (this *Reporter) RegisterStat(key string) {
	b := true
	value := float64(0)
	stat := protoStat.ProtoStat{Key: &key, Value: &value, Repeat: &b}
	this.async <- &stat
}

//Add a stat that should be repeated with 0 when not seen
func (this *Reporter) RegisterStatWIndex(key string, indexKey string) {
	b := true
	value := float64(0)
	stat := protoStat.ProtoStat{Key: &key, Value: &value, IndexKey: &indexKey, Repeat: &b}
	this.async <- &stat
}

//Add a basic key value stat
func (this *Reporter) AddStat(key string, value float64) {
	stat := protoStat.ProtoStat{Key: &key, Value: &value}
	this.async <- &stat
}

//Add multiple stats into the same graph
func (this *Reporter) AddStatWIndex(key string, value float64, indexKey string) {
	stat := protoStat.ProtoStat{Key: &key, Value: &value, IndexKey: &indexKey}
	this.async <- &stat
}

//Close the reporter and the background thread
func (this *Reporter) Close() {
	close(this.async)
	close(this.conf)
	fmt.Println("Closed queue, waiting for cleanup")
	wg.Wait()
}
