package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/elaugier/ApiGo/pkg/apigoprocessor"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/elaugier/ApiGo/pkg/apigoconfig"
	"github.com/kardianos/osext"
)

func main() {

	fullBinaryName, err := osext.Executable()
	if err != nil {
		log.Fatal(err)
	}

	folderPath, err := osext.ExecutableFolder()
	if err != nil {
		log.Fatal(err)
	}

	binaryName := strings.Replace(strings.Replace(fullBinaryName, folderPath, "", -1), string(os.PathSeparator), "", -1)

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile | log.LUTC)
	log.SetPrefix(binaryName + " " + strconv.Itoa(os.Getpid()) + " ")

	config, err := apigoconfig.Get()
	if err != nil {
		log.Fatal(err)
	}

	timestampStart := strconv.FormatInt(time.Time.UnixNano(time.Now()), 10)
	logFile := os.ExpandEnv(config.GetString("logFolder")) + "/" + timestampStart + "_" + binaryName + ".log"
	log.Println("log file location => '" + logFile + "'")
	f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	multi := io.MultiWriter(f, os.Stdout)
	log.SetOutput(multi)

	broker := config.GetString("KafkaConsumer.BootstrapServers")
	group := config.GetString("KafkaConsumer.GroupId")
	debug := config.GetString("KafkaConsumer.Debug")
	topics := []string{config.GetString("WorkerTopic")}

	m := config.GetString("MaxConcurrentJobs")
	maxConcurrentJobs, err := strconv.ParseInt(m, 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":    broker,
		"group.id":             group,
		"debug":                debug,
		"session.timeout.ms":   6000,
		"default.topic.config": kafka.ConfigMap{"auto.offset.reset": "earliest"}})

	if err != nil {
		log.Fatalf("Failed to create consumer: %s\n", err)
	}

	log.Printf("Created Consumer %v\n", c)

	err = c.SubscribeTopics(topics, nil)

	run := true

	P := apigoprocessor.NewProcessor()

	done := make(chan string, 1)

	for run == true {
		select {
		case sig := <-sigchan:

			log.Printf("Caught signal %v: terminating\n", sig)
			run = false

		default:
			ev := c.Poll(100)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:

				log.Printf("%% Message on %s:\n%s\n",
					e.TopicPartition, string(e.Value))
				if e.Headers != nil {
					log.Printf("%% Headers: %v\n", e.Headers)
				}
				if maxConcurrentJobs == 0 || P.GetCurrentJobsCount() < maxConcurrentJobs {
					go P.Process(e, done)
				} else {
					<-done
				}

			case kafka.PartitionEOF:

				log.Printf("%% Reached %v\n", e)

			case kafka.Error:

				log.Fatalf("%% Error: %v\n", e)
				run = false

			default:

				log.Printf("Ignored %v\n", e)

			}
		}
	}

	log.Printf("Closing consumer\n")
	c.Close()
}
