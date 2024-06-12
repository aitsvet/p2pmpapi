package main

import (
	"github.com/aitsvet/saga_example/internal/model"
	"github.com/aitsvet/saga_example/internal/service/domain"
	"github.com/aitsvet/saga_example/pkg/mp/v1"
	_ "github.com/lib/pq"
	"google.golang.org/protobuf/proto"
)

var migration = `
DROP TABLE shipment;
CREATE TABLE shipment (
	shipment_id SERIAL PRIMARY KEY,
	order_id BIGINT,
	status SMALLINT
);
`

var insertShipment = "INSERT INTO shipment(order_id, status) VALUES (?, ?) RETURNING shipment_id"
var updateShipment = "UPDATE shipment SET status = ? WHERE shipment_id = ? RETURNING order_id"
var updateOrder = "UPDATE shipment SET status = ? WHERE order_id = ? RETURNING shipment_id"

func main() {

	s := domain.NewService("shipments", migration, "orders")
	defer s.Close()

	fromGateway := s.Rabbit.ListenTo("shipments")
	fromOrder := s.Kafka.ListenTo("orders")

	for {
		o := &mp.ShipmentResponse{}
		m := &model.Message{}
		select {
		case m = <-fromGateway:
			r := &mp.ShipmentRequest{}
			err := proto.Unmarshal(m.Body, r)
			if err != nil {
				o.Status = 1
				o.OrderId = r.OrderId
				s.DB.QueryRow(updateShipment, o.OrderId, o.Status).Scan(&o.ShipmentId)
			} else {
				r := &mp.DeliveryRequest{}
				_ = proto.Unmarshal(m.Body, r)
				o.Status = 3
				o.ShipmentId = r.ShipmentId
				s.DB.QueryRow(updateShipment, o.Status, o.ShipmentId).Scan(&o.OrderId)
			}
		case m = <-fromOrder:
			r := &mp.OrderResponse{}
			_ = proto.Unmarshal(m.Body, r)
			o.Status = r.Status
			o.OrderId = r.OrderId
			s.DB.QueryRow(updateOrder, o.Status, o.OrderId).Scan(&o.ShipmentId)
		}
		body, _ := proto.Marshal(o)
		s.Kafka.Publish(&model.Message{
			Queue: "shipments",
			TxId:  m.TxId,
			Body:  body,
		})
	}
}
