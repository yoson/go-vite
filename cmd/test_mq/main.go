package main

import (
	"github.com/Shopify/sarama"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"
)

func main() {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	var (
		wg                          sync.WaitGroup
		enqueued, successes, errors int
	)

	producer, err := sarama.NewAsyncProducer([]string{"118.25.228.148:9092"}, config)
	if err != nil {
		log.Panic(err)
	}

	// Trap SIGINT to trigger a shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for range producer.Successes() {
			successes++
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for err := range producer.Errors() {
			log.Println(err)
			errors++
		}
	}()

	turnInterval := time.Duration(2000)
ProducerLoop:
	for {
		time.Sleep(turnInterval * time.Millisecond)
		message := &sarama.ProducerMessage{Topic: "test", Value: sarama.StringEncoder("testing 123")}

		select {
		case producer.Input() <- message:
			enqueued++

		case <-signals:
			producer.AsyncClose() // Trigger a shutdown of the producer.
			break ProducerLoop
		}
	}
}
