package server

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	pb "github.com/nguyentrunghieu15/vcs-common-prj/apu/server"
)

type IValidator interface {
	Validate(interface{}) error
}

type ServerServiceValidator struct {
	validator *validator.Validate
}

func (u *ServerServiceValidator) Validate(value interface{}) error {
	err := u.validator.Struct(value)
	if err != nil {
		msgErr := err.Error()
		splitedMsg := strings.Split(msgErr, "Error:")
		return fmt.Errorf(splitedMsg[1])
	}
	return err
}

func NewServerServiceValidator() *ServerServiceValidator {
	var val = validator.New()

	val.RegisterValidation("daterfc3339", DateRFC3339)

	for k, v := range ValidateRules {
		fmt.Printf("Set rule for %v \n", k)
		val.RegisterStructValidationMapRules(v.rule, v.types)
	}

	return &ServerServiceValidator{
		validator: val,
	}
}

type ValidateRule struct {
	types interface{}
	rule  map[string]string
}

func DateRFC3339(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	_, err := time.Parse(time.RFC3339, value)
	return err == nil
}

var ValidateRules map[string]ValidateRule = map[string]ValidateRule{
	"CreateServerRequest": {
		types: pb.CreateServerRequest{},
		rule: map[string]string{
			"Name":    "required",
			"Ipv4":    "required,ipv4",
			"Status ": "omitempty,required,min=0,max=2",
		},
	},
	"DeleteServerByIdRequest": {
		types: pb.DeleteServerByIdRequest{},
		rule: map[string]string{
			"Id": "required,uuid4",
		},
	},
	"DeleteServerByNameRequest": {
		types: pb.DeleteServerByNameRequest{},
		rule: map[string]string{
			"Name": "required",
		},
	},
	"FileExport": {
		types: pb.FileExport{},
		rule: map[string]string{
			"FileName": "required,endswith=.xlsx",
		},
	},
	"FilterServer": {
		types: pb.FilterServer{},
		rule: map[string]string{
			"CreatedAtFrom": "omitempty,required,daterfc3339",
			"CreatedAtTo":   "omitempty,required,daterfc3339,gtefield=CreatedAtFrom",
			"UpdatedAtFrom": "omitempty,required,daterfc3339",
			"UpdatedAtTo":   "omitempty,required,daterfc3339,gtefield=UpdatedAtFrom",
			"Status":        "omitempty,required,min=0,max=2",
		},
	},
	"PaginationExportRequest": {
		types: pb.PaginationExportRequest{},
		rule: map[string]string{
			"PageSize": "omitempty,required,min=1",
			"FromPage": "omitempty,required,min=1",
			"ToPage":   "omitempty,required,min=1",
			"Sort":     "omitempty,required,min=0,max=2",
		},
	},
	"ExportServerRequest": {
		types: pb.ExportServerRequest{},
		rule: map[string]string{
			"UserId":     "required",
			"File":       "required",
			"Filter":     "omitempty,required",
			"Pagination": "omitempty,required",
		},
	},
	"GetServerByIdRequest": {
		types: pb.GetServerByIdRequest{},
		rule: map[string]string{
			"Id": "required,uuid4",
		},
	},
	"GetServerByNameRequest": {
		types: pb.GetServerByNameRequest{},
		rule: map[string]string{
			"Name": "required",
		},
	},

	"Pagination": {
		types: pb.Pagination{},
		rule: map[string]string{
			"Limit":    "omitempty,required,min=1",
			"Page":     "omitempty,required,min=1",
			"PageSize": "omitempty,required,min=1",
			"Sort":     "omitempty,required,min=0,max=2",
		},
	},
	"ListServerRequest": {
		types: pb.ListServerRequest{},
		rule: map[string]string{
			"Query":      "omitempty",
			"Filter":     "omitempty,required",
			"Pagination": "omitempty,required",
		},
	},
	"UpdateServerRequest": {
		types: pb.UpdateServerRequest{},
		rule: map[string]string{
			"Id":      "required,uuid4",
			"Name":    "omitempty,required",
			"Ipv4":    "omitempty,required,ipv4",
			"Status ": "omitempty,required,min=0,max=2",
		},
	},
}
