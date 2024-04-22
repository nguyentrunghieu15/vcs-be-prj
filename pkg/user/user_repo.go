package user

import (
	"encoding/json"
	"fmt"

	"github.com/nguyentrunghieu15/vcs-common-prj/apu/server"
	"github.com/nguyentrunghieu15/vcs-common-prj/apu/user"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
	"gorm.io/gorm"
)

type UserRepositoryDecorator struct {
	*model.UserRepository
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepositoryDecorator {
	return &UserRepositoryDecorator{model.CreateUserRepository(db), db}
}

func TypeSortToString(v model.TypeSort) string {
	switch v {
	case model.ASC, model.NONE:
		return ""
	case model.DESC:
		return "desc"
	}
	return ""
}

func (u *UserRepositoryDecorator) FindUsers(req *user.ListUsersRequest) ([]model.User, error) {
	var user []model.User
	result := u.db
	if req.GetPagination() != nil {
		if limit := req.GetPagination().Limit; limit != nil && *limit > 1 {
			result = result.Limit(int(*limit))
		}
		page := req.GetPagination().Page
		pageSize := req.GetPagination().PageSize
		if page != nil && pageSize != nil && *page > 0 && *pageSize > 0 {
			result.Offset(int((*page - 1) * (*pageSize)))
		}
		if orderBy := req.GetPagination().SortBy; orderBy != nil {
			if req.GetPagination().Sort != nil && req.GetPagination().Sort == server.TypeSort_DESC.Enum() {
				result = result.Order(fmt.Sprintf("%v %v", orderBy, "DESC"))
			} else {
				result = result.Order(orderBy)
			}

		}
	}
	result = result.Find(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return user, nil
}

func ParseMapCreateUserRequest(req *user.CreateUserRequest) (map[string]interface{}, error) {
	t, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	mapRequest := make(map[string]interface{})
	result := make(map[string]interface{})

	if err = json.Unmarshal(t, &mapRequest); err != nil {
		return nil, err
	}

	for i := 0; i < len(DefinedFieldCreateUserRequest); i++ {
		if value, ok := mapRequest[DefinedFieldCreateUserRequest[i]["fieldNameProto"]]; ok {
			result[DefinedFieldCreateUserRequest[i]["fieldNameModel"]] = value
		}
	}

	if _, ok := result["IsSupperAdmin"]; !ok {
		result["IsSupperAdmin"] = false
	}

	if _, ok := result["Roles"]; ok {
		if user.UserRole(result["Roles"].(float64)) == user.UserRole_RoleAdmin {
			result["Roles"] = model.RoleAdmin
		}
		if user.UserRole(result["Roles"].(float64)) == user.UserRole_RoleUser {

			result["Roles"] = model.RoleUser
		}
	}

	return result, nil
}

func ParseMapUpdateUserRequest(req *user.UpdateUserByIdRequest) (map[string]interface{}, error) {
	var fieldName = []string{"Email", "FullName", "Phone", "Avatar", "IsSupperAdmin", "Roles"}
	var fieldProtoName = []string{"email", "full_name", "phone", "avatar", "is_supper_admin", "roles"}

	t, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	mapRequest := make(map[string]interface{})
	result := make(map[string]interface{})

	if err = json.Unmarshal(t, &mapRequest); err != nil {
		return nil, err
	}

	for i := 0; i < len(fieldName); i++ {
		if value, ok := mapRequest[fieldProtoName[i]]; ok {
			result[fieldName[i]] = value
		}
	}

	if _, ok := result["IsSupperAdmin"]; !ok {
		result["IsSupperAdmin"] = false
	}

	if _, ok := result["Roles"]; ok {
		if req.GetRoles() == user.UserRole_RoleAdmin {
			result["Roles"] = model.RoleAdmin
		}
		if req.GetRoles() == user.UserRole_RoleUser {
			result["Roles"] = model.RoleUser
		}
	}

	return result, nil
}
