package main

import (
	"context"
	"fmt"
	"log"

	"github.com/micro/go-micro"
	pb "github.com/ryanyogan/shippy-service-consignment/proto/consignment"
	vesselProto "github.com/ryanyogan/shippy-service-vessel/proto/vessel"
)

type repository interface {
	Create(*pb.Consignment) (*pb.Consignment, error)
	GetAll() []*pb.Consignment
}

// Repository - Dummy repo, this simulates the use of a
// datastore.
type Repository struct {
	consignments []*pb.Consignment
}

// Create a new consignment
func (repo *Repository) Create(consignment *pb.Consignment) (*pb.Consignment, error) {
	updated := append(repo.consignments, consignment)
	repo.consignments = updated
	return consignment, nil
}

// GetAll consignments
func (repo *Repository) GetAll() []*pb.Consignment {
	return repo.consignments
}

// Service should implement all of the methods to satisfy the
// service we defined in our protobuf definition.  You can
// check the interface in the generated code for the exact
// method signatures, etc.
type service struct {
	repo         repository
	vesselClient vesselProto.VesselServiceClient
}

// CreateConsignment - we created just one method on our service.
// which is a create method, which takes a context and a request
// as arguments, these are handled by the gRPC server.
func (s *service) CreateConsignment(ctx context.Context, req *pb.Consignment, res *pb.Response) error {

	// Call a client instance of our vessel service with our consignment weight,
	// and the amount of containers as the capacity value.
	vesselResponse, err := s.vesselClient.FindAvailable(context.Background(), &vesselProto.Specification{
		MaxWeight: req.Weight,
		Capacity:  int32(len(req.Containers)),
	})
	if err != nil {
		return err
	}
	log.Printf("Found a vessel! // %s \n", vesselResponse.Vessel.Name)

	// Set the VesselID as the vessel we got back from our
	// vessel service
	req.VesselId = vesselResponse.Vessel.Id

	// save our consignment
	consignment, err := s.repo.Create(req)
	if err != nil {
		return err
	}

	res.Created = true
	res.Consignment = consignment
	return nil
}

// GetConsignments - returns all consignments, non-filtered, non-paginated
func (s *service) GetConsignments(ctx context.Context, req *pb.GetRequest, res *pb.Response) error {
	consignments := s.repo.GetAll()
	res.Consignments = consignments
	return nil
}

func main() {
	repo := &Repository{}

	// Create a new service.  Optionally inclusde some options
	srv := micro.NewService(
		// This name must match the package name given in proto-buf def
		micro.Name("shippy.service.consignment"),
	)

	// Init will parse the CLI flags
	srv.Init()

	vesselClient := vesselProto.NewVesselServiceClient("shippy.service.vessel", srv.Client())

	// Register the handler
	pb.RegisterShippingServiceHandler(srv.Server(), &service{repo, vesselClient})

	// Run the server
	if err := srv.Run(); err != nil {
		fmt.Println(err)
	}
}
