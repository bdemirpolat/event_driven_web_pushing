package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
)

func listener() {
	conn, err := amqp.Dial("amqp://user:bitnami@localhost:5672/")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ch.Close()
	msgs, err := ch.Consume(
		"TEST_QUEUE",             // queue
		"SELAM BEN BİR CONSUMER", // consumer tag
		true,                     // auto-ack
		false,                    // exclusive
		false,                    // no-local
		false,                    // no-wait
		nil,                      // args
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	forever := make(chan bool)
	go func() {
		for d := range msgs {
			eventChan <- string(d.Body)
		}
	}()
	<-forever
}

var eventChan chan string

func main() {
	eventChan = make(chan string)
	go listener()
	r := mux.NewRouter()
	r.HandleFunc("/get-info", getInfo)
	srv := &http.Server{
		Addr:    ":3001",
		Handler: r,
	}
	srv.ListenAndServe()
}
func getInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	flusher, _ := w.(http.Flusher)
	for {
		select {
		case message := <-eventChan:
			w.Write([]byte(fmt.Sprintf("data: %v \n\n", message)))
			flusher.Flush()
		case <-r.Context().Done():
			fmt.Println("context done")
			return
		}
	}
}
