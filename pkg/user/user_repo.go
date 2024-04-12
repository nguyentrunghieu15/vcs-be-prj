package user

import (
	"encoding/json"
	"strings"

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

func (u *UserRepositoryDecorator) FindUsers(filter model.FilterQueryInterface) ([]model.User, error) {
	var user []model.User
	var orderQuery string
	offSet := filter.GetPage() * filter.GetPageSize()
	if strings.Trim(filter.GetSortBy(), " ") == "" {
		orderQuery = strings.Trim(filter.GetSortBy(), " ") + " " + TypeSortToString(filter.GetSort())
	}

	var result = u.db

	if filter.GetLimit() != -1 {
		result = result.Limit(int(filter.GetLimit()))
	}

	if filter.GetPage() != -1 && filter.GetPageSize() != -1 {
		result = result.Offset(int(offSet))
	}

	if filter.GetSortBy() != "" {
		result = result.Order(orderQuery)
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

	if _, ok := result["Roles"]; !ok {
		result["Roles"] = model.RoleAdmin
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

	if _, ok := result["Roles"]; !ok {
		result["Roles"] = model.RoleAdmin
	} else {
		result["Roles"] = model.RoleUser
	}

	return result, nil
}
