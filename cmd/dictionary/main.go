package main

import (
	"time"

	"github.com/aitsvet/saga_example/internal/model"
	"github.com/aitsvet/saga_example/internal/service/domain"
	"github.com/aitsvet/saga_example/pkg/mp/v1"
	_ "github.com/lib/pq"
	"google.golang.org/protobuf/proto"
)

var migration = `
DROP TABLE dictionary;
CREATE TABLE dictionary (
	id SERIAL PRIMARY KEY,
	code TEXT,
	title TEXT
);
INSERT INTO dictionary(code, title) VALUES
	('UNKNOWN', 'Неизвестен'),
	('NEW', 'Новый'),
	('CONFIRMED', 'Подтвержден'),
	('DECLINED', 'Отклонен'),
	('PROCESSING', 'В пути'),
	('DELIVERED', 'Доставлен'),
	('AD', 'Объявление'),
	('DEAL', 'Сделка'),
	('SHIPPING', 'Отправление');
`

func main() {

	s := domain.NewService("dictionary", migration, "dictionary")
	defer s.Close()

	dict := &mp.Dictionary{}
	res, _ := s.DB.Query("SELECT id, code, title FROM dictionary")
	for res.Next() {
		e := &mp.Dictionary_Entry{}
		res.Scan(&e.Id, &e.Code, &e.Title)
		dict.Entries = append(dict.Entries, e)
	}
	body, _ := proto.Marshal(dict)

	time.Sleep(5 * time.Second)
	s.Kafka.Publish(&model.Message{
		Queue: "dictionary",
		TxId:  0,
		Body:  body,
	})
}
