package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
)

type Weather struct {
	Location struct {
		Name      string `json:"name"`
		Region    string `json:"region"`
		Country   string `json:"country"`
		Localtime string `json:"localtime"`
	} `json:"location"`
	Current struct {
		TempC float64 `json:"temp_c"`
		IsDay int     `json:"is_day"`
	} `json:"current"`
}

type Json struct {
	Value string `json:"value"`
}

type ErrorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func trimAllSpace(s string) string {
	return strings.Join(strings.Fields(s), "")
}

func createJsonFile(key string) {
	file, err := os.Create("config.json")
	if err != nil {
		exitWithError(err.Error())
	}

	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(Json{key})
	if err != nil {
		exitWithError(err.Error())
	}
}

var ctx, cancel = context.WithCancel(context.Background())
var wg = sync.WaitGroup{}

func main() {
	cityFlag := flag.String("city", "", "City to get weather for, e.g. London")
	flag.Parse()

	file, err := os.Open("config.json")
	if file == nil {
		fmt.Print("Please enter your weatherapi.com key: \n")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			exitWithError(err.Error())
			return
		}
		fmt.Println(input)
		createJsonFile(input)
	}

	if *cityFlag == "" {
		fmt.Print("What city are you looking for? \n")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			exitWithError(err.Error())
			return
		}
		*cityFlag = trimAllSpace(input)
	}

	if err != nil {
		exitWithError(err.Error())
	}

	defer file.Close()

	var key Json
	err = json.NewDecoder(file).Decode(&key)
	if err != nil {
		exitWithError(err.Error())
	}

	query := fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&q=%s&aqi=no", key.Value, *cityFlag)
	res, err := http.Get(query)
	if err != nil {
		exitWithError(err.Error())
	}

	wg.Add(1)
	go func() {
		<-ctx.Done()
		if err := res.Body.Close(); err != nil {
			fmt.Println(err)
			wg.Done()
			return
		}
		wg.Done()
	}()

	// Handle API error

	if res.StatusCode != 200 {
		var errorResponse ErrorResponse
		err = json.NewDecoder(res.Body).Decode(&errorResponse)
		if err != nil {
			exitWithError(err.Error())
		}

		fmt.Println(errorResponse.Error.Message)
		return
	}

	var weather Weather
	err = json.NewDecoder(res.Body).Decode(&weather)
	if err != nil {
		exitWithError(err.Error())
	}

	location, current, localtime := weather.Location.Name, weather.Current.TempC, weather.Location.Localtime

	fmt.Printf(
		"Temperature in %s: %.2f°C, %s\n",
		location,
		current,
		localtime)

	cancel()
	wg.Wait()
}

func exitWithError(message string) {
	fmt.Println(message)
	cancel()
	wg.Wait()
	os.Exit(1)
}
