package main

import (
	"flag"
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

func createGroup() *giicache.Group {
	return giicache.NewGroup("scores", 2<<10, giicache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB](callback) search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheServer(addr string, addrs []string, gii *giicache.Group) {
	peers := giicache.NewHTTPPool(addr)
	peers.Set(addrs...)
	gii.RegisterPeers(peers)
	log.Println("gii cache server start at: ", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startApiServer(addr string, gii *giicache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := gii.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())
		}))

	log.Println("fontend server is running at: ", addr)
	log.Fatal(http.ListenAndServe(addr[7:], nil))
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "server port")
	flag.BoolVar(&api, "api", false, "start api server")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, addr := range addrMap {
		addrs = append(addrs, addr)
	}

	gii := createGroup()
	if api {
		go startApiServer(apiAddr, gii)
	}

	startCacheServer(addrMap[port], addrs, gii)
}
