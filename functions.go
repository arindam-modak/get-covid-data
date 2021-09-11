package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo"
	"go.mongodb.org/mongo-driver/bson"
)

func fetchDataAndSave(c echo.Context) error {
	responseData := httpGetRequest(os.Getenv("COVID_DATA_URL"))
	fmt.Println("[fetchDataAndSave] responseData: ", string(responseData))

	createUpdateFile("data.csv", responseData)

	covidDataFile := readFile("data.csv")
	defer covidDataFile.Close()

	covidData := []*CovidData{}

	if err := gocsv.UnmarshalFile(covidDataFile, &covidData); err != nil { // Load data from file
		panic(err)
	}

	var docs []interface{}

	for _, client := range covidData {
		fmt.Println("[fetchDataAndSave] Cases: ", client.State, client.Confirmed, client.LastUpdatedTime)
		docs = append(docs, bson.D{{"state", client.State}, {"confirmed_cases", client.Confirmed}, {"recovered_cases", client.Recovered}, {"last_updated_time", client.LastUpdatedTime}, {"data_updated_at", time.Now().String()}})
	}

	/*
	   Connect to my cluster
	*/
	client := getMongoClient()
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	collection := client.Database(os.Getenv("MONGO_DB_NAME")).Collection(os.Getenv("MONGO_COLLECTION_NAME"))
	collection.Drop(ctx)

	/*
		Insert documents
	*/
	res, insertErr := collection.InsertMany(ctx, docs)
	if insertErr != nil {
		log.Fatal(insertErr)
	}
	fmt.Println("[fetchDataAndSave] dbInsertResponse: ", res)

	/*
		Iterate a cursor and print it
	*/
	cur, currErr := collection.Find(ctx, bson.D{})

	if currErr != nil {
		panic(currErr)
	}
	defer cur.Close(ctx)

	var covidDataDB []CovidDataDB
	if err = cur.All(ctx, &covidDataDB); err != nil {
		panic(err)
	}
	fmt.Println("[fetchDataAndSave] covidDataDB: ", covidDataDB)
	fmt.Println("[fetchDataAndSave] Latest Data Fetched")

	return c.String(http.StatusOK, "Latest Data Fetched")
}

func getDataFromLocation(c echo.Context) error {
	latitude := c.QueryParam("latitude")
	longitude := c.QueryParam("longitude")

	locationUri := fmt.Sprintf("%s?at=%s,%s&apikey=%s", os.Getenv("REVERSE_GEOCODE_URL"), latitude, longitude, os.Getenv("REVERSE_GEOCODE_APIKEY"))
	responseData := httpGetRequest(locationUri)
	fmt.Println("[getDataFromLocation] responseData: ", string(responseData))

	var data = new(GeoLocationOutput)
	_ = json.Unmarshal(responseData, &data)

	fmt.Println("[getDataFromLocation] GeoLocationOutputData: ", data)

	conn := getRedisConn()
	defer conn.Close()
	fmt.Println("[getDataFromLocation] Connected to Redis")

	var covidDataDB []CovidDataDB

	redisResult, err := redis.String(conn.Do("GET", data.Items[0].Address.State))
	fmt.Println("[getDataFromLocation] From Redis: ", redisResult)
	if err != nil {
		fmt.Println("[getDataFromLocation] Cache Missed, So Fetching From DB")
		/*
			Connect to my cluster
		*/
		client := getMongoClient()
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		err := client.Connect(ctx)
		if err != nil {
			log.Fatal(err)
		}
		defer client.Disconnect(ctx)

		/*
			Get my collection instance
		*/
		collection := client.Database(os.Getenv("MONGO_DB_NAME")).Collection(os.Getenv("MONGO_COLLECTION_NAME"))

		cur, currErr := collection.Find(ctx, bson.D{{"state", data.Items[0].Address.State}})

		if cur == nil {
			panic(currErr)
		}
		defer cur.Close(ctx)

		if err = cur.All(ctx, &covidDataDB); err != nil {
			panic(err)
		}

		covidDataDBStringified, err := json.Marshal(covidDataDB)
		// pushing to redis with 30 min expiration time
		redisResult, err := redis.String(conn.Do("SET", data.Items[0].Address.State, covidDataDBStringified, "EX", "1800"))
		fmt.Println("[getDataFromLocation] Data Pushed To Cache With Response: ", redisResult)
	} else {
		json.Unmarshal([]byte(redisResult), &covidDataDB)
	}

	fmt.Println(covidDataDB)

	return c.JSON(http.StatusOK, covidDataDB)
}
