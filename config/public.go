package config

import (
	"github.com/go-chi/jwtauth/v5"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

func GetDB() *gorm.DB {
	return db
}

func GetRDB() *redis.Client {
	return rdb
}

func GetAppPort() string {
	return appPort
}

func GetJWT() *jwtauth.JWTAuth {
	return jwt
}

func GetConnWarehouseGRPC() *grpc.ClientConn {
	return clientWarehouse
}

func GetConnPaymentGRPC() *grpc.ClientConn {
	return clientPayment
}

func GetHost() string {
	return host
}

func GetRabbitChannel() *amqp091.Channel {
	return rabbitChannel
}
