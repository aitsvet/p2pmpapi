package mp

import (
	"context"

	"github.com/aitsvet/saga_example/internal/model"
	"github.com/aitsvet/saga_example/internal/service/domain"
	mp "github.com/aitsvet/saga_example/pkg/mp/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type Server struct {
	s          *domain.Service
	cancel     context.CancelFunc
	ads        map[uint64]*mp.AdResponse
	orders     map[uint64]*mp.OrderResponse
	shipments  map[uint64]*mp.ShipmentResponse
	dictionary *mp.Dictionary

	mp.UnimplementedMarketplaceServer
}

func NewServer() *Server {
	s := &Server{}
	s.s = domain.NewService("", "", "ads", "orders", "shipments")
	var ctx context.Context
	ctx, s.cancel = context.WithCancel(context.Background())

	fromDictionary := s.s.Kafka.ListenTo("dictionary")
	fromAds := s.s.Kafka.ListenTo("ads")
	fromOrders := s.s.Kafka.ListenTo("orders")
	fromShipments := s.s.Kafka.ListenTo("shipments")

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case m := <-fromDictionary:
				_ = proto.Unmarshal(m.Body, s.dictionary)
			case m := <-fromAds:
				ad := &mp.AdResponse{}
				_ = proto.Unmarshal(m.Body, ad)
				s.ads[ad.AdId] = ad
			case m := <-fromOrders:
				or := &mp.OrderResponse{}
				_ = proto.Unmarshal(m.Body, or)
				s.orders[or.OrderId] = or
			case m := <-fromShipments:
				sh := &mp.ShipmentResponse{}
				_ = proto.Unmarshal(m.Body, sh)
				s.shipments[sh.ShipmentId] = sh
			}
		}
	}()

	return s
}

func (s *Server) Close() {
	s.cancel()
}

func (s *Server) CreateAd(ctx context.Context, r *mp.AdRequest) (*mp.Empty, error) {
	s.s.Rabbit.Publish(model.NewMessage("ads", r.SellerId, r))
	return &mp.Empty{}, nil
}
func (Server) ListAds(r *mp.ListAdsRequest, s mp.Marketplace_ListAdsServer) error {
	return status.Errorf(codes.Unimplemented, "method ListAds not implemented")
}
func (s *Server) Order(ctx context.Context, r *mp.OrderRequest) (*mp.Empty, error) {
	s.s.Rabbit.Publish(model.NewMessage("orders", r.BuyerId, r))
	return &mp.Empty{}, nil
}
func (Server) ListOrders(r *mp.ListOrdersRequest, s mp.Marketplace_ListOrdersServer) error {
	return status.Errorf(codes.Unimplemented, "method ListOrders not implemented")
}
func (s *Server) Ship(ctx context.Context, r *mp.ShipmentRequest) (*mp.Empty, error) {
	s.s.Rabbit.Publish(model.NewMessage("shipments", r.SellerId, r))
	return &mp.Empty{}, nil
}
func (Server) ListShipments(r *mp.ListShipmentsRequest, s mp.Marketplace_ListShipmentsServer) error {
	return status.Errorf(codes.Unimplemented, "method ListShipmentss not implemented")
}
func (s *Server) Receive(ctx context.Context, r *mp.DeliveryRequest) (*mp.Empty, error) {
	s.s.Rabbit.Publish(model.NewMessage("shipments", r.BuyerId, r))
	return &mp.Empty{}, nil
}
func (Server) ListStatuses(ctx context.Context, r *mp.Empty) (*mp.Dictionary, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListStatuses not implemented")
}
