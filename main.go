package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"sync"

	svg "github.com/ajstarks/svgo"
)

var cache sync.Map

type IP struct {
	IP          string `json:"ip"`
	CountryName string `json:"country_name"`
	City        string `json:"city"`
	Location    struct {
		CountryFlagEmoji string `json:"country_flag_emoji"`
	} `json:"location"`
}

func main() {
	http.HandleFunc("/", greet)
	err := http.ListenAndServe(os.Getenv("ADDRESS"), nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

// On error: draw only greeting text
func greet(w http.ResponseWriter, req *http.Request) {
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		fmt.Println(err.Error())
	}
	ip, err := lookup(host)
	if err != nil {
		fmt.Println(err.Error())
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("charset", "utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	s := svg.New(w)
	s.Start(400, 500)
	s.Link("https://doublegrey.dev", "https://doublegrey.dev")
	s.Text(200, 100, ip.IP, "text-anchor:middle;font-size:30px;font-family:sans-serif")
	s.Text(200, 135, fmt.Sprintf("%s, %s %s", ip.CountryName, ip.City, ip.Location.CountryFlagEmoji), "text-anchor:middle;font-size:20px;font-family:sans-serif")
	// s.Image(15, 150, 370, 300, "https://media1.tenor.com/images/b85ecfd8cff510945f6659786312ba28/tenor.gif?itemid=8126276")
	s.LinkEnd()
	s.End()
}

func lookup(ip string) (IP, error) {
	if value, exists := cache.Load(ip); exists {
		return value.(IP), nil
	}
	resp, err := http.Get(fmt.Sprintf("http://api.ipstack.com/%s?access_key=%s", ip, os.Getenv("API_KEY")))
	if err != nil {
		return IP{}, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return IP{}, err
	}
	value := IP{}
	err = json.Unmarshal(body, &value)
	if err != nil {
		return IP{}, err
	}
	cache.Store(ip, value)
	return value, nil
}
