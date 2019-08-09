package main

import (
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

type Server struct {
	*http.ServeMux
	Config
	Logger
}

func NewServer(c Config, l Logger) Server {
	s := Server{Logger: l, Config: c, ServeMux: http.NewServeMux()}
	l.Println("creating server", s)
	return s
}

func StartServer(s Server) {
	s.Println("starting server on", s.Port)
	if err := http.ListenAndServe(s.Port, s); err != nil {
		s.Println("error starting:", err.Error())
	}
}

func (s Server) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	s.Println("handling", pattern)
	s.ServeMux.HandleFunc(pattern, handler)
}

type Logger struct {
	*log.Logger
}

func NewLogger(c Config) Logger {
	return Logger{Logger: log.New(os.Stderr, c.ServiceName+": ", 0)}
}

type Metrics struct {
	Logger
	mx *sync.Map
}

func (m Metrics) Count(metric string) {
	count, _ := m.mx.LoadOrStore(metric, new(uint64))
	atomic.AddUint64(count.(*uint64), 1)
}

func (m Metrics) dump() {
	m.mx.Range(m.logAndClear)
}

func (m Metrics) logAndClear(k, v interface{}) bool {
	count := atomic.LoadUint64(v.(*uint64))
	atomic.AddUint64(v.(*uint64), -count)
	time.AfterFunc(time.Second, m.dump)
	if count != 0 {
		m.Println(k, ":", count)
	}
	return true
}

func NewMetrics(l Logger) Metrics {
	m := Metrics{Logger: l, mx: new(sync.Map)}
	time.AfterFunc(time.Second, m.dump)
	return m
}
