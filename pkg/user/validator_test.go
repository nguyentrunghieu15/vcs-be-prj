package user

import (
	"fmt"
	"testing"

	"github.com/nguyentrunghieu15/vcs-common-prj/apu/server"
	pb "github.com/nguyentrunghieu15/vcs-common-prj/apu/user"
)

func TestUserServiceValidator_Validate(t *testing.T) {
	errMsg := "Field validation for '%v' failed on the '%v' tag"
	var negativeNum int64 = -1
	var positiveNum int64 = 1

	var OutOfMinTypeSort server.TypeSort = -1
	var OutOfMaxTypeSort server.TypeSort = 3

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
				err: fmt.Errorf("Field validation for 'Id' failed on the 'min' tag"),
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
				err: fmt.Errorf(
					"Field validation for 'Email' failed on the 'email' tag",
				),
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

		"ListUsersRequest_Nil": {
			in:       pb.ListUsersRequest{},
			expected: expectation{err: nil},
		},
		"ListUsersRequest_Limit_Negative": {
			in: pb.ListUsersRequest{
				Pagination: &server.Pagination{
					Limit: &negativeNum,
				},
			},
			expected: expectation{
				err: fmt.Errorf(errMsg, "Limit", "min"),
			},
		},
		"ListUsersRequest_Limit_Positive": {
			in: pb.ListUsersRequest{
				Pagination: &server.Pagination{
					Limit: &positiveNum,
				},
			},
			expected: expectation{
				err: nil,
			},
		},
		"ListUsersRequest_Page_Negative": {
			in: pb.ListUsersRequest{
				Pagination: &server.Pagination{
					Page: &negativeNum,
				},
			},
			expected: expectation{
				err: fmt.Errorf(errMsg, "Page", "min"),
			},
		},
		"ListUsersRequest_Page_Positive": {
			in: pb.ListUsersRequest{
				Pagination: &server.Pagination{
					Page: &positiveNum,
				},
			},
			expected: expectation{
				err: nil,
			},
		},
		"ListUsersRequest_PageSize_Negative": {
			in: pb.ListUsersRequest{
				Pagination: &server.Pagination{
					PageSize: &negativeNum,
				},
			},
			expected: expectation{
				err: fmt.Errorf(errMsg, "PageSize", "min"),
			},
		},
		"ListUsersRequest_PageSize_Positive": {
			in: pb.ListUsersRequest{
				Pagination: &server.Pagination{
					PageSize: &positiveNum,
				},
			},
			expected: expectation{
				err: nil,
			},
		},
		"ListUsersRequest_Sort_OutOfMin": {
			in: pb.ListUsersRequest{
				Pagination: &server.Pagination{
					Sort: &OutOfMinTypeSort,
				},
			},
			expected: expectation{
				err: fmt.Errorf(errMsg, "Sort", "min"),
			},
		},
		"ListUsersRequest_Sort_OutOfMax": {
			in: pb.ListUsersRequest{
				Pagination: &server.Pagination{
					Sort: &OutOfMaxTypeSort,
				},
			},
			expected: expectation{
				err: fmt.Errorf(errMsg, "Sort", "max"),
			},
		},
		"ListUsersRequest_Sort_Must_Pass": {
			in: pb.ListUsersRequest{
				Pagination: &server.Pagination{
					Sort: server.TypeSort_ASC.Enum(),
				},
			},
			expected: expectation{
				err: nil,
			},
		},

		"CreateUserRequest_Invalid_Format_Email": {
			in: &pb.CreateUserRequest{
				Email:    "dsadsadsadsa",
				Password: "dasdasdasdasd",
			},
			expected: expectation{
				err: fmt.Errorf(errMsg, "Email", "email"),
			},
		},
		"CreateUserRequest_Password_Shortest": {
			in: &pb.CreateUserRequest{
				Email:    "hieu@gmail.com",
				Password: "ss",
			},
			expected: expectation{
				err: fmt.Errorf(errMsg, "Password", "min"),
			},
		},
		"CreateUserRequest_Invalid_Format_Avatar": {
			in: &pb.CreateUserRequest{
				Email:    "hieu@gmail.com",
				Password: "dasdasdasdasd",
				Avatar:   "dsadsadsadsadsa",
			},
			expected: expectation{
				err: fmt.Errorf(errMsg, "Avatar", "base64"),
			},
		},
		"CreateUserRequest_Roles_OutOfMin": {
			in: pb.CreateUserRequest{
				Email:    "hieu@gmail.com",
				Password: "dasdasdasdasd",
				Roles:    pb.CreateUserRequest_Role(-1),
			},
			expected: expectation{
				err: fmt.Errorf(errMsg, "Roles", "min"),
			},
		},
		"CreateUserRequest_Roles_OutOfMax": {
			in: pb.CreateUserRequest{
				Email:    "hieu@gmail.com",
				Password: "dasdasdasdasd",
				Roles:    pb.CreateUserRequest_Role(3),
			},
			expected: expectation{
				err: fmt.Errorf(errMsg, "Roles", "max"),
			},
		},
		"CreateUserRequest_Must_Pass": {
			in: &pb.CreateUserRequest{
				Email:    "hieu@gmail.com",
				Password: "dasdasdasdasd",
			},
			expected: expectation{
				err: nil,
			},
		},
		"UpdateUserByIdRequest_Invalid_Format_Email": {
			in: &pb.UpdateUserByIdRequest{
				Id:    1,
				Email: "dsadsadsadsa",
			},
			expected: expectation{
				err: fmt.Errorf(errMsg, "Email", "email"),
			},
		},
		"UpdateUserByIdRequest_IdNegative": {
			in: &pb.UpdateUserByIdRequest{
				Id: -1,
			},
			expected: expectation{
				err: fmt.Errorf("Field validation for 'Id' failed on the 'min' tag"),
			},
		},
		"UpdateUserByIdRequest_Invalid_Format_Avatar": {
			in: &pb.UpdateUserByIdRequest{
				Id:     1,
				Avatar: "dsadsadsadsadsa",
			},
			expected: expectation{
				err: fmt.Errorf(errMsg, "Avatar", "base64"),
			},
		},
		"UpdateUserByIdRequest_Roles_OutOfMin": {
			in: pb.UpdateUserByIdRequest{
				Id:    1,
				Roles: pb.UpdateUserByIdRequest_Role(-1),
			},
			expected: expectation{
				err: fmt.Errorf(errMsg, "Roles", "min"),
			},
		},
		"UpdateUserByIdRequest_Roles_OutOfMax": {
			in: pb.UpdateUserByIdRequest{
				Id:    1,
				Roles: pb.UpdateUserByIdRequest_Role(3),
			},
			expected: expectation{
				err: fmt.Errorf(errMsg, "Roles", "max"),
			},
		},
		"UpdateUserByIdRequest_Must_Pass": {
			in: &pb.UpdateUserByIdRequest{
				Id: 1,
			},
			expected: expectation{
				err: nil,
			},
		},
		"DeleteUserByIdRequest_IdNegative": {
			in: &pb.DeleteUserByIdRequest{
				Id: -1,
			},
			expected: expectation{
				err: fmt.Errorf("Field validation for 'Id' failed on the 'min' tag"),
			},
		},
		"DeleteUserByIdRequest_Must_Pass": {
			in: &pb.DeleteUserByIdRequest{
				Id: 1,
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
