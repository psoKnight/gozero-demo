package logic

import (
	"context"
	"math"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gozero-demo/pbdata/calculatorpb"
)

func TestAdd(t *testing.T) {
	reply, err := NewAddLogic(context.Background(), nil).Add(&calculatorpb.BinaryOperationRequest{
		Left:  21,
		Right: 21,
	})
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if got, want := reply.GetResult(), int64(42); got != want {
		t.Fatalf("Add() = %d, want %d", got, want)
	}
}

func TestMultiply(t *testing.T) {
	reply, err := NewMultiplyLogic(context.Background(), nil).Multiply(&calculatorpb.BinaryOperationRequest{
		Left:  6,
		Right: 7,
	})
	if err != nil {
		t.Fatalf("Multiply() error = %v", err)
	}
	if got, want := reply.GetResult(), int64(42); got != want {
		t.Fatalf("Multiply() = %d, want %d", got, want)
	}
}

func TestCalculatorRejectsOverflow(t *testing.T) {
	tests := []struct {
		name string
		call func() error
	}{
		{
			name: "add positive",
			call: func() error {
				_, err := NewAddLogic(context.Background(), nil).Add(&calculatorpb.BinaryOperationRequest{
					Left: math.MaxInt64, Right: 1,
				})
				return err
			},
		},
		{
			name: "add negative",
			call: func() error {
				_, err := NewAddLogic(context.Background(), nil).Add(&calculatorpb.BinaryOperationRequest{
					Left: math.MinInt64, Right: -1,
				})
				return err
			},
		},
		{
			name: "multiply",
			call: func() error {
				_, err := NewMultiplyLogic(context.Background(), nil).Multiply(&calculatorpb.BinaryOperationRequest{
					Left: math.MaxInt64, Right: 2,
				})
				return err
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := status.Code(test.call()); got != codes.OutOfRange {
				t.Fatalf("code = %v, want %v", got, codes.OutOfRange)
			}
		})
	}
}
