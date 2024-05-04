package server

import (
	"testing"

	"github.com/segmentio/kafka-go"
)

func TestNewProducer(t *testing.T) {
	type expectation struct {
		out *ProducerKafka
	}

	tests := map[string]struct {
		in       ConfigProducer
		expected expectation
	}{
		// TODO: Add test cases.
		"Must_Pass": {
			in: ConfigProducer{
				Broker:   []string{"sds"},
				Topic:    "sd",
				Balancer: &kafka.LeastBytes{},
			},
			expected: expectation{out: &ProducerKafka{
				k: kafka.NewWriter(
					kafka.WriterConfig{
						Brokers:  []string{"sds"},
						Topic:    "sd",
						Balancer: &kafka.LeastBytes{},
					},
				),
			}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := NewProducer(tt.in); false {
				t.Errorf("NewProducer() = %v, want %v", got, tt.expected.out)
			}
		})
	}
}
