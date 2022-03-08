package db

import (
	"fmt"
	"sync"

	"github.com/auth/config"
	"github.com/auth/entity"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"

	"github.com/go-redis/redis/v7"
	_ "github.com/joho/godotenv/autoload"
)

type Db struct {
	Connection *gorm.DB
}

type Redis struct {
	Connection *redis.Client
}

type GormLogger struct{}

// Print - Log Formatter
func (*GormLogger) Print(v ...interface{}) {
	switch v[0] {
	case "sql":
		log.WithFields(
			log.Fields{
				"module":        "gorm",
				"type":          "sql",
				"rows_returned": v[5],
				"src":           v[1],
				"values":        v[4],
				"duration":      v[2],
			},
		).Info(v[3])
	case "log":
		log.WithFields(log.Fields{"module": "gorm", "type": "log"}).Print(v[2])
	}
}

var singleton *Db
var redisClient *Redis
var dbOnce sync.Once
var redisOnce sync.Once

func GetDatabaseConnection() *Db {
	dbOnce.Do(func() {
		psqlInfo := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s",
			config.GetEnv("DB_HOST", "127.0.0.1"),
			config.GetEnv("DB_USER", "srinidhip"),
			config.GetEnv("DB_NAME", "auth"),
			config.GetEnv("DB_PASSWORD", ""),
		)
		db, err := gorm.Open("postgres", psqlInfo)
		if err != nil {
			panic(err.Error())
		}
		singleton = &Db{Connection: db}
	})
	return singleton
}

func GetRedisConnection() *Redis {
	redisOnce.Do(func() {
		dsn := config.GetEnv("REDIS_DSN", "localhost:6379")
		client := redis.NewClient(&redis.Options{
			Addr: dsn, //redis port
		})
		_, err := client.Ping().Result()
		if err != nil {
			panic("Could not connect to redis")
		}
		redisClient = &Redis{Connection: client}
	})
	return redisClient
}

func (db *Db) SetLogger() {
	db.Connection.SetLogger(&GormLogger{})
	db.Connection.LogMode(true)
	formatter := new(log.JSONFormatter)
	log.SetFormatter(formatter)
	formatter.PrettyPrint = true
}

func (db *Db) MigrateModels() {
	db.Connection.AutoMigrate(
		&entity.User{},
		&entity.Setting{},
	)
}
