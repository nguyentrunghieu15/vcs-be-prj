package user

import (
	"reflect"
	"testing"

	"github.com/nguyentrunghieu15/vcs-common-prj/apu/user"
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
				Email: "hieu@gmail.com",
			},
			expected: expectation{
				out: map[string]interface{}{
					"email": "hieu@gmail.com",
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
