package server

import (
	"reflect"
	"testing"

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
				ID:     uuid.MustParse("96191014-0f10-4862-b37b-87b0943d2b04"),
				Name:   "Test",
				Status: model.On,
				Ipv4:   "0.0.0.0",
			},
			expected: expectation{
				out: &pb.Server{
					Id:     "96191014-0f10-4862-b37b-87b0943d2b04",
					Name:   "Test",
					Ipv4:   "0.0.0.0",
					Status: pb.Server_ON,
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
