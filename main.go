package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type Weather struct {
	Location struct {
		Name string `json:"name"`
		Region string `json:"region"`
		Country string `json:"country"`
		Localtime string `json:"localtime"`
	} `json:"location"`	
	Current struct {
		Temp_c float64 `json:"temp_c"`
		Is_day int `json:"is_day"`
	} `json:"current"`

}

func main() {

	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	res, err := http.Get(os.Getenv("WEATHER_API_KEY"))
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		panic("API is not available")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	
	var weather Weather
	err = json.Unmarshal(body, &weather)
	if err != nil {
		panic(err)
	}

	location, current, localtime := weather.Location.Name, weather.Current.Temp_c, weather.Location.Localtime

	fmt.Printf( 
		"%s: %.2f°C, %s",
		location, 
		current, 
		localtime)
}