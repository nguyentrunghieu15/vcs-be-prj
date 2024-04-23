package user

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	pb "github.com/nguyentrunghieu15/vcs-common-prj/apu/user"
)

type IValidator interface {
	Validate(interface{}) error
}

type UserServiceValidator struct {
	validator *validator.Validate
}

func (u *UserServiceValidator) Validate(value interface{}) error {
	return u.validator.Struct(value)
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
			"Id": "omitempty,required,min=0",
		},
	},
	"GetUserByEmailRequest": {
		types: pb.GetUserByEmailRequest{},
		rule: map[string]string{
			"Email": "required,email",
		},
	},
}
