package user

import (
	"reflect"
	"testing"
	"time"

	"github.com/nguyentrunghieu15/vcs-common-prj/apu/user"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
)

func TestParseMapCreateUserRequest(t *testing.T) {
	type expectation struct {
		out map[string]interface{}
		err error
	}

	tests := map[string]struct {
		in       *user.CreateUserRequest
		expected expectation
	}{
		"Must_Pass": {
			in: &user.CreateUserRequest{
				Email:    "hieu@gmail.com",
				FullName: "Nguyen Trung Hieu",
				Phone:    "0932432423",
				Avatar:   "a",
				Password: "d",
				Roles:    user.CreateUserRequest_user,
			},
			expected: expectation{
				out: map[string]interface{}{
					"email":     "hieu@gmail.com",
					"full_name": "Nguyen Trung Hieu",
					"phone":     "0932432423",
					"avatar":    "a",
					"password":  "d",
					"roles":     model.RoleUser,
				},
				err: nil,
			},
		},
		"Missing_Roles_Must_Pass": {
			in: &user.CreateUserRequest{
				Email:    "hieu@gmail.com",
				FullName: "Nguyen Trung Hieu",
				Phone:    "0932432423",
				Avatar:   "a",
				Password: "d",
			},
			expected: expectation{
				out: map[string]interface{}{
					"email":     "hieu@gmail.com",
					"full_name": "Nguyen Trung Hieu",
					"phone":     "0932432423",
					"avatar":    "a",
					"password":  "d",
				},
				err: nil,
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ParseMapCreateUserRequest(tt.in)
			if err != nil {
				if err != tt.expected.err {
					t.Errorf("ParseMapCreateUserRequest() error = %v, wantErr %v", err, tt.expected.err)
					return
				}
			} else if !reflect.DeepEqual(got, tt.expected.out) {
				t.Errorf("ParseMapCreateUserRequest() = %v, want %v", got, tt.expected.out)
			}
		})
	}
}

func TestParseMapUpdateUserRequest(t *testing.T) {
	type expectation struct {
		out map[string]interface{}
		err error
	}

	tests := map[string]struct {
		in       *user.UpdateUserByIdRequest
		expected expectation
	}{
		"Must_Pass": {
			in: &user.UpdateUserByIdRequest{
				Id:       1,
				Email:    "hieu@gmail.com",
				FullName: "Nguyen Trung Hieu",
				Phone:    "0932432423",
				Avatar:   "a",
				Roles:    user.UpdateUserByIdRequest_user,
			},
			expected: expectation{
				out: map[string]interface{}{
					"email":     "hieu@gmail.com",
					"full_name": "Nguyen Trung Hieu",
					"phone":     "0932432423",
					"avatar":    "a",
					"roles":     model.RoleUser,
				},
				err: nil,
			},
		},
		"Missing_Roles_Must_Pass": {
			in: &user.UpdateUserByIdRequest{
				Id:       1,
				Email:    "hieu@gmail.com",
				FullName: "Nguyen Trung Hieu",
				Phone:    "0932432423",
				Avatar:   "a",
			},
			expected: expectation{
				out: map[string]interface{}{
					"email":     "hieu@gmail.com",
					"full_name": "Nguyen Trung Hieu",
					"phone":     "0932432423",
					"avatar":    "a",
				},
				err: nil,
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ParseMapUpdateUserRequest(tt.in)
			if err != nil {
				if err != tt.expected.err {
					t.Errorf("ParseMapUpdateUserRequest() error = %v, wantErr %v", err, tt.expected.err)
					return
				}
			} else if !reflect.DeepEqual(got, tt.expected.out) {
				t.Errorf("ParseMapUpdateUserRequest() = %v, want %v", got, tt.expected.out)
			}
		})
	}
}

func TestConvertUserModelToUserProto(t *testing.T) {
	type expectation struct {
		out *user.ResponseUser
		err error
	}

	tests := map[string]struct {
		in       model.User
		expected expectation
	}{
		"Must_Pass": {
			in: model.User{
				BaseModel: model.BaseModel{
					ID:        1111,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					CreatedBy: 1,
					UpdatedBy: 1,
				},
				Email:         "hieu@gmail.com",
				FullName:      "Nguyen Trung Hieu",
				Phone:         "0929473298",
				Avatar:        "test",
				IsSupperAdmin: true,
				Roles:         model.RoleAdmin,
				Password:      "test",
			},
			expected: expectation{
				out: &user.ResponseUser{
					Id:            1111,
					CreatedAt:     time.Now().Format(time.RFC3339),
					UpdatedAt:     time.Now().Format(time.RFC3339),
					CreatedBy:     1,
					UpdatedBy:     1,
					Email:         "hieu@gmail.com",
					FullName:      "Nguyen Trung Hieu",
					Phone:         "0929473298",
					Avatar:        "test",
					IsSupperAdmin: true,
					Roles:         user.ResponseUser_admin,
				},
				err: nil,
			},
		},
		"Missing_FullName_Must_Pass": {
			in: model.User{
				Email:         "hieu@gmail.com",
				Phone:         "0929473298",
				Avatar:        "test",
				IsSupperAdmin: true,
				Roles:         model.RoleAdmin,
				Password:      "test",
			},
			expected: expectation{
				out: &user.ResponseUser{
					Email:         "hieu@gmail.com",
					Phone:         "0929473298",
					Avatar:        "test",
					IsSupperAdmin: true,
					Roles:         user.ResponseUser_admin,
				},
				err: nil,
			},
		},
		"MissiSupperAdmin_Must_Pass": {
			in: model.User{
				Email:    "hieu@gmail.com",
				Phone:    "0929473298",
				Avatar:   "test",
				Roles:    model.RoleAdmin,
				Password: "test",
			},
			expected: expectation{
				out: &user.ResponseUser{
					Email:  "hieu@gmail.com",
					Phone:  "0929473298",
					Avatar: "test",
					Roles:  user.ResponseUser_admin,
				},
				err: nil,
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ConvertUserModelToUserProto(tt.in)
			if err != nil {
				if err != tt.expected.err {
					t.Errorf("ConvertUserModelToUserProto() error = %v, wantErr %v", err, tt.expected.err)
					return
				}
			} else if !reflect.DeepEqual(got, tt.expected.out) {
				t.Errorf("ConvertUserModelToUserProto() = %v, want %v", got, tt.expected.out)
			}
		})
	}
}
