package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/labstack/echo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// State,Confirmed,Recovered,Deaths,Active,Last_Updated_Time,Migrated_Other,State_code,Delta_Confirmed,Delta_Recovered,Delta_Deaths,State_Notes

type CovidData struct { // Our example struct, you can use "-" to ignore a field
	State           string `csv:"State"`
	Confirmed       int    `csv:"Confirmed"`
	Recovered       int    `csv:"Recovered"`
	LastUpdatedTime string `csv:"Last_Updated_Time"`
}

/*
   Define my document struct
*/
type CovidDataDB struct {
	State           string `bson:"state,omitempty"`
	ConfirmedCases  int    `bson:"confirmed_cases,omitempty"`
	RecoveredCases  int    `bson:"recovered_cases,omitempty"`
	LastUpdatedTime string `bson:"last_updated_time,omitempty"`
	DataUpdtedAt    string `bson:"data_updated_at,omitempty"`
}

/*
	Define GeoLocation struct
*/
type GeoLocationOutput struct {
	Items []GeoLocationItem `json:"items"`
}

type GeoLocationItem struct {
	Address GeoLocationItemAddress `json:"address"`
}

type GeoLocationItemAddress struct {
	State string `json:"state"`
}

func main() {
	fmt.Println("Hello Covid! Please go away now.")
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello from web server")
	})

	e.GET("/fetch-data-and-save", func(c echo.Context) error {
		response, err := http.Get(os.Getenv("COVID_DATA_URL"))
		if err != nil {
			fmt.Print(err.Error())
			os.Exit(1)
		}

		responseData, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(responseData))

		f, err := os.Create("data.csv")

		if err != nil {
			log.Fatal(err)
		}

		defer f.Close()

		_, err2 := f.Write(responseData)

		if err2 != nil {
			log.Fatal(err2)
		}

		covidDataFile, err := os.OpenFile("data.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			panic(err)
		}
		defer covidDataFile.Close()

		covidData := []*CovidData{}

		if err := gocsv.UnmarshalFile(covidDataFile, &covidData); err != nil { // Load clients from file
			panic(err)
		}

		var docs []interface{}

		for _, client := range covidData {
			fmt.Println("Number of Cases ", client.State, client.Confirmed, client.LastUpdatedTime)
			docs = append(docs, bson.D{{"state", client.State}, {"confirmed_cases", client.Confirmed}, {"recovered_cases", client.Recovered}, {"last_updated_time", client.LastUpdatedTime}, {"data_updated_at", time.Now().String()}})
		}

		/*
		   Connect to my cluster
		*/
		dbUri := fmt.Sprintf("mongodb+srv://%s:%s@%s?retryWrites=true&w=majority", os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"), os.Getenv("MONGO_HOST"))
		client, err := mongo.NewClient(options.Client().ApplyURI(dbUri))
		if err != nil {
			log.Fatal(err)
		}
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		err = client.Connect(ctx)
		if err != nil {
			log.Fatal(err)
		}
		defer client.Disconnect(ctx)

		/*
			List databases
		*/
		databases, err := client.ListDatabaseNames(ctx, bson.M{})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(databases)

		/*
		   Get my collection instance
		*/

		collection := client.Database(os.Getenv("MONGO_DB_NAME")).Collection(os.Getenv("MONGO_COLLECTION_NAME"))
		collection.Drop(ctx)

		/*
			Insert documents
		*/

		res, insertErr := collection.InsertMany(ctx, docs)
		if insertErr != nil {
			log.Fatal(insertErr)
		}
		fmt.Println(res)
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
		fmt.Println(covidDataDB)

		return c.String(http.StatusOK, "Latest Data Fetched")
	})

	e.GET("/get-data-from-location", func(c echo.Context) error {
		latitude := c.QueryParam("latitude")
		longitude := c.QueryParam("longitude")

		locationUri := fmt.Sprintf("%s?at=%s,%s&apikey=%s", os.Getenv("REVERSE_GEOCODE_URL"), latitude, longitude, os.Getenv("REVERSE_GEOCODE_APIKEY"))
		response, err := http.Get(locationUri)
		if err != nil {
			fmt.Print(err.Error())
			os.Exit(1)
		}

		responseData, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(responseData))

		// decoder := json.NewDecoder(response.Body)
		// var data GeoLocationOutput
		// err = decoder.Decode(&data)

		var data = new(GeoLocationOutput)
		err = json.Unmarshal(responseData, &data)

		fmt.Println(data)

		/*
		   Connect to my cluster
		*/
		dbUri := fmt.Sprintf("mongodb+srv://%s:%s@%s?retryWrites=true&w=majority", os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"), os.Getenv("MONGO_HOST"))
		client, err := mongo.NewClient(options.Client().ApplyURI(dbUri))
		if err != nil {
			log.Fatal(err)
		}
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		err = client.Connect(ctx)
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

		var covidDataDB []CovidDataDB
		if err = cur.All(ctx, &covidDataDB); err != nil {
			panic(err)
		}
		fmt.Println(covidDataDB)

		return c.JSON(http.StatusOK, covidDataDB)
	})

	e.Start(":8000")
}
