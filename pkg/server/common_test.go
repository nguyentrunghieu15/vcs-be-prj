package server

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	pb "github.com/nguyentrunghieu15/vcs-common-prj/apu/server"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
	"github.com/segmentio/kafka-go"
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
		"Missing_date": {
			in: model.Server{
				ID:   uuid.MustParse("96191014-0f10-4862-b37b-87b0943d2b04"),
				Name: "Test",
				Ipv4: "0.0.0.0",
			},
			expected: expectation{
				out: &pb.Server{
					Id:   "96191014-0f10-4862-b37b-87b0943d2b04",
					Name: "Test",
					Ipv4: "0.0.0.0",
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
				"id":     "96191014-0f10-4862-b37b-87b0943d2b04",
				"name":   "Test",
				"ipv4":   "0.0.0.0",
				"status": "on",
			},
			expected: expectation{err: nil},
		},
		"Invalid_Ip_Address": {
			in: map[string]interface{}{
				"id":     "96191014-0f10-4862-b37b-87b0943d2b04",
				"name":   "Test",
				"ipv4":   "0.0.0.sadsa0",
				"status": "on",
			},
			expected: expectation{err: fmt.Errorf("IP address is not valid")},
		},
		"Invalid_Ipv4": {
			in: map[string]interface{}{
				"id":     "96191014-0f10-4862-b37b-87b0943d2b04",
				"name":   "Test",
				"ipv4":   "2001:db8:3333:4444:5555:6666:7777:8888",
				"status": "on",
			},
			expected: expectation{err: fmt.Errorf("IP address is not v4")},
		},
		"Invalid_Status": {
			in: map[string]interface{}{
				"id":     "96191014-0f10-4862-b37b-87b0943d2b04",
				"name":   "Test",
				"ipv4":   "0.0.0.0",
				"status": "osada",
			},
			expected: expectation{err: fmt.Errorf("Status is valid")},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if err := ValidateServerFormMap(tt.in); err != nil && err.Error() != tt.expected.err.Error() {
				t.Errorf("ValidateServerFormMap() error = %v, wantErr %v", err, tt.expected.err)
			}
		})
	}
}

func TestConvertStatusServerModelToStatusServerProto(t *testing.T) {
	type expectation struct {
		out pb.Server_ServerStatus
	}

	tests := map[string]struct {
		in       model.ServerStatus
		expected expectation
	}{
		"On": {
			in:       model.On,
			expected: expectation{out: pb.Server_ON},
		},
		"Off": {
			in:       model.Off,
			expected: expectation{out: pb.Server_OFF},
		},
		"nil": {
			in:       "",
			expected: expectation{out: pb.Server_NONE},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := ConvertStatusServerModelToStatusServerProto(tt.in); !reflect.DeepEqual(got, tt.expected.out) {
				t.Errorf("ConvertStatusServerModelToStatusServerProto() = %v, want %v", got, tt.expected.out)
			}
		})
	}
}

func TestParseExportRequestToKafkaMessage(t *testing.T) {
	type expectation struct {
		out *kafka.Message
		err error
	}

	jsonReq, e := json.Marshal(&pb.ExportServerRequest{
		UserId: 1,
		File: &pb.FileExport{
			FileName: "Test.xlsx",
		}})

	tests := map[string]struct {
		in       *pb.ExportServerRequest
		expected expectation
	}{
		"Must_pass": {
			in: &pb.ExportServerRequest{
				UserId: 1,
				File: &pb.FileExport{
					FileName: "Test.xlsx",
				},
			},
			expected: expectation{
				out: &kafka.Message{
					Key:   []byte("Test.xlsx"),
					Value: jsonReq,
				},
				err: e,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ParseExportRequestToKafkaMessage(tt.in)
			if err != nil && err.Error() != tt.expected.err.Error() {
				t.Errorf("ParseExportRequestToKafkaMessage() error = %v, wantErr %v", err, tt.expected.err)
				return
			}
			if !reflect.DeepEqual(got, tt.expected.out) {
				t.Errorf("ParseExportRequestToKafkaMessage() = %v, want %v", got, tt.expected.out)
			}
		})
	}
}

func TestParseMapCreateServerRequest(t *testing.T) {
	type expectation struct {
		out map[string]interface{}
		err error
	}

	tests := map[string]struct {
		in       *pb.CreateServerRequest
		expected expectation
	}{
		// TODO: Add test cases.

		"Must_Pass_status_on": {
			in: &pb.CreateServerRequest{
				Name:   "Test1",
				Status: pb.CreateServerRequest_ON,
				Ipv4:   "0.0.0.0",
			},
			expected: expectation{
				out: map[string]interface{}{
					"status": model.On,
					"name":   "Test1",
					"ipv4":   "0.0.0.0",
				},
				err: nil,
			},
		},
		"Must_Pass_status_off": {
			in: &pb.CreateServerRequest{
				Name:   "Test1",
				Status: pb.CreateServerRequest_OFF,
				Ipv4:   "0.0.0.0",
			},
			expected: expectation{
				out: map[string]interface{}{
					"status": model.Off,
					"name":   "Test1",
					"ipv4":   "0.0.0.0",
				},
				err: nil,
			},
		},
		"Missing_status": {
			in: &pb.CreateServerRequest{
				Name: "Test1",
				Ipv4: "0.0.0.0",
			},
			expected: expectation{
				out: map[string]interface{}{
					"name": "Test1",
					"ipv4": "0.0.0.0",
				},
				err: nil,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ParseMapCreateServerRequest(tt.in)
			if (err != nil) && err.Error() != tt.expected.err.Error() {
				t.Errorf("ParseMapCreateServerRequest() error = %v, wantErr %v", err, tt.expected.err)
				return
			}
			if !reflect.DeepEqual(got, tt.expected.out) {
				t.Errorf("ParseMapCreateServerRequest() = %v, want %v", got, tt.expected.out)
			}
		})
	}
}

