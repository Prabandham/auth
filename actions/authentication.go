package actions

import (
	db "github.com/auth/db"
	"github.com/auth/entity"
)

func CheckUserAuth(data map[string]string) (string, entity.TokenDetails, bool) {
	var user entity.User
	pg := db.GetDatabaseConnection()
	redis := db.GetRedisConnection()
	if err := pg.Connection.Where("email = ?", data["email"]).First(&user).Error; err != nil {
		return "Invalid User", entity.TokenDetails{}, false
	}

	if !user.ValidatePassword(data["password"]) {
		return "Invalid Password", entity.TokenDetails{}, false
	}

	token, err := entity.CreateAuth(user.ID.String(), redis.Connection)
	if err != nil {
		return "Error performing Auth", entity.TokenDetails{}, true
	}
	return "Auth Success", *token, true
}
