package auth

import (
	"github.com/nguyentrunghieu15/vcs-common-prj/db/model"
)

type Authorizer struct{}

func (*Authorizer) HavePermisionToCreateUser(role model.UserRole) bool {
	return role == model.RoleAdmin
}

func (*Authorizer) HavePermisionToUpdateUser(role model.UserRole) bool {
	return role == model.RoleAdmin
}

func (*Authorizer) HavePermisionToViewUser(role model.UserRole) bool {
	return true
}

func (*Authorizer) HavePermisionToDeleteUser(role model.UserRole) bool {
	return role == model.RoleAdmin
}

func (*Authorizer) HavePermisionToCreateServer(role model.UserRole) bool {
	return role == model.RoleAdmin
}

func (*Authorizer) HavePermisionToUpdateServer(role model.UserRole) bool {
	return role == model.RoleAdmin
}

func (*Authorizer) HavePermisionToViewServer(role model.UserRole) bool {
	return true
}

func (*Authorizer) HavePermisionToDeleteServer(role model.UserRole) bool {
	return role == model.RoleAdmin
}

func (*Authorizer) HavePermisionToImportServer(role model.UserRole) bool {
	return role == model.RoleAdmin
}

func (*Authorizer) HavePermisionToExportServer(role model.UserRole) bool {
	return role == model.RoleAdmin
}

func (*Authorizer) HavePermisionToSendMail(role model.UserRole) bool {
	return role == model.RoleAdmin
}
