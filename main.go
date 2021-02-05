package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

type CustomHandler struct {
	wg *sync.WaitGroup
}

func NewCustomHandler(wg *sync.WaitGroup) *CustomHandler {
	return &CustomHandler{wg: wg}
}

func (h *CustomHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobName := vars["jobName"]

	fmt.Fprintf(w, "job %s started", jobName)

	h.wg.Add(3)
	go slowJob1(jobName, h.wg)
	go slowJob2(jobName, h.wg)
	go slowJob3(jobName, h.wg)
}

func slowJob1(name string, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("starting job 1 for %s\n", name)
	time.Sleep(5 * time.Second)
	fmt.Printf("finished job 1 for %s\n", name)
}

func slowJob2(name string, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("starting job 2 for %s\n", name)
	time.Sleep(5 * time.Second)
	fmt.Printf("finished job 2 for %s\n", name)
}

func slowJob3(name string, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("starting job 3 for %s\n", name)
	time.Sleep(5 * time.Second)
	fmt.Printf("finished job 3 for %s\n", name)
}

func main() {
	wg := &sync.WaitGroup{}
	customHandler := NewCustomHandler(wg)

	router := mux.NewRouter()
	router.Handle("/{jobName}", customHandler)

	httpServer := &http.Server{
		Addr:    "127.0.0.1:8080",
		Handler: router,
	}

	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-termChan
		log.Print("SIGTERM received. Shutdown process initiated\n")
		httpServer.Shutdown(context.Background())
	}()

	if err := httpServer.ListenAndServe(); err != nil {
		if err.Error() != "http: Server closed" {
			log.Printf("HTTP server closed with: %v\n", err)
		}
		log.Printf("HTTP server shut down")
	}

	log.Println("waiting for running jobs to finish")
	wg.Wait()
	log.Println("jobs finished. exiting")
}
