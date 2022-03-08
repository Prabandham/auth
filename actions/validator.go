package actions

import "github.com/thoas/go-funk"

// CheckValidaRequestType will check if the request type is supported by this service.
func CheckValidRequestType(requestType string) bool {
	return funk.Contains(
		[]string{"authenticateUser", "authorizeUser", "registerUser", "deleteUserAccessToken", "refreshUserAccessToken"},
		requestType,
	)
}
