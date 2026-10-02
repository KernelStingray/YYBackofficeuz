package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	baseURL    = flag.String("url", "http://localhost:8080/api/v1", "API Base URL")
	wsURL      = flag.String("ws", "ws://localhost:8080/ws", "WebSocket URL")
	driversNum = flag.Int("d", 20, "Number of drivers")
	passNum    = flag.Int("p", 80, "Number of passengers")
)

type Coordinate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

var cities = map[string]Coordinate{
	"Tashkent":  {Latitude: 41.2995, Longitude: 69.2401},
	"Samarkand": {Latitude: 39.6542, Longitude: 66.9597},
	"Bukhara":   {Latitude: 39.7747, Longitude: 64.4286},
	"Fergana":   {Latitude: 40.3842, Longitude: 71.7843},
	"Nukus":     {Latitude: 42.4619, Longitude: 59.6166},
	"Navoi":     {Latitude: 40.0844, Longitude: 65.3792},
	"Andijan":   {Latitude: 40.8154, Longitude: 72.2837},
	"Namangan":  {Latitude: 41.0011, Longitude: 71.6676},
	"Karshi":    {Latitude: 38.8615, Longitude: 65.7951},
	"Urgench":   {Latitude: 41.5500, Longitude: 60.6333},
}

func getCityCoords() (Coordinate, Coordinate) {
	keys := make([]string, 0, len(cities))
	for k := range cities {
		keys = append(keys, k)
	}
	startKey := keys[rand.Intn(len(keys))]
	endKey := startKey
	for endKey == startKey {
		endKey = keys[rand.Intn(len(keys))]
	}
	return cities[startKey], cities[endKey]
}

func auth(phone string) (string, error) {
	payload := map[string]string{"phone_number": phone}
	b, _ := json.Marshal(payload)
	_, err := http.Post(*baseURL+"/auth/send-otp", "application/json", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	time.Sleep(500 * time.Millisecond)

	payload2 := map[string]string{
		"phone_number": phone,
		"code":         "000000",
		"device_id":    "sim_" + phone,
		"os_type":      "iOS",
	}
	b2, _ := json.Marshal(payload2)
	resp, err := http.Post(*baseURL+"/auth/verify-otp", "application/json", bytes.NewReader(b2))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("auth failed: %s %s", resp.Status, string(bodyBytes))
	}

	var res struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}
	return res.AccessToken, nil
}

