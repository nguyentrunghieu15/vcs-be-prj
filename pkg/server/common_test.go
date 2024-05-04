package server

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	pb "github.com/nguyentrunghieu15/vcs-common-prj/apu/server"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
)

func TestConvertServerModelToServerProto(t *testing.T) {
	type expectation struct {
		out *pb.Server
		err error
	}

	tests := map[string]struct {
		in       model.Server
		expected expectation
	}{
		"Must_Pass": {
			in: model.Server{
				ID:        uuid.MustParse("96191014-0f10-4862-b37b-87b0943d2b04"),
				Name:      "Test",
				Status:    model.On,
				Ipv4:      "0.0.0.0",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				CreatedBy: 1,
				UpdatedBy: 1,
				DeletedBy: 1,
			},
			expected: expectation{
				out: &pb.Server{
					Id:        "96191014-0f10-4862-b37b-87b0943d2b04",
					Name:      "Test",
					Ipv4:      "0.0.0.0",
					Status:    pb.Server_ON,
					UpdatedBy: 1,
					DeletedBy: 1,
					CreatedBy: 1,
					CreatedAt: time.Now().Format(time.RFC3339),
					UpdatedAt: time.Now().Format(time.RFC3339),
				},
				err: nil,
			},
		},
		"Missing_status": {
			in: model.Server{
				ID:        uuid.MustParse("96191014-0f10-4862-b37b-87b0943d2b04"),
				Name:      "Test",
				Ipv4:      "0.0.0.0",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				CreatedBy: 1,
				UpdatedBy: 1,
				DeletedBy: 1,
			},
			expected: expectation{
				out: &pb.Server{
					Id:        "96191014-0f10-4862-b37b-87b0943d2b04",
					Name:      "Test",
					Ipv4:      "0.0.0.0",
					UpdatedBy: 1,
					DeletedBy: 1,
					CreatedBy: 1,
					CreatedAt: time.Now().Format(time.RFC3339),
					UpdatedAt: time.Now().Format(time.RFC3339),
				},
				err: nil,
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ConvertServerModelToServerProto(tt.in)
			if err != nil {
				if tt.expected.err == nil {
					t.Errorf("TestServerServiceValidator_Validate() error = %v, wantErr %v", err, tt.expected.err)
					return
				}
				if err.Error() != tt.expected.err.Error() {
					t.Errorf("TestServerServiceValidator_Validate() error = %v, wantErr %v", err, tt.expected.err)
					return
				}
			} else if !reflect.DeepEqual(got, tt.expected.out) {
				t.Errorf("ConvertUserModelToUserProto() = %v, want %v", got, tt.expected.out)
			}
		})
	}
}

func TestConvertListServerModelToListServerProto(t *testing.T) {
	type expectation struct {
		out []*pb.Server
		err error
	}

	tests := map[string]struct {
		in       []model.Server
		expected expectation
	}{
		"Must_Pass": {
			in: []model.Server{
				{
					ID:        uuid.MustParse("96191014-0f10-4862-b37b-87b0943d2b04"),
					Name:      "Test",
					Status:    model.On,
					Ipv4:      "0.0.0.0",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					CreatedBy: 1,
					UpdatedBy: 1,
					DeletedBy: 1,
				},
				{
					ID:        uuid.MustParse("96191014-0f10-4862-b37b-87b0943d2b04"),
					Name:      "Test",
					Status:    model.On,
					Ipv4:      "0.0.0.0",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					CreatedBy: 1,
					UpdatedBy: 1,
					DeletedBy: 1,
				},
			},
			expected: expectation{
				out: []*pb.Server{
					{
						Id:        "96191014-0f10-4862-b37b-87b0943d2b04",
						Name:      "Test",
						Ipv4:      "0.0.0.0",
						Status:    pb.Server_ON,
						UpdatedBy: 1,
						DeletedBy: 1,
						CreatedBy: 1,
						CreatedAt: time.Now().Format(time.RFC3339),
						UpdatedAt: time.Now().Format(time.RFC3339),
					},
					{
						Id:        "96191014-0f10-4862-b37b-87b0943d2b04",
						Name:      "Test",
						Ipv4:      "0.0.0.0",
						Status:    pb.Server_ON,
						UpdatedBy: 1,
						DeletedBy: 1,
						CreatedBy: 1,
						CreatedAt: time.Now().Format(time.RFC3339),
						UpdatedAt: time.Now().Format(time.RFC3339),
					},
				},
				err: nil,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ConvertListServerModelToListServerProto(tt.in)
			if err != nil && err != tt.expected.err {
				t.Errorf("ConvertListServerModelToListServerProto() error = %v, wantErr %v", err, tt.expected.err)
				return
			}
			if !reflect.DeepEqual(got, tt.expected.out) {
				t.Errorf("ConvertListServerModelToListServerProto() = %v, want %v", got, tt.expected.out)
			}
		})
	}
}

func TestValidateServerFormMap(t *testing.T) {
	keyValid := []string{"id", "name", "ipv4", "status"}
	type expectation struct {
		err error
	}

	tests := map[string]struct {
		in       map[string]interface{}
		expected expectation
	}{
		"Invalid_Key": {
			in: map[string]interface{}{
				"Test": "test",
			},
			expected: expectation{err: fmt.Errorf("Only accept key in %v", keyValid)},
		},
		"Must_Pass": {
			in: map[string]interface{}{
				"id":     "",
				"name":   "",
				"ipv4":   "",
				"status": "",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if err := ValidateServerFormMap(tt.in); err != nil && err != tt.expected.err {
				t.Errorf("ValidateServerFormMap() error = %v, wantErr %v", err, tt.expected.err)
			}
		})
	}
}
