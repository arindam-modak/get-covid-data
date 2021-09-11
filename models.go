package main

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
	DataUpdatedAt   string `bson:"data_updated_at,omitempty"`
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
