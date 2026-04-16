package typer

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/descriptorpb"

	pt "github.com/accretional/proto-type"
)

func TestResolve(t *testing.T) {
	s := New()

	fds := &descriptorpb.FileDescriptorSet{
		File: []*descriptorpb.FileDescriptorProto{{
			Name:    sp("test.proto"),
			Package: sp("sqlite"),
			MessageType: []*descriptorpb.DescriptorProto{
				{
					Name: sp("SelectStmt"),
					NestedType: []*descriptorpb.DescriptorProto{
						{Name: sp("Inner")},
					},
				},
				{Name: sp("AlterTable")},
			},
		}},
	}
	s.Register(fds)

	ctx := context.Background()
	got, err := s.Resolve(ctx, &pt.Type{Name: "sqlite.SelectStmt"})
	if err != nil || got.GetMessage().GetName() != "SelectStmt" {
		t.Fatalf("resolve sqlite.SelectStmt: got=%v err=%v", got, err)
	}
	if _, err := s.Resolve(ctx, &pt.Type{Name: "sqlite.SelectStmt.Inner"}); err != nil {
		t.Fatalf("resolve nested: %v", err)
	}
	_, err = s.Resolve(ctx, &pt.Type{Name: "sqlite.Nope"})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected NotFound, got %v", err)
	}
	_, err = s.Resolve(ctx, &pt.Type{Name: ""})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
}

func sp(s string) *string { return &s }
