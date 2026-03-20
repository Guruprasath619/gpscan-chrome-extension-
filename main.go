package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ScanResult struct {
	Port   int    `json:"port"`
	Status string `json:"status"`
}

func scanPort(target string, port int, timeout time.Duration) bool {
	address := fmt.Sprintf("%s:%d", target, port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func parsePorts(portRange string) ([]int, error) {
	var ports []int
	ranges := strings.Split(portRange, ",")
	for _, r := range ranges {
		r = strings.TrimSpace(r)
		if strings.Contains(r, "-") {
			bounds := strings.Split(r, "-")
			if len(bounds) != 2 {
				return nil, fmt.Errorf("invalid port range: %s", r)
			}
			start, err1 := strconv.Atoi(bounds[0])
			end, err2 := strconv.Atoi(bounds[1])
			if err1 != nil || err2 != nil || start > end {
				return nil, fmt.Errorf("invalid port range: %s", r)
			}
			for p := start; p <= end; p++ {
				ports = append(ports, p)
			}
		} else {
			p, err := strconv.Atoi(r)
			if err != nil {
				return nil, fmt.Errorf("invalid port: %s", r)
			}
			ports = append(ports, p)
		}
	}
	return ports, nil
}

func enableCORS(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func apiScanHandler(w http.ResponseWriter, r *http.Request) {
	// Debug log
	log.Printf("Received request from %s: %s %s", r.RemoteAddr, r.Method, r.URL.Path)

	enableCORS(&w)
	if r.Method == "OPTIONS" {
		return
	}

	target := r.URL.Query().Get("target")
	portsStr := r.URL.Query().Get("ports")
	timeoutStr := r.URL.Query().Get("timeout")
	timeout := 1 * time.Second
	if timeoutStr != "" {
		if t, err := strconv.Atoi(timeoutStr); err == nil && t > 0 {
			timeout = time.Duration(t) * time.Second
		}
	}

	ports, err := parsePorts(portsStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var results []ScanResult
	// Basic concurrency for speed
	resultChan := make(chan ScanResult)
	for _, port := range ports {
		go func(p int) {
			open := scanPort(target, p, timeout)
			status := "closed"
			if open {
				status = "open"
			}
			resultChan <- ScanResult{Port: p, Status: status}
		}(port)
	}

	for range ports {
		results = append(results, <-resultChan)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "GPSCAN Backend is Running! (HTTP/2 + TLS)\n\nEndpoint: /api/scan")
	})
	mux.HandleFunc("/api/scan", apiScanHandler)

	// Fallback/Main HTTPS server standard (Go automatically enables HTTP/2 over TLS)
	fmt.Println("gpscan (Go Backend) running on https://localhost:7373 (HTTP/2 enabled)")
	fmt.Println("Make sure you have 'cert.pem' and 'key.pem' in this directory.")

	err := http.ListenAndServeTLS(":7373", "cert.pem", "key.pem", mux)
	if err != nil {
		log.Fatal(err)
	}
}
