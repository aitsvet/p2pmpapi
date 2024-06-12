package domain

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/aitsvet/saga_example/internal/db"
	"github.com/aitsvet/saga_example/internal/kafka"
	"github.com/aitsvet/saga_example/internal/rabbit"
)

type Service struct {
	DB     *sql.DB
	Rabbit *rabbit.Client
	Kafka  *kafka.Client
}

func NewService(name string, migration string, listening ...string) *Service {
	var err error
	s := &Service{}
	s.DB, err = db.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	_, err = s.DB.Exec(migration)
	if err != nil {
		log.Fatal(err)
	}
	s.Rabbit, err = rabbit.Connect(name)
	if err != nil {
		log.Fatal("Failed to create RabbitMQ client:", err)
	}
	s.Kafka, err = kafka.Connect(listening...)
	if err != nil {
		log.Fatal("Failed to create Kafka client:", err)
	}
	fmt.Println(name + " service started")
	return s
}

func (s *Service) Close() {
	if s.DB != nil {
		_ = s.DB.Close()
	}
	if s.Rabbit != nil {
		s.Rabbit.Close()
	}
	if s.Kafka != nil {
		s.Kafka.Close()
	}
}
