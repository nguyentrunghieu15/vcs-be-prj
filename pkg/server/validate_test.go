package server

import (
	"fmt"
	"testing"

	pb "github.com/nguyentrunghieu15/vcs-common-prj/apu/server"
)

func TestServerServiceValidator_Validate(t *testing.T) {
	errMsg := "Field validation for '%v' failed on the '%v' tag"
	// var negativeNum int64 = -1
	// var positiveNum int64 = 1

	// var OutOfMinTypeSort pb.TypeSort = -1
	// var OutOfMaxTypeSort pb.TypeSort = 3

	timeTemp := "2011-10-05T14:48:00.000Z"

	type expectation struct {
		err error
	}

	tests := map[string]struct {
		in       interface{}
		expected expectation
	}{
		"CreateServerRequest_Must_Pass": {
			in: pb.CreateServerRequest{
				Name:   "Test",
				Status: pb.CreateServerRequest_ON,
				Ipv4:   "0.0.0.0",
			},
			expected: expectation{
				err: nil,
			},
		},
		"CreateServerRequest_Invalid_Ipv4": {
			in: pb.CreateServerRequest{
				Name:   "Test",
				Status: pb.CreateServerRequest_ON,
				Ipv4:   "0.0.",
			},
			expected: expectation{
				err: fmt.Errorf(errMsg, "Ipv4", "ipv4"),
			},
		},
		"CreateServerRequest_Invalid_OutOfRangeMinStatus": {
			in: pb.CreateServerRequest{
				Name:   "Test",
				Status: pb.CreateServerRequest_ServerStatus(-1),
				Ipv4:   "0.0.0.0",
			},
			expected: expectation{
				err: fmt.Errorf(errMsg, "Status", "min"),
			},
		},
		"CreateServerRequest_Invalid_OutOfRangeMaxStatus": {
			in: pb.CreateServerRequest{
				Name:   "Test",
				Status: pb.CreateServerRequest_ServerStatus(3),
				Ipv4:   "0.0.0.0",
			},
			expected: expectation{
				err: fmt.Errorf(errMsg, "Status", "min"),
			},
		},
		"CreateServerRequest_Missing": {
			in: pb.CreateServerRequest{
				Status: pb.CreateServerRequest_ServerStatus(3),
				Ipv4:   "0.0.0.0",
			},
			expected: expectation{
				err: fmt.Errorf(errMsg, "Name", "required"),
			},
		},
		"DeleteServerByIdRequest_Must_Pass": {
			in: pb.DeleteServerByIdRequest{
				Id: "96191014-0f10-4862-b37b-87b0943d2b04",
			},
			expected: expectation{
				err: nil,
			},
		},
		"DeleteServerByIdRequest_Invalid": {
			in: pb.DeleteServerByIdRequest{
				Id: "sda",
			},
			expected: expectation{
				err: fmt.Errorf(errMsg, "Id", "uuid4"),
			},
		},
		"DeleteServerByNameRequest_Must_Pass": {
			in: pb.DeleteServerByNameRequest{
				Name: "dsad",
			},
			expected: expectation{err: nil},
		},
		"DeleteServerByNameRequest_Missing": {
			in: pb.DeleteServerByNameRequest{},
			expected: expectation{
				err: fmt.Errorf(errMsg, "Name", "required"),
			},
		},
		"ExportServerRequest_Invalid": {
			in: pb.ExportServerRequest{
				UserId: int64(1),
				File: &pb.FileExport{
					FileName: "Test.xlsx",
				},
				Filter: &pb.FilterServer{
					CreatedAtFrom: &timeTemp,
				},
			},
			expected: expectation{err: fmt.Errorf("S")},
		},
		"GetServerByIdRequest_Must_Pass": {
			in: pb.GetServerByIdRequest{
				Id: "96191014-0f10-4862-b37b-87b0943d2b04",
			},
			expected: expectation{err: nil},
		},
		"GetServerByNameRequest_Must_Pass": {
			in:       pb.GetServerByNameRequest{Name: "Ds"},
			expected: expectation{err: nil},
		},
		"ListServerRequest_Must_Pass": {
			in:       pb.ListServerRequest{},
			expected: expectation{err: nil},
		},
		"UpdateServerRequest_Must_Pass": {
			in: pb.UpdateServerRequest{
				Id: "96191014-0f10-4862-b37b-87b0943d2b04",
			},
			expected: expectation{err: nil},
		},
	}

	serverVal := NewServerServiceValidator()

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := serverVal.Validate(tt.in)
			if err != nil {
				if err.Error() != tt.expected.err.Error() {
					t.Errorf("TestServerServiceValidator_Validate() error = %v, wantErr %v", err, tt.expected.err)
					return
				}
			}
		})
	}
}
