package user

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nguyentrunghieu15/vcs-common-prj/apu/user"
	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
)

func ParseMapCreateUserRequest(req *user.CreateUserRequest) (map[string]interface{}, error) {
	t, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	mapRequest := make(map[string]interface{})

	if err = json.Unmarshal(t, &mapRequest); err != nil {
		return nil, err
	}

	if v, ok := mapRequest["roles"]; ok {
		if v.(float64) == 2 {
			mapRequest["roles"] = model.RoleUser
		}
		if v.(float64) == 1 {
			mapRequest["roles"] = model.RoleAdmin
		}
	}

	return mapRequest, nil
}

func ParseMapUpdateUserRequest(req *user.UpdateUserByIdRequest) (map[string]interface{}, error) {
	t, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	mapRequest := make(map[string]interface{})

	if err = json.Unmarshal(t, &mapRequest); err != nil {
		return nil, err
	}

	delete(mapRequest, "id")

	if v, ok := mapRequest["roles"]; ok {
		if v.(float64) == 2 {
			mapRequest["roles"] = model.RoleUser
		}
		if v.(float64) == 1 {
			mapRequest["roles"] = model.RoleAdmin
		}
	}

	return mapRequest, nil
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

func ConvertUserModelToUserProto(u model.User) (*user.ResponseUser, error) {
	t, err := json.Marshal(u)
	if err != nil {
		return nil, err
	}

	mapTemp := make(map[string]interface{})

	if err = json.Unmarshal(t, &mapTemp); err != nil {
		return nil, err
	}

	delete(mapTemp, "password")
	switch u.Roles {
	case model.RoleAdmin:
		mapTemp["roles"] = user.ResponseUser_admin
	case model.RoleUser:
		mapTemp["roles"] = user.ResponseUser_user
	default:
		mapTemp["roles"] = user.ResponseUser_none
	}

	t, err = json.Marshal(mapTemp)
	if err != nil {
		return nil, err
	}

	var result user.ResponseUser
	err = json.Unmarshal(t, &result)
	if err != nil {
		return nil, err
	}

	if _, ok := mapTemp["isSupperAdmin"]; ok {
		result.IsSupperAdmin = u.IsSupperAdmin
	}
	if _, ok := mapTemp["fullName"]; ok {
		result.FullName = u.FullName
	}
	if _, ok := mapTemp["createdAt"]; ok && mapTemp["createdAt"] != "0001-01-01T00:00:00Z" {
		result.CreatedAt = u.CreatedAt.Format(time.RFC3339)
	}
	if _, ok := mapTemp["updatedAt"]; ok && mapTemp["updatedAt"] != "0001-01-01T00:00:00Z" {
		result.UpdatedAt = u.UpdatedAt.Format(time.RFC3339)
	}
	if _, ok := mapTemp["deletedAt"]; ok && mapTemp["deletedAt"] != nil {
		fmt.Println(mapTemp["deletedAt"])
		result.DeletedAt = u.DeletedAt.Time.Format(time.RFC3339)
	}

	if _, ok := mapTemp["createdBy"]; ok {
		result.CreatedBy = int64(u.CreatedBy)
	}
	if _, ok := mapTemp["updatedBy"]; ok {
		result.UpdatedBy = int64(u.UpdatedBy)
	}
	if _, ok := mapTemp["deletedBy"]; ok {
		result.DeletedBy = int64(u.DeletedBy)
	}

	return &result, nil
}

func ConvertListUserModelToListUserProto(u []model.User) ([]*user.ResponseUser, error) {
	var result []*user.ResponseUser = make([]*user.ResponseUser, 0)
	for _, v := range u {
		t, err := ConvertUserModelToUserProto(v)
		if err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	return result, nil
}
