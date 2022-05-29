package main

import (
	"fmt"
	giicache "github.com/gii-cache"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func main() {
	giicache.NewGroup("scores", 2<<10, giicache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB](callback) search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	// Start HTTP server.
	addr := "localhost:8080"
	peers := giicache.NewHTTPPool(addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
