package main

import (
"encoding/json"
"fmt"
"github.com/gorilla/mux"
"github.com/streadway/amqp"
"log"
"net/http"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", err, msg)
	}
}

type Message struct {
	Id int
	Name string
}

func homeHandler(w http.ResponseWriter, r *http.Request)  {
	fmt.Fprintf(w, "Welcome home!")
}

func main() {

	// handle some web requests
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homeHandler)
	go func() {
		log.Fatal(http.ListenAndServe(":8080", router))
	}()

	// listen on a Rabbit Queue
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Could not open channel")
	defer ch.Close()

	msgs, err := ch.Consume(
		"supplyQueue", // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			var payload Message
			err := json.Unmarshal(d.Body, &payload)
			failOnError(err, "Failed to unmarshal body")
			log.Printf("The id is [%d], and the name is [%s]", payload.Id, payload.Name)
		}
	}()
	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
