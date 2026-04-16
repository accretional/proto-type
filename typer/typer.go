// Package typer implements the Typer gRPC service: a name → Descriptor
// registry. It loads FileDescriptorSets (as produced by `protoc --descriptor_set_out`)
// and answers Resolve(Type{name}) with the matching Descriptor.
package typer

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	pt "github.com/accretional/proto-type"
)

// Server implements proto_type.TyperServer. Safe for concurrent use.
type Server struct {
	pt.UnimplementedTyperServer

	mu    sync.RWMutex
	types map[string]*descriptorpb.DescriptorProto // fully-qualified name → message
}

func New() *Server {
	return &Server{types: map[string]*descriptorpb.DescriptorProto{}}
}

// Register adds every message in fds to the registry. Messages are indexed
// by their fully-qualified name ("<package>.<MessageName>"), with nested
// messages flattened ("<package>.<Outer>.<Inner>"). Re-registering a name
// overwrites the previous entry.
func (s *Server) Register(fds *descriptorpb.FileDescriptorSet) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, f := range fds.GetFile() {
		pkg := f.GetPackage()
		for _, m := range f.GetMessageType() {
			s.indexMessage(pkg, m)
		}
	}
}

func (s *Server) indexMessage(prefix string, m *descriptorpb.DescriptorProto) {
	name := m.GetName()
	if prefix != "" {
		name = prefix + "." + name
	}
	s.types[name] = m
	for _, nested := range m.GetNestedType() {
		s.indexMessage(name, nested)
	}
}

// Resolve returns the Descriptor registered under t.Name.
func (s *Server) Resolve(ctx context.Context, t *pt.Type) (*pt.Descriptor, error) {
	name := t.GetName()
	if name == "" {
		return nil, status.Error(codes.InvalidArgument, "type name is empty")
	}
	s.mu.RLock()
	dp, ok := s.types[name]
	s.mu.RUnlock()
	if !ok {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("unknown type %q", name))
	}
	clone := proto.Clone(dp).(*descriptorpb.DescriptorProto)
	return &pt.Descriptor{Kind: &pt.Descriptor_Message{Message: clone}}, nil
}
