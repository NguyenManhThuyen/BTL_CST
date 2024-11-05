package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"math"
	"log"
)

type Address struct {
	Label       string `json:"label"`
	CountryCode string `json:"countryCode"`
	CountryName string `json:"countryName"`
	County      string `json:"county"`
	City        string `json:"city"`
	District    string `json:"district"`
	Street      string `json:"street"`
	PostalCode  string `json:"postalCode"`
	HouseNumber string `json:"houseNumber,omitempty"`
}

type Position struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Contact struct {
	Phone []struct {
		Value     string `json:"value"`
		Categories []struct {
			ID string `json:"id"`
		} `json:"categories,omitempty"`
	} `json:"phone,omitempty"`
	WWW []struct {
		Value     string `json:"value"`
		Categories []struct {
			ID string `json:"id"`
		} `json:"categories,omitempty"`
	} `json:"www,omitempty"`
}

type Category struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Primary bool   `json:"primary"`
}

type Item struct {
	Title       string   `json:"title"`
	ID          string   `json:"id"`
	ResultType  string   `json:"resultType"`
	Address     Address  `json:"address"`
	Position    Position `json:"position"`
	Access      []Position `json:"access,omitempty"`
	Distance    int        `json:"distance"`
	Categories  []Category `json:"categories"`
}

type HereResponse struct {
	Items []Item `json:"items"`
}

