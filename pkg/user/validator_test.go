package user

import (
	"fmt"
	"testing"

	pb "github.com/nguyentrunghieu15/vcs-common-prj/apu/user"
)

func TestUserServiceValidator_Validate(t *testing.T) {
	type expectation struct {
		err error
	}

	tests := map[string]struct {
		in       interface{}
		expected expectation
	}{
		"GetUserByIdRequest_IdNegative": {
			in: &pb.GetUserByIdRequest{
				Id: -1,
			},
			expected: expectation{
				err: fmt.Errorf("Key: 'GetUserByIdRequest.Id' Error:Field validation for 'Id' failed on the 'min' tag"),
			},
		},
		"GetUserByIdRequest_Must_Pass": {
			in: &pb.GetUserByIdRequest{
				Id: 1,
			},
			expected: expectation{
				err: nil,
			},
		},
		"GetUserByEmailRequest_Invalid_Format_Email": {
			in: &pb.GetUserByEmailRequest{
				Email: "dsadsadsadsa",
			},
			expected: expectation{
				err: fmt.Errorf("Key: 'GetUserByEmailRequest.Email' Error:Field validation for 'Email' failed on the 'email' tag"),
			},
		},
		"GetUserByEmailRequest_Must_Pass": {
			in: &pb.GetUserByEmailRequest{
				Email: "hieu@gmail.com",
			},
			expected: expectation{
				err: nil,
			},
		},
	}

	userVal := NewUserServiceValidator()

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := userVal.Validate(tt.in)
			if err != nil {
				if err.Error() != tt.expected.err.Error() {
					t.Errorf("TestUserServiceValidator_Validate() error = %v, wantErr %v", err, tt.expected.err)
					return
				}
			}
		})
	}
}
