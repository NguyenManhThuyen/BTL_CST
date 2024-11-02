package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/umahmood/haversine"
)

type Point struct {
	ID        string  `json:"id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Status    bool    `json:"status"`
}

// DistancePoint represents a point with its distance to another point
type DistancePoint struct {
	ID        string  `json:"id"`
	Distance  float64 `json:"distance"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Result represents the result for each point
type Result struct {
	Point     Point           `json:"point"`
	Distances []DistancePoint `json:"distances"`
}

// HERE API response structure
type HereResponse struct {
	Routes []struct {
		Sections []struct {
			Summary struct {
				Length int `json:"length"`
			} `json:"summary"`
		} `json:"sections"`
	} `json:"routes"`
}

// HERE API keys
var hereAPIKeys = "Ra-29P9uDjHk-iJKspY1hRnowsD8xV2fmdpR55ResUo"

// GetDistanceFromHereAPI fetches the distance between two points using HERE API
func GetDistanceFromHereAPI(originLat, originLon, destLat, destLon float64) (float64, error) {
	url := fmt.Sprintf("https://router.hereapi.com/v8/routes?transportMode=car&origin=%.6f,%.6f&destination=%.6f,%.6f&return=summary&apiKey=%s", originLat, originLon, destLat, destLon, hereAPIKeys)
	fmt.Println(url)
	resp, err := http.Get(url)
	// Tăng biến đếm số lần gọi API
	apiCallCount++
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	var hereResp HereResponse
	err = json.Unmarshal(body, &hereResp)
	if err != nil {
		return 0, err
	}

	fmt.Println(hereResp)
	if len(hereResp.Routes) > 0 && len(hereResp.Routes[0].Sections) > 0 {
		return float64(hereResp.Routes[0].Sections[0].Summary.Length) / 1000, nil // Convert meters to kilometers
	}

	return 0, fmt.Errorf("no route found")
}

// Biến đếm số lần gọi API
var apiCallCount = 0

func main() {
	// Load points from JSON file
	jsonFile, err := os.ReadFile("points.json")
	if err != nil {
		fmt.Println("Error reading points.json:", err)
		return
	}

	var points []Point
	err = json.Unmarshal(jsonFile, &points)
	if err != nil {
		fmt.Println("Error parsing points.json:", err)
		return
	}

	var results []Result

	for i, point := range points {
		// Kiểm tra nếu số lần gọi API đã đạt 1000
		if apiCallCount >= 500 {
			fmt.Println("Reached maximum API call limit of 1000. Stopping.")
			break
		}

		// Khởi tạo Status cho mỗi point là false
		result := Result{Point: point}
		result.Point.Status = false

		// Kiểm tra nếu Status là false thì bỏ qua
		if !point.Status {
			for j := i + 1; j < len(points); j++ {
				otherPoint := points[j]

				// Create haversine.Coord for both points
				pointCoord := haversine.Coord{Lat: point.Latitude, Lon: point.Longitude}
				otherPointCoord := haversine.Coord{Lat: otherPoint.Latitude, Lon: otherPoint.Longitude}

				// Calculate distance between two points using haversine
				haversineDistance, _ := haversine.Distance(pointCoord, otherPointCoord) // Handle the second return value if needed
				if haversineDistance < 3 && haversineDistance > 0.2{
					// Delay to ensure at least 1 second between API calls
					time.Sleep(1000 * time.Millisecond)
					distance, err := GetDistanceFromHereAPI(point.Latitude, point.Longitude, otherPoint.Latitude, otherPoint.Longitude)
					if err != nil {
						fmt.Printf("Error calculating distance between %s and %s: %v\n", point.ID, otherPoint.ID, err)
						continue
					}
					fmt.Println("thuyen", haversineDistance, distance, apiCallCount)
					if distance <= 5 {
						// Thêm khoảng cách và tọa độ vào cả hai điểm
						result.Distances = append(result.Distances, DistancePoint{
							ID:        otherPoint.ID,
							Distance:  distance,
							Latitude:  otherPoint.Latitude,
							Longitude: otherPoint.Longitude,
						})
					}

					// Nếu j là len(points) - 1, cập nhật Status thành true
					if j == len(points)-1 {
						result.Point.Status = true
					}
				}
			}
		} else {
			fmt.Println("skip status true", point.ID)
		}
		results = append(results, result)
	}
	// Write results to JSON file
	file, err := os.Create("distances.json")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(results)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	fmt.Println("Distances calculated and saved to distances.json")
}
