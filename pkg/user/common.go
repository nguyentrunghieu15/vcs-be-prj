package user

import (
	"encoding/json"
	"fmt"

	"github.com/nguyentrunghieu15/vcs-common-prj/apu/user"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
)

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

	fmt.Println(mapRequest)

	for i := 0; i < len(DefinedFieldCreateUserRequest); i++ {
		if value, ok := mapRequest[DefinedFieldCreateUserRequest[i]["fieldNameProto"]]; ok {
			result[DefinedFieldCreateUserRequest[i]["fieldNameModel"]] = value
		}
	}

	if _, ok := result["IsSupperAdmin"]; !ok {
		result["IsSupperAdmin"] = false
	}

	if _, ok := result["Roles"]; ok {
		if user.CreateUserRequest_Role(result["Roles"].(float64)) == user.CreateUserRequest_admin {
			result["Roles"] = model.RoleAdmin
		}
		if user.CreateUserRequest_Role(result["Roles"].(float64)) == user.CreateUserRequest_user {

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
		if req.GetRoles() == user.UpdateUserByIdRequest_admin {
			result["Roles"] = model.RoleAdmin
		}
		if req.GetRoles() == user.UpdateUserByIdRequest_user {
			result["Roles"] = model.RoleUser
		}
	}

	return result, nil
}

func ConvertUserRoleModelToUserRoleProto(role model.UserRole) user.ResponseUser_Role {
	switch role {
	case model.RoleAdmin:
		return user.ResponseUser_admin
	case model.RoleUser:
		return user.ResponseUser_user
	default:
		return user.ResponseUser_none
	}
}

func ConvertUserRoleProtoToUserRoleModel(role user.User_Role) model.UserRole {
	switch role {
	case user.User_admin:
		return model.RoleAdmin
	case user.User_user:
		return model.RoleUser
	default:
		return ""
	}
}

func ConvertUserModelToUserProto(u model.User) *user.ResponseUser {
	return &user.ResponseUser{
		Id:            int64(u.ID),
		CreatedAt:     u.CreatedAt.String(),
		CreatedBy:     int64(u.CreatedBy),
		UpdatedAt:     u.UpdatedAt.String(),
		UpdatedBy:     int64(u.UpdatedBy),
		Email:         u.Email,
		FullName:      u.FullName,
		Phone:         u.Phone,
		Avatar:        u.Avatar,
		IsSupperAdmin: u.IsSupperAdmin,
		Roles:         ConvertUserRoleModelToUserRoleProto(u.Roles),
	}
}

func ConvertListUserModelToListUserProto(u []model.User) []*user.ResponseUser {
	var result []*user.ResponseUser = make([]*user.ResponseUser, 0)
	for _, v := range u {
		result = append(result, ConvertUserModelToUserProto(v))
	}
	return result
}
