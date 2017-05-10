package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"
)

type smhiProvider struct{}

func NewSmhiProvider() WeatherProvider {
	return &smhiProvider{}
}

func (*smhiProvider) GetCurrentTemperature(lat, lon float64) (float64, string, error) {

	fc, err := getForecast(lat, lon)

	if err == nil {
		dp := findDataPointClosestToNow(fc)

		temp, err := findParameterByName("t", dp.Parameters)

		if err == nil && len(temp.Values) > 0 {
			return temp.Values[0], temp.Unit, nil
		}
	}

	return 0, "", err
}

func durationAbs(x time.Duration) time.Duration {
	if int64(x) < 0 {
		return time.Duration(-int64(x))
	}
	return x
}

type forecast struct {
	ApprovedTime time.Time
	TimeSeries   []dataPoint
}

type dataPoint struct {
	ValidTime  time.Time
	Parameters []parameter
}

type parameter struct {
	Name      string
	LevelType string
	Level     int
	Unit      string
	Values    []float64
}

const base = "http://opendata-download-metfcst.smhi.se/api/category/pmp2g/version/2/"

func getForecast(lat, lon float64) (*forecast, error) {

	url := fmt.Sprintf(base+"geotype/point/lon/%f/lat/%f/data.json", lon, lat)

	httpClient := &http.Client{Timeout: 5 * time.Second}

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/json")
	resp, err := httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var fc forecast
	err = json.NewDecoder(resp.Body).Decode(&fc)

	if err != nil {
		return nil, err
	}

	// make sure there are at least one datapoint
	if len(fc.TimeSeries) > 0 {
		return &fc, nil
	} else {
		return nil, err
	}

}

// assumes there is at least one dataPoint
func findDataPointClosestToNow(fc *forecast) dataPoint {
	const durationMax = time.Duration(math.MaxInt64)

	// closest to time.Now()
	var closestDataPoint dataPoint
	closestDistance := durationMax

	for _, dp := range fc.TimeSeries {
		distance := durationAbs(dp.ValidTime.Sub(time.Now()))
		if distance < closestDistance {
			closestDataPoint = dp
			closestDistance = distance
		}
	}

	return closestDataPoint
}

func findParameterByName(name string, parameters []parameter) (*parameter, error) {
	for _, p := range parameters {
		if p.Name == name {
			return &p, nil
		}
	}
	return nil, errors.New("No parameter by that name")
}
