package user

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/nguyentrunghieu15/vcs-common-prj/apu/server"
	pb "github.com/nguyentrunghieu15/vcs-common-prj/apu/user"
)

type IValidator interface {
	Validate(interface{}) error
}

type UserServiceValidator struct {
	validator *validator.Validate
}

func (u *UserServiceValidator) Validate(value interface{}) error {
	err := u.validator.Struct(value)
	if err != nil {
		msgErr := err.Error()
		splitedMsg := strings.Split(msgErr, "Error:")
		return fmt.Errorf(splitedMsg[1])
	}
	return err
}

func NewUserServiceValidator() *UserServiceValidator {
	var val = validator.New()

	for k, v := range ValidateRules {
		fmt.Printf("Set rule for %v \n", k)
		val.RegisterStructValidationMapRules(v.rule, v.types)
	}

	return &UserServiceValidator{
		validator: val,
	}
}

type ValidateRule struct {
	types interface{}
	rule  map[string]string
}

var ValidateRules map[string]ValidateRule = map[string]ValidateRule{
	"GetUserByIdRequest": {
		types: pb.GetUserByIdRequest{},
		rule: map[string]string{
			"Id": "required,min=1",
		},
	},
	"GetUserByEmailRequest": {
		types: pb.GetUserByEmailRequest{},
		rule: map[string]string{
			"Email": "required,email",
		},
	},
	"ListUsersRequest": {
		types: pb.ListUsersRequest{},
		rule: map[string]string{
			"Pagination": "omitempty,required",
		},
	},
	"Server.Pagination": {
		types: server.Pagination{},
		rule: map[string]string{
			"Limit":    "omitempty,required,min=1",
			"Page":     "omitempty,required,min=1",
			"PageSize": "omitempty,required,min=1",
			"Sort":     "omitempty,required,min=0,max=2",
		},
	},
	"CreateUserRequest": {
		types: pb.CreateUserRequest{},
		rule: map[string]string{
			"Email":    "required,email",
			"Avatar":   "omitempty,required,base64",
			"Password": "required,min=6",
			"Roles":    "omitempty,required,min=0,max=2",
		},
	},
	"UpdateUserByIdRequest": {
		types: pb.UpdateUserByIdRequest{},
		rule: map[string]string{
			"Id":     "required,min=0",
			"Email":  "omitempty,required,email",
			"Avatar": "omitempty,required,base64",
			"Roles":  "omitempty,required,min=0,max=2",
		},
	},
	"DeleteUserByIdRequest": {
		types: pb.DeleteUserByIdRequest{},
		rule: map[string]string{
			"Id": "required,min=0",
		},
	},
}
