package api

import (
	"context"
	"fmt"
	"net"

	"github.com/rrb115/tfte/internal/engine"
	"github.com/rrb115/tfte/internal/proofs"
	"github.com/rrb115/tfte/internal/storage"
	"github.com/rrb115/tfte/proto/gen/tfte"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	tfte.UnimplementedTfteServiceServer
	engine *engine.Engine
	store  storage.Store // Direct access if needed, or via engine
}

func NewServer(eng *engine.Engine, store storage.Store) *Server {
	return &Server{
		engine: eng,
		store:  store,
	}
}

func (s *Server) IngestEvents(ctx context.Context, req *tfte.IngestEventsRequest) (*tfte.IngestEventsResponse, error) {
	if err := s.store.IngestEvents(ctx, req.Events); err != nil {
		return &tfte.IngestEventsResponse{Error: err.Error()}, nil
	}
	return &tfte.IngestEventsResponse{EventsProcessed: int32(len(req.Events))}, nil
}

func (s *Server) GetGraphSnapshot(ctx context.Context, req *tfte.GetGraphRequest) (*tfte.GraphSnapshot, error) {
	return s.engine.GetGraphAt(ctx, req.Timestamp)
}

func (s *Server) GetEvents(ctx context.Context, req *tfte.GetEventsRequest) (*tfte.GetEventsResponse, error) {
	return s.engine.GetEvents(ctx, req)
}

func (s *Server) GetEdgeEvidence(ctx context.Context, req *tfte.GetEdgeEvidenceRequest) (*tfte.EdgeEvidence, error) {
	return s.engine.GetEdgeEvidence(ctx, req)
}

func (s *Server) GetRootCause(ctx context.Context, req *tfte.GetRootCauseRequest) (*tfte.GetRootCauseResponse, error) {
	// 1. Get Graph at timestamp
	snapshot, err := s.engine.GetGraphAt(ctx, req.Timestamp)
	if err != nil {
		return nil, err
	}

	// 2. Identify candidates
	candidates := s.engine.GetRootCauseCandidates(snapshot)

	return &tfte.GetRootCauseResponse{Candidates: candidates}, nil
}

// GetProofBundle generates and returns a verification bundle.
func (s *Server) GetProofBundle(ctx context.Context, req *tfte.GetProofBundleRequest) (*tfte.GetProofBundleResponse, error) {
	bundleBytes, hash, err := proofs.GenerateProofBundle(ctx, s.engine, req.RootEventId, req.Timestamp)
	if err != nil {
		return nil, err
	}

	return &tfte.GetProofBundleResponse{
		BundleTarGz: bundleBytes,
		Sha256Hash:  hash,
	}, nil
}

// Start gRPC server
func RunGRPCServer(port int, srv *Server) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer()
	tfte.RegisterTfteServiceServer(grpcServer, srv)
	reflection.Register(grpcServer) // Enable CLI tools like grpcurl

	fmt.Printf("gRPC server listening on :%d\n", port)
	return grpcServer.Serve(lis)
}
