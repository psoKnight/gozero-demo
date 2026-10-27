package logic

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gozero-demo/pbdata/greeterpb"
)

func TestSayHello(t *testing.T) {
	reply, err := NewSayHelloLogic(context.Background(), nil).SayHello(&greeterpb.HelloRequest{
		Name: " GoFrame ",
	})
	if err != nil {
		t.Fatalf("SayHello() error = %v", err)
	}
	if got, want := reply.GetMessage(), "Hello, GoFrame! (Greeter)"; got != want {
		t.Fatalf("SayHello() = %q, want %q", got, want)
	}
}

func TestSayHelloRequiresName(t *testing.T) {
	_, err := NewSayHelloLogic(context.Background(), nil).SayHello(&greeterpb.HelloRequest{})
	if got := status.Code(err); got != codes.InvalidArgument {
		t.Fatalf("SayHello() code = %v, want %v", got, codes.InvalidArgument)
	}
}