func getRoute(token string, start, end Coordinate) ([]Coordinate, error) {
	url := fmt.Sprintf("%s/routing/route?origin_lat=%f&origin_lon=%f&dest_lat=%f&dest_lon=%f", *baseURL, start.Latitude, start.Longitude, end.Latitude, end.Longitude)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("routing failed: %s %s", resp.Status, string(bodyBytes))
	}

	var routeResp struct {
		Geometry []Coordinate `json:"geometry"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&routeResp); err != nil {
		return nil, err
	}
	return routeResp.Geometry, nil
}

func publishRide(token string, start, end Coordinate) (string, error) {
	payload := map[string]interface{}{
		"origin":          map[string]interface{}{"latitude": start.Latitude, "longitude": start.Longitude},
		"destination":     map[string]interface{}{"latitude": end.Latitude, "longitude": end.Longitude},
		"departure_time":  time.Now().Add(5 * time.Minute).Format(time.RFC3339),
		"available_seats": 4,
		"price_per_seat":  50000,
	}
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", *baseURL+"/rides", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("publish failed: %s %s", resp.Status, string(bodyBytes))
	}

	var res struct {
		ID string `json:"id"`
	}
	json.NewDecoder(resp.Body).Decode(&res)
	return res.ID, nil
}

func simulateDriver(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	phone := fmt.Sprintf("+998001000%03d", id)

	log.Printf("[Driver %d] Starting auth...", id)
	token, err := auth(phone)
	if err != nil {
		log.Printf("[Driver %d] Auth failed: %v", id, err)
		return
	}

	// register driver profile if not exists
	upgradePayload := map[string]interface{}{
		"front_photo_url": "sim.jpg",
		"back_photo_url":  "sim.jpg",
	}
	bUp, _ := json.Marshal(upgradePayload)
	reqUp, _ := http.NewRequest("POST", *baseURL+"/users/driver/upgrade", bytes.NewReader(bUp))
	reqUp.Header.Set("Authorization", "Bearer "+token)
	reqUp.Header.Set("Content-Type", "application/json")
	http.DefaultClient.Do(reqUp)

	// Create car if not exists
	carPayload := map[string]interface{}{
		"make":          "Chevrolet",
		"model":         "Cobalt",
		"year":          2022,
		"color":         "White",
		"license_plate": fmt.Sprintf("01A%03dAA", id),
	}
	bCar, _ := json.Marshal(carPayload)
	reqCar, _ := http.NewRequest("POST", *baseURL+"/users/cars", bytes.NewReader(bCar))
	reqCar.Header.Set("Authorization", "Bearer "+token)
	reqCar.Header.Set("Content-Type", "application/json")
	http.DefaultClient.Do(reqCar)

	for {
		start, end := getCityCoords()
		log.Printf("[Driver %d] Routing from %v to %v...", id, start, end)
		geom, err := getRoute(token, start, end)
		if err != nil || len(geom) == 0 {
			log.Printf("[Driver %d] Routing err: %v", id, err)
			time.Sleep(5 * time.Second)
			continue
		}

		rideID, err := publishRide(token, start, end)
		if err != nil {
			log.Printf("[Driver %d] Publish err: %v", id, err)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Printf("[Driver %d] Published ride %s. Connecting WS...", id, rideID)

		// Connect to WS
		header := http.Header{"Authorization": []string{"Bearer " + token}}
		c, _, err := websocket.DefaultDialer.Dial(*wsURL, header)
		if err != nil {
			log.Printf("[Driver %d] WS dial error: %v", id, err)
			continue
		}

		// Move along geometry
		for _, pt := range geom {
			loc := map[string]interface{}{
				"type": "location_update",
				"payload": map[string]interface{}{
					"latitude":  pt.Latitude,
					"longitude": pt.Longitude,
				},
			}
			c.WriteJSON(loc)
			time.Sleep(2 * time.Second)
		}
		c.Close()

		log.Printf("[Driver %d] Arrived. Completing ride %s", id, rideID)

		// Stop ride
		reqStop, _ := http.NewRequest("POST", *baseURL+"/rides/"+rideID+"/stop", nil)
		reqStop.Header.Set("Authorization", "Bearer "+token)
		http.DefaultClient.Do(reqStop)

		time.Sleep(5 * time.Second)
	}
}

func simulatePassenger(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	phone := fmt.Sprintf("+998002000%03d", id)

	log.Printf("[Pass %d] Starting auth...", id)
	token, err := auth(phone)
	if err != nil {
		log.Printf("[Pass %d] Auth failed: %v", id, err)
		return
	}

	for {
		time.Sleep(time.Duration(rand.Intn(15)+15) * time.Second)
		start, end := getCityCoords()
		url := fmt.Sprintf("%s/rides/search?origin_lat=%f&origin_lon=%f&dest_lat=%f&dest_lon=%f", *baseURL, start.Latitude, start.Longitude, end.Latitude, end.Longitude)
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			continue
		}

		var searchRes struct {
			Rides []struct {
				ID string `json:"id"`
			} `json:"rides"`
		}
		json.NewDecoder(resp.Body).Decode(&searchRes)
		resp.Body.Close()

		if len(searchRes.Rides) > 0 {
			rideID := searchRes.Rides[0].ID
			log.Printf("[Pass %d] Found ride %s. Booking...", id, rideID)

			bookPayload := map[string]interface{}{
				"seats": 1,
			}
			b, _ := json.Marshal(bookPayload)
			reqBook, _ := http.NewRequest("POST", *baseURL+"/rides/"+rideID+"/book", bytes.NewReader(b))
			reqBook.Header.Set("Authorization", "Bearer "+token)
			reqBook.Header.Set("Content-Type", "application/json")
			http.DefaultClient.Do(reqBook)
		}
	}
}

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	log.Printf("Starting Map Simulator. Drivers: %d, Passengers: %d", *driversNum, *passNum)
	var wg sync.WaitGroup

	for i := 1; i <= *driversNum; i++ {
		wg.Add(1)
		go simulateDriver(i, &wg)
		time.Sleep(100 * time.Millisecond) // stagger
	}

	for i := 1; i <= *passNum; i++ {
		wg.Add(1)
		go simulatePassenger(i, &wg)
		time.Sleep(50 * time.Millisecond) // stagger
	}

	wg.Wait()
}
