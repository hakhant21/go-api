package rbac

const (
	PermUsersRead   = "users:read"
	PermUsersWrite  = "users:write"
	PermUsersDelete = "users:delete"

	PermRolesManage = "roles:manage"
)

var DefaultRoles = map[string][]string{
	"user": {
		PermUsersRead,
	},
	"moderator": {
		PermUsersRead,
		PermUsersWrite,
	},
	"admin": {
		PermUsersRead, PermUsersWrite, PermUsersDelete,
		PermRolesManage,
	},
}
