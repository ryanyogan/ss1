package main

import (
	"context"
	"log"
	"net"
	"sync"

	"github.com/ryanyogan/shippy-service-consignment/proto/consignment"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":50051"
)

type repository interface {
	Create(*consignment.Consignment) (*consignment.Consignment, error)
}

type Repository struct {
	mu           sync.Mutex
	consignments []*consignment.Consignment
}

func (repo *Repository) Create(consignment *consignment.Consignment) (*consignment.Consignment, error) {
	repo.mu.Lock()
	updated := append(repo.consignments, consignment)
	repo.consignments = updated
	repo.mu.Unlock()
	return consignment, nil
}

type service struct {
	repo repository
}

func (s *service) CreateConsignment(ctx context.Context, req *consignment.Consignment) (*consignment.Response, error) {
	createdConsginment, err := s.repo.Create(req)
	if err != nil {
		return nil, err
	}

	return &consignment.Response{Created: true, Consignment: createdConsginment}, nil
}

func main() {
	repo := &Repository{}

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	consignment.RegisterShippingServiceServer(s, &service{repo})

	reflection.Register(s)

	log.Println("Running on port:", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
