package main

import (
	"github.com/aitsvet/saga_example/internal/model"
	"github.com/aitsvet/saga_example/internal/service/domain"
	"github.com/aitsvet/saga_example/pkg/mp/v1"
	_ "github.com/lib/pq"
	"google.golang.org/protobuf/proto"
)

var migration = `
DROP TABLE ad;
CREATE TABLE ad (
	ad_id SERIAL PRIMARY KEY,
	product_name VARCHAR(255),
	price BIGINT,
	seller_id BIGINT,
	status SMALLINT
);
`

var insertAd = "INSERT INTO ad(product_name, price, seller_id, status) VALUES (?, ?, ?, ?) RETURNING ad_id, status"
var updateOrder = "UPDATE ad SET status = ? WHERE ad_id = ? RETURNING product_name, price, seller_id"

func main() {

	s := domain.NewService("ads", migration, "orders")
	defer s.Close()

	fromGateway := s.Rabbit.ListenTo("ads")
	fromOrder := s.Kafka.ListenTo("orders")

	for {
		o := &mp.AdResponse{}
		m := &model.Message{}
		select {
		case m = <-fromGateway:
			_ = proto.Unmarshal(m.Body, o.Props)
			res := s.DB.QueryRow(insertAd, o.Props.ProductName, o.Props.Price, o.Props.SellerId, 1)
			res.Scan(&o.AdId, &o.Status)
		case m = <-fromOrder:
			r := &mp.OrderResponse{}
			_ = proto.Unmarshal(m.Body, r)
			res := s.DB.QueryRow(updateOrder, r.Status, r.Props.AdId)
			o.AdId = r.Props.AdId
			res.Scan(&o.Props.ProductName, &o.Props.Price, &o.Props.SellerId)
		}
		body, _ := proto.Marshal(o)
		s.Kafka.Publish(&model.Message{
			Queue: "ads",
			TxId:  m.TxId,
			Body:  body,
		})
	}
}
