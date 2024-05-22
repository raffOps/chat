package util

import "github.com/raffops/chat/pb"

func StringToRole(roleStr string) pb.Role {
	switch roleStr {
	case "ADMIN":
		return pb.Role_ADMIN
	case "USER":
		return pb.Role_USER
	default:
		return pb.Role_USER // default value if the string doesn't match any Role
	}
}
