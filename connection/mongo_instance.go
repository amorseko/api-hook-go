package connection

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

var dbMongo *mongo.Database

func init() {
	e := godotenv.Load()
	if e != nil {
		fmt.Print(e)
	}

	protocol := os.Getenv("db_protocol_mongo")
	username := os.Getenv("db_user_mongo")
	password := os.Getenv("db_pass_mongo")
	dbName := os.Getenv("db_name_mongo")
	dbHost := os.Getenv("db_host_mongo")
	dbPort := os.Getenv("db_port_mongo")

	dbUri := fmt.Sprintf("%s://%s:%s@%s:%s", protocol, username, password, dbHost, dbPort)

	clientOptions := options.Client()
	clientOptions.ApplyURI(dbUri + "/?compressors=disabled&gssapiServiceName=mongodb")
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		fmt.Print(err)
	}

	err = client.Connect(context.Background())
	if err != nil {
		fmt.Print(err)
	}
	dbMongo = client.Database(dbName)
}

func GetDBMongo() *mongo.Database {
	return dbMongo
}
