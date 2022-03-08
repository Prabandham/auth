package entity

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/auth/config"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis/v7"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	crypt "golang.org/x/crypto/bcrypt"
)

// User model
type User struct {
	Base
	Email             string `gorm:"index;unique;size:256" json:"Email" sql:"not null"`
	FirstName         string `gorm:"index;size:256" json:"FirstName" sql:"not null"`
	LastName          string `gorm:"index;size:256" json:"LastName" sql:"not null"`
	Password          string `gorm:"-" json:"Password"`
	EncryptedPassword string `json:"-"`
	Addresses         []Address
	Contacts          []Contact
}

// TokenDetails specifies all the necessary items to generate an jwt token.
type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUuid   string
	RefreshUuid  string
	AtExpires    int64
	RtExpires    int64
}

// AccessDetails contains the AccessUUID and the current users ID.
type AccessDetails struct {
	AccessUuid string
	UserId     string
}

// Before we save a user we check if the password is present
// if present we will hash it and save it.
func (u *User) BeforeSave(tx *gorm.DB) (err error) {
	if u.Password != "" {
		encryptedPassword, err := hashPassword(u.Password)

		if err != nil {
			return err
		}

		u.EncryptedPassword = encryptedPassword
	}
	return
}

// ValidatePassword will check if passed in password matches the encrypted password
func (u *User) ValidatePassword(password string) bool {
	return doPasswordsMatch(u.EncryptedPassword, password)
}

// Hash password using the Crypt hashing algorithm
// and then return the hashed password as a
// base64 encoded string
func hashPassword(password string) (string, error) {
	var passwordBytes = []byte(password)
	hashedPasswordBytes, err := crypt.GenerateFromPassword(passwordBytes, crypt.MinCost)
	return string(hashedPasswordBytes), err
}

// Check if two passwords match using crypt's CompareHashAndPassword
// which return nil on success and an error on failure.
func doPasswordsMatch(hashedPassword, currentPassword string) bool {
	err := crypt.CompareHashAndPassword(
		[]byte(hashedPassword), []byte(currentPassword))
	return err == nil
}

// Generate a token for a given userId
func CreateToken(userid string) (*TokenDetails, error) {
	var err error
	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(time.Minute * 60).Unix() // Expire in 60 minutes
	td.AccessUuid = uuid.NewV4().String()
	td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix() // Expire in 7 days
	td.RefreshUuid = uuid.NewV4().String()

	// Creating Access Token
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.AccessUuid
	atClaims["user_id"] = userid
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(config.GetEnv("ACCESS_SECRET", "")))
	if err != nil {
		return nil, err
	}

	// Creating Refresh Token
	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUuid
	rtClaims["user_id"] = userid
	rtClaims["exp"] = td.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(config.GetEnv("REFRESH_SECRET", "")))
	if err != nil {
		return nil, err
	}
	return td, nil
}

// CreateAuth will persist the data to redis, this will help us invalidate the token as soon as the user logs out.
func CreateAuth(userid string, redisClient *redis.Client) (*TokenDetails, error) {
	td, err := CreateToken(userid)
	if err != nil {
		panic(err)
	}
	at := time.Unix(td.AtExpires, 0) //converting Unix to UTC(to Time object)
	rt := time.Unix(td.RtExpires, 0)
	now := time.Now()

	errAccess := redisClient.Set(td.AccessUuid, userid, at.Sub(now)).Err()
	if errAccess != nil {
		return nil, errAccess
	}
	errRefresh := redisClient.Set(td.RefreshUuid, userid, rt.Sub(now)).Err()
	if errRefresh != nil {
		return nil, errRefresh
	}
	return td, nil
}

func ExtractToken(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	//normally Authorization the_token_xxx
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

func VerifyToken(r *http.Request) (*jwt.Token, error) {
	tokenString := ExtractToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.GetEnv("ACCESS_SECRET", "")), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func TokenValid(r *http.Request) error {
	token, err := VerifyToken(r)
	if err != nil {
		return err
	}
	if !token.Valid {
		return err
	}
	return nil
}

// ExtractTokenMetadata will extract the token passed in from the http request
func ExtractTokenMetadata(r *http.Request) (*AccessDetails, error) {
	token, err := VerifyToken(r)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		accessUuid, ok := claims["access_uuid"].(string)
		if !ok {
			return nil, err
		}
		userId, err := claims["user_id"].(string)
		if !err {
			return nil, errors.New("could not fetch user Id from claims")
		}
		return &AccessDetails{
			AccessUuid: accessUuid,
			UserId:     userId,
		}, nil
	}
	return nil, err
}

// FetchAuth will try to get the details rom redis.
func FetchAuth(authD *AccessDetails, redisClient *redis.Client) (string, error) {
	userId, err := redisClient.Get(authD.AccessUuid).Result()
	if err != nil {
		return "", err
	}
	return userId, nil
}

// DeleteAuth will try to find the saved token from redis and delete it.
func DeleteAuth(givenUuid string, redisClient *redis.Client) (int64, error) {
	deleted, err := redisClient.Del(givenUuid).Result()
	if err != nil {
		return 0, err
	}
	return deleted, nil
}
