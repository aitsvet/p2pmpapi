package main

import (
	"context"
	_ "embed"
	"log"
	"net"
	"net/http"

	internal "github.com/aitsvet/saga_example/internal/service/gateway"
	pkg "github.com/aitsvet/saga_example/pkg/mp/v1"
	"github.com/flowchartsman/swaggerui"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	s := grpc.NewServer()
	pkg.RegisterMarketplaceServer(s, &internal.Server{})
	reflection.Register(s)
	lis, _ := net.Listen("tcp", ":50051")
	go s.Serve(lis)
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := pkg.RegisterMarketplaceHandlerFromEndpoint(ctx, mux, "0.0.0.0:50051", opts)
	if err != nil {
		panic(err)
	}
	err = mux.HandlePath("GET", "/*", runtime.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			swaggerui.Handler(pkg.Spec).ServeHTTP(w, r)
		}))
	if err != nil {
		panic(err)
	}
	log.Println("Starting to listen on port 8081")
	http.ListenAndServe(":8081", mux)
}
