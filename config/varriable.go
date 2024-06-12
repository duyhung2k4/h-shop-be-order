package config

import (
	"github.com/go-chi/jwtauth/v5"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

const (
	APP_PORT      = "APP_PORT"
	DB_HOST       = "DB_HOST"
	DB_PORT       = "DB_PORT"
	DB_NAME       = "DB_NAME"
	DB_PASSWORD   = "DB_PASSWORD"
	DB_USER       = "DB_USER"
	URL_REDIS     = "URL_REDIS"
	HOST          = "HOST"
	URL_RABBIT_MQ = "URL_RABBIT_MQ"
)

var (
	appPort     string
	dbHost      string
	dbPort      string
	dbName      string
	dbPassword  string
	dbUser      string
	urlRedis    string
	host        string
	urlRabbitMq string

	db  *gorm.DB
	rdb *redis.Client
	jwt *jwtauth.JWTAuth

	clientWarehouse *grpc.ClientConn
	clientPayment   *grpc.ClientConn
	rabbitChannel   *amqp091.Channel
)