func TestParseMapUpdateServerRequest(t *testing.T) {
	type expectation struct {
		out map[string]interface{}
		err error
	}

	tests := map[string]struct {
		in       *pb.UpdateServerRequest
		expected expectation
	}{
		// TODO: Add test cases.

		"Must_Pass_status_on": {
			in: &pb.UpdateServerRequest{
				Name:   "Test1",
				Status: pb.UpdateServerRequest_ON,
				Ipv4:   "0.0.0.0",
			},
			expected: expectation{
				out: map[string]interface{}{
					"status": model.On,
					"name":   "Test1",
					"ipv4":   "0.0.0.0",
				},
				err: nil,
			},
		},
		"Must_Pass_status_off": {
			in: &pb.UpdateServerRequest{
				Name:   "Test1",
				Status: pb.UpdateServerRequest_OFF,
				Ipv4:   "0.0.0.0",
			},
			expected: expectation{
				out: map[string]interface{}{
					"status": model.Off,
					"name":   "Test1",
					"ipv4":   "0.0.0.0",
				},
				err: nil,
			},
		},
		"Missing_status": {
			in: &pb.UpdateServerRequest{
				Name: "Test1",
				Ipv4: "0.0.0.0",
			},
			expected: expectation{
				out: map[string]interface{}{
					"name": "Test1",
					"ipv4": "0.0.0.0",
				},
				err: nil,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ParseMapUpdateServerRequest(tt.in)
			if (err != nil) && err.Error() != tt.expected.err.Error() {
				t.Errorf("ParseMapUpdateServerRequest() error = %v, wantErr %v", err, tt.expected.err)
				return
			}
			if !reflect.DeepEqual(got, tt.expected.out) {
				t.Errorf("ParseMapUpdateServerRequest() = %v, want %v", got, tt.expected.out)
			}
		})
	}
}

func TestConvertServerModelMapToServerProto(t *testing.T) {
	type expectation struct {
		out *pb.Server
		err error
	}

	tests := map[string]struct {
		in       map[string]interface{}
		expected expectation
	}{
		// TODO: Add test cases.
		"Must_pass": {
			in: map[string]interface{}{
				"id":     "96191014-0f10-4862-b37b-87b0943d2b04",
				"name":   "Test",
				"ipv4":   "0.0.0.0",
				"status": "on",
			},
			expected: expectation{
				err: nil,
				out: &pb.Server{
					Id:     "96191014-0f10-4862-b37b-87b0943d2b04",
					Name:   "Test",
					Ipv4:   "0.0.0.0",
					Status: pb.Server_ON,
				},
			},
		},
		"Misssing_Status": {
			in: map[string]interface{}{
				"id":   "96191014-0f10-4862-b37b-87b0943d2b04",
				"name": "Test",
				"ipv4": "0.0.0.0",
			},
			expected: expectation{
				err: nil,
				out: &pb.Server{
					Id:   "96191014-0f10-4862-b37b-87b0943d2b04",
					Name: "Test",
					Ipv4: "0.0.0.0",
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ConvertServerModelMapToServerProto(tt.in)
			if (err != nil) && err != tt.expected.err {
				t.Errorf("ConvertServerModelMapToServerProto() error = %v, wantErr %v", err, tt.expected.err)
				return
			}
			if !reflect.DeepEqual(got, tt.expected.out) {
				t.Errorf("ConvertServerModelMapToServerProto() = %v, want %v", got, tt.expected.out)
			}
		})
	}
}

func TestConvertListServerModelMapToListServerProto(t *testing.T) {
	type expectation struct {
		out []*pb.Server
		err error
	}

	tests := map[string]struct {
		in       []map[string]interface{}
		expected expectation
	}{
		"Must_pass": {
			in: []map[string]interface{}{
				{
					"id":     "96191014-0f10-4862-b37b-87b0943d2b04",
					"name":   "Test",
					"ipv4":   "0.0.0.0",
					"status": "on",
				},
			},
			expected: expectation{
				err: nil,
				out: []*pb.Server{
					&pb.Server{
						Id:     "96191014-0f10-4862-b37b-87b0943d2b04",
						Name:   "Test",
						Ipv4:   "0.0.0.0",
						Status: pb.Server_ON,
					},
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ConvertListServerModelMapToListServerProto(tt.in)
			if (err != nil) && err != tt.expected.err {
				t.Errorf("ConvertListServerModelMapToListServerProto() error = %v, wantErr %v", err, tt.expected.err)
				return
			}
			if !reflect.DeepEqual(got, tt.expected.out) {
				t.Errorf("ConvertListServerModelMapToListServerProto() = %v, want %v", got, tt.expected.out)
			}
		})
	}
}
