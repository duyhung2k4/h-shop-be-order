package config

import (
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func connectGPRCServerWarehouse() {
	var errConn error

	creds, errKey := credentials.NewClientTLSFromFile("keys/server-warehouse/public.pem", "localhost")
	if errKey != nil {
		log.Fatalln(errKey)
	}

	clientWarehouse, errConn = grpc.Dial(host+":20005", grpc.WithTransportCredentials(creds))
	if errConn != nil {
		log.Fatalln(errConn)
	}
}