// GetRandomPointsFromAPI fetches random points from HERE API based on categories
func GetRandomPointsFromAPI(lat, lon float64, categoryID string) ([]Item, error) {
	apiKey := "VixcUdhSK9ZNGkHVwM0dgY-f68twZUnhzYzTIASxGvQ"
	url := fmt.Sprintf("https://browse.search.hereapi.com/v1/browse?at=%.6f,%.6f&categories=%s&limit=100&apikey=%s", lat, lon, categoryID, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var hereResp HereResponse
	err = json.Unmarshal(body, &hereResp)
	if err != nil {
		return nil, err
	}

	return hereResp.Items, nil
}

// ExportPointsToJSON exports the generated points to a JSON file
func ExportPointsToJSON(items []Item, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(items)
}

func FetchHereMapData(lat, lon float64, apiKey string) (HereResponse, error) {
	url := fmt.Sprintf("https://browse.search.hereapi.com/v1/browse?at=%.6f,%.6f&limit=100&apikey=%s", lat, lon, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return HereResponse{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return HereResponse{}, err
	}

	var hereResp HereResponse
	err = json.Unmarshal(body, &hereResp)
	if err != nil {
		return HereResponse{}, err
	}

	return hereResp, nil
}

// FetchPointsFromStringArray fetches points from HERE API based on an array of coordinate strings
func FetchPointsFromStringArray(coords []string, apiKey string) (HereResponse, error) {
	var points HereResponse

	for _, coordStr := range coords {
		var coord Position
		_, err := fmt.Sscanf(coordStr, "%f, %f", &coord.Lat, &coord.Lng)
		if err != nil {
			return HereResponse{}, fmt.Errorf("invalid coordinate format: %s", coordStr)
		}

		url := fmt.Sprintf("https://browse.search.hereapi.com/v1/browse?at=%.6f,%.6f&apikey=%s&limit=100", coord.Lat, coord.Lng, apiKey)

		resp, err := http.Get(url)
		if err != nil {
			return HereResponse{}, err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return HereResponse{}, err
		}

		var hereResp HereResponse
		err = json.Unmarshal(body, &hereResp)
		if err != nil {
			return HereResponse{}, err
		}

		for _, item := range hereResp.Items {
			if item.Address.HouseNumber != "" {
				points.Items = append(points.Items, item)
				break
			}
		}
	}

	return points, nil
}

// SavePointsToJSON saves the points to a JSON file
func SavePointsToJSON(points HereResponse, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(points)
}

// LoadPointsFromJSON loads points from a JSON file
func LoadPointsFromJSON(filename string) ([]Item, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var items []Item
	err = json.Unmarshal(file, &items)
	if err != nil {
		return nil, err
	}
	fmt.Println(len(items))

	return items, nil
}

// FetchFilteredPointsFromAPI fetches points from HERE API based on coordinates from a JSON file
func FetchFilteredPointsFromAPI(filename string, apiKey string) ([]Item, error) {
	var filteredPoints []Item
	idSet := make(map[string]struct{}) // To track unique IDs

	// Load points from the JSON file
	items, err := LoadPointsFromJSON(filename)
	if err != nil {
		return nil, err
	}

	categories := []string{
		"200-2000-0011",
		"100-1000-0000",
		"700-7450-0114",
		"200-2100-0019",
		"400-4100-0036",
		"600-6100-0062",
		"600-6900-0103",
		"600-6000-0061",
		"600-6300-0064",
		"600-6800-0000",
		"700-7000-0107",
	}

	for _, item := range items {
		lat := item.Position.Lat
		lng := item.Position.Lng

		// Select 2 random category IDs
		categoryIDs := make([]string, 0, 2)
		for len(categoryIDs) < 2 {
			categoryID := categories[rand.Intn(len(categories))]
			if !contains(categoryIDs, categoryID) {
				categoryIDs = append(categoryIDs, categoryID)
			}
		}

		for _, categoryID := range categoryIDs {
			url := fmt.Sprintf("https://browse.search.hereapi.com/v1/browse?at=%.6f,%.6f&categories=%s&limit=100&apikey=%s", lat, lng, categoryID, apiKey)

			resp, err := http.Get(url)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}

			var hereResp HereResponse
			err = json.Unmarshal(body, &hereResp)
			if err != nil {
				return nil, err
			}

			// Collect points that meet the criteria
			count := 0
			for _, place := range hereResp.Items {
				if place.Address.HouseNumber != "" && place.Distance > 200 && place.Distance < 1500 {
					// Check for unique ID before appending
					if _, exists := idSet[place.ID]; !exists {
						filteredPoints = append(filteredPoints, place) // Add address with houseNumber to filteredPoints
						idSet[place.ID] = struct{}{} // Mark this ID as seen
						count++
					}
				}
				if count >= 3 { // Ensure at least 4 valid points
					break
				}
			}
		}
	}
	fmt.Println(len(filteredPoints))
	return filteredPoints, nil
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// SaveFilteredPointsToJSON saves the filtered points to a JSON file
func SaveFilteredPointsToJSON(points []Item, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(points)
}

// Hàm tính khoảng cách giữa hai điểm (lat1, lng1) và (lat2, lng2) bằng công thức Haversine
func haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371000 // Bán kính trái đất (m)
	phi1 := lat1 * math.Pi / 180
	phi2 := lat2 * math.Pi / 180
	deltaPhi := (lat2 - lat1) * math.Pi / 180
	deltaLambda := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(deltaPhi/2)*math.Sin(deltaPhi/2) +
		math.Cos(phi1)*math.Cos(phi2)*
		math.Sin(deltaLambda/2)*math.Sin(deltaLambda/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c // Khoảng cách (m)
}

// Hàm kiểm tra số lượng điểm có khoảng cách < 50m
func countClosePoints(points []Item) int {
	count := 0
	for i := 0; i < len(points); i++ {
		for j := i + 1; j < len(points); j++ {
			if haversine(points[i].Position.Lat, points[i].Position.Lng, points[j].Position.Lat, points[j].Position.Lng) < 50 {
				count++
			}
		}
	}
	return count
}

func main() {
	// // Load points from JSON file
	// jsonFile, err := os.ReadFile("5diem.json")
	// if err != nil {
	// 	fmt.Println("Error reading 5diem.json:", err)
	// 	return
	// }

	// var inputPoints []struct {
	// 	ID       string  `json:"id"`
	// 	Position struct {
	// 		Lat float64 `json:"lat"`
	// 		Lng float64 `json:"lng"`
	// 	} `json:"position"`
	// }
	// err = json.Unmarshal(jsonFile, &inputPoints)
	// if err != nil {
	// 	fmt.Println("Error parsing 5diem.json:", err)
	// 	return
	// }

	// var allRandomItems []Item
	// categories := []string{
	// 	"400-4100-0036", // bustion
	// 	"600-6100-0062",
	// 	"600-6900-0103",
	// }

	// for _, point := range inputPoints {
	// 	for _, category := range categories {
	// 		randomItems, err := GetRandomPointsFromAPI(point.Position.Lat, point.Position.Lng, category)
	// 		if err != nil {
	// 			fmt.Printf("Error fetching random points for category %s: %v\n", category, err)
	// 			continue
	// 		}
	// 		allRandomItems = append(allRandomItems, randomItems...) // Add random items to the existing items
	// 	}
	// }

	// // Export all random items to a new JSON file
	// err = ExportPointsToJSON(allRandomItems, "random_points.json")
	// if err != nil {
	// 	fmt.Println("Error exporting points to JSON:", err)
	// 	return
	// }

	// fmt.Println("Random points generated and saved to random_points.json")

	// coords := []string{
	// 	"10.793711, 106.669042",
	// 	"10.795655, 106.663980",
	// 	"10.797017, 106.664614",
	// 	"10.793729, 106.668998",
	// 	"10.797958, 106.671525",
	// 	"10.794230, 106.677936",
	// 	"10.795823, 106.681884",
	// 	"10.798624, 106.687578",
	// 	"10.797200, 106.693704",
	// 	"10.800623, 106.696463",
	// 	"10.803000, 106.709379",
	// 	"10.809959, 106.712614",
	// 	"10.813228, 106.716309",
	// 	"10.813331, 106.709553",
	// 	"10.820045, 106.703944",
	// 	"10.811053, 106.701132",
	// 	"10.811245, 106.695512",
	// 	"10.811219, 106.690986",
	// 	"10.815561, 106.688703",
	// 	"10.814545, 106.686853",
	// 	"10.810707, 106.690211",
	// 	"10.808324, 106.688073",
	// 	"10.809943, 106.683963",
	// 	"10.810820, 106.684207",
	// 	"10.818804, 106.682417",
	// 	"10.821990, 106.678851",
	// 	"10.818064888941008, 106.67613864104564",
	// 	"10.815304, 106.675673",
	// 	"10.816463, 106.670347",
	// 	"10.813134, 106.669986",
	// 	"10.808684, 106.669795",
	// 	"10.805758, 106.671680",
	// 	"10.802560, 106.676569",
	// 	"10.801446, 106.652391",
	// 	"10.789098, 106.660689",
	// 	"10.785349, 106.671271",
	// 	"10.790080, 106.683512",
	// 	"10.795003, 106.697940",
	// 	"10.821568, 106.677514",
	// 	"10.813749, 106.682634",
	// 	"10.812715, 106.678636",
	// 	"10.820070, 106.703979",
	// 	"10.816436, 106.670399",
	// 	"10.804138, 106.683806",
	// 	"10.811520, 106.681434",
	// 	"10.812808, 106.687818",
	// 	"10.794701, 106.709341",
	// 	"10.796343, 106.705873",
	// 	"10.794112, 106.704696",
	// 	"10.792323, 106.703380",
	// 	"10.791899, 106.695831",
	// 	"10.785081, 106.689121",
	// 	"10.796644, 106.701241",
	// 	"10.802163, 106.701124",
	// 	"10.797678, 106.698870",
	// 	"10.799316, 106.693559",
	// 	"10.822286, 106.706998",
	// 	"10.825297, 106.706683",
	// 	"10.826890, 106.703149",
	// 	"10.819958, 106.694328",
	// 	"10.789386, 106.716558",
	// 	"10.805044, 106.721114",
	// 	"10.811182, 106.705201",
	// 	"10.826238, 106.689347",
	// }

	//apiKey := "n3iNc8FnNXEBLWCfBZYJMd3nWpvfTt1Fm_iol6i3KEY"

	// points, err := FetchPointsFromStringArray(coords, apiKey)
	// if err != nil {
	// 	fmt.Println("Error fetching points:", err)
	// 	return
	// }

	// fmt.Println(len(points.Items))

	// err = SavePointsToJSON(points, "List38point.json")
	// if err != nil {
	// 	fmt.Println("Error saving points to JSON:", err)
	// 	return
	// }

	// fmt.Println("Points with house numbers saved to points_with_house_numbers.json")

	
	// filteredPoints, err := FetchFilteredPointsFromAPI("List38point.json", apiKey)
	// if err != nil {
	// 	fmt.Println("Error fetching filtered points:", err)
	// 	return
	// }

	// err = SaveFilteredPointsToJSON(filteredPoints, "filtered_points.json")
	// if err != nil {
	// 	fmt.Println("Error saving filtered points to JSON:", err)
	// 	return
	// }

	// fmt.Println("Filtered points saved to filtered_points.json")

	// Đọc dữ liệu từ file JSON
	// Đọc dữ liệu từ file JSON
	var points []Item
	file, err := os.Open("filtered_points.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&points); err != nil {
		log.Fatal(err)
	}

	// Tạo một danh sách mới để lưu các điểm không bị xóa
	var filteredPoints []Item
	isDeleted := make(map[int]bool) // Đánh dấu các chỉ số điểm đã bị xóa

	// Kiểm tra số lượng điểm có khoảng cách < 50m
	for i := 0; i < len(points); i++ {
		for j := i + 1; j < len(points); j++ {
			if haversine(points[i].Position.Lat, points[i].Position.Lng, points[j].Position.Lat, points[j].Position.Lng) < 50 {
				isDeleted[i] = true
				isDeleted[j] = true
			}
		}
	}

	// Lưu các điểm không bị xóa vào danh sách mới
	for i, point := range points {
		if !isDeleted[i] {
			filteredPoints = append(filteredPoints, point)
		}
	}

	// Lưu lại danh sách mới vào file JSON
	err = SaveFilteredPointsToJSON(filteredPoints, "filtered_points.json")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(filteredPoints))
	fmt.Printf("Đã xóa các điểm có khoảng cách < 50m và lưu lại vào file filtered_points.json\n")

}
