package user

import (
	"reflect"
	"testing"

	"github.com/nguyentrunghieu15/vcs-common-prj/apu/server"
	"github.com/nguyentrunghieu15/vcs-common-prj/apu/user"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func TestTypeSortToString(t *testing.T) {
	type expectation struct {
		out string
	}

	type in model.TypeSort
	tests := map[string]struct {
		in
		expected expectation
	}{
		"ASC": {
			in:       in(model.ASC),
			expected: expectation{out: ""},
		},
		"DESC": {
			in: in(model.DESC),
			expected: expectation{
				out: "desc",
			},
		},
		"None": {
			in: in(model.NONE),
			expected: expectation{
				out: "",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := TypeSortToString(model.TypeSort(tt.in)); got != tt.expected.out {
				t.Errorf("TypeSortToString() = %v, want %v", got, tt.expected.out)
			}
		})
	}
}

func TestAddPagination(t *testing.T) {
	type expectation struct {
		out *gorm.DB
	}

	var positiveNum int64 = 1
	var tempString string = "test"
	db := &gorm.DB{
		Config: &gorm.Config{},
		Statement: &gorm.Statement{
			Clauses: make(map[string]clause.Clause),
		},
	}
	type in struct {
		statement *gorm.DB
		req       *user.ListUsersRequest
	}
	tests := map[string]struct {
		in
		expected expectation
	}{
		"Must_Pass": {
			in: in{statement: db, req: &user.ListUsersRequest{
				Pagination: &server.Pagination{
					Limit:    &positiveNum,
					Page:     &positiveNum,
					PageSize: &positiveNum,
					Sort:     server.TypeSort_ASC.Enum(),
					SortBy:   &tempString,
				},
			}},
			expected: expectation{out: db},
		},
	}
	for scenario, tt := range tests {
		t.Run(scenario, func(t *testing.T) {
			if got := AddPagination(tt.in.statement, tt.in.req); !reflect.DeepEqual(got, tt.expected.out) {
				t.Errorf("AddPagination() = %v, want %v", got, tt.expected.out)
			}
		})
	}
}
