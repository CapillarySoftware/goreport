package goreporter

import (
	"github.com/CapillarySoftware/gostat/protoStat"
	nano "github.com/op/go-nanomsg"
	// "strings"
	// "fmt"
	"time"
)

type Reporter struct {
	socket *nano.PushSocket
	stats  map[string]float64
}

func NewReporter(url *string) (r Reporter, err error) {
	r = Reporter{stats: make(map[string]float64)}
	err = r.connect(url)
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
	this.stats[key] += value
}

//Send data to queue
func (this *Reporter) Flush() (err error) {
	var s []*protoStat.ProtoStat
	stats := new(protoStat.ProtoStats)
	for k, v := range this.stats {
		stat := protoStat.ProtoStat{Key: &k, Value: &v}
		s = append(s, &stat)
	}
	stats.Stats = s
	bytes, err := stats.Marshal()
	_, err = this.socket.Send(bytes, 0) //blocking
	return
}
