package logic

import (
	"encoding/json"
	"math"
	"testing"
)

func TestParseInt64Field(t *testing.T) {
	tests := []struct {
		name    string
		raw     json.RawMessage
		want    int64
		wantErr bool
	}{
		{name: "number", raw: json.RawMessage("21"), want: 21},
		{name: "quoted", raw: json.RawMessage(`"21"`), want: 21},
		{name: "max int64", raw: json.RawMessage("9223372036854775807"), want: math.MaxInt64},
		{name: "min int64", raw: json.RawMessage(`"-9223372036854775808"`), want: math.MinInt64},
		{name: "empty", raw: nil, wantErr: true},
		{name: "decimal", raw: json.RawMessage("21.5"), wantErr: true},
		{name: "overflow", raw: json.RawMessage("9223372036854775808"), wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseInt64Field(test.raw)
			if (err != nil) != test.wantErr {
				t.Fatalf("parseInt64Field() error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && got != test.want {
				t.Fatalf("parseInt64Field() = %d, want %d", got, test.want)
			}
		})
	}
}
