package main

import (
	"github.com/aitsvet/saga_example/internal/model"
	"github.com/aitsvet/saga_example/internal/service/domain"
	"github.com/aitsvet/saga_example/pkg/mp/v1"
	_ "github.com/lib/pq"
	"google.golang.org/protobuf/proto"
)

var migration = `
DROP TABLE order;
CREATE TABLE order (
	order_id SERIAL PRIMARY KEY,
	ad_id BIGINT,
	buyer_id BIGINT,
	status SMALLINT
);
`

var insertOrder = "INSERT INTO order(ad_id, buyer_id, status) VALUES (?, ?, ?) RETURNING order_id"
var updateShipment = "UPDATE order SET status = ? WHERE order_id = ? RETURNING ad_id, buyer_id"
var updateAd = "UPDATE order SET status = ? WHERE ad_id = ? RETURNING order_id, buyer_id"

func main() {

	s := domain.NewService("orders", migration, "ads", "shipments")
	defer s.Close()

	fromGateway := s.Rabbit.ListenTo("orders")
	fromAd := s.Kafka.ListenTo("ads")
	fromShipment := s.Kafka.ListenTo("shipments")

	for {
		o := &mp.OrderResponse{}
		m := &model.Message{}
		select {
		case m = <-fromGateway:
			_ = proto.Unmarshal(m.Body, o.Props)
			res := s.DB.QueryRow(insertOrder, o.Props.AdId, o.Props.BuyerId, 1)
			res.Scan(o.OrderId)
		case m = <-fromShipment:
			r := &mp.ShipmentResponse{}
			_ = proto.Unmarshal(m.Body, r)
			res := s.DB.QueryRow(updateShipment, r.Status, r.OrderId)
			o.Status = r.Status
			o.OrderId = r.OrderId
			res.Scan(&o.Props.AdId, &o.Props.BuyerId)
		case m = <-fromAd:
			r := &mp.AdResponse{}
			_ = proto.Unmarshal(m.Body, r)
			res := s.DB.QueryRow(updateAd, r.Status, r.AdId)
			o.Status = r.Status
			o.Props.AdId = r.AdId
			res.Scan(&o.OrderId, &o.Props.BuyerId)
		}
		body, _ := proto.Marshal(o)
		s.Kafka.Publish(&model.Message{
			Queue: "orders",
			TxId:  m.TxId,
			Body:  body,
		})
	}
}
