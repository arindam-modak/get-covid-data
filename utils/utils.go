package utils

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gomodule/redigo/redis"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func HttpGetRequest(url string) []byte {
	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	return responseData
}

func CreateUpdateFile(filename string, data []byte) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()
	_, err2 := f.Write(data)

	if err2 != nil {
		log.Fatal(err2)
	}
}

func ReadFile(filename string) *os.File {
	covidDataFile, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}

	return covidDataFile
}

func GetMongoClient() *mongo.Client {
	/*
	   Connect to my cluster
	*/
	dbUri := fmt.Sprintf("mongodb+srv://%s:%s@%s?retryWrites=true&w=majority", os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"), os.Getenv("MONGO_HOST"))
	client, err := mongo.NewClient(options.Client().ApplyURI(dbUri))
	if err != nil {
		log.Fatal(err)
	}

	return client
}

func GetRedisConn() redis.Conn {
	conn, err := redis.Dial("tcp", os.Getenv("REDIS_URI"), redis.DialPassword(os.Getenv("REDIS_PASSWORD")))
	if err != nil {
		log.Fatal(err)
	}

	return conn
}
