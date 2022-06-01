package giicache

import (
	"fmt"
	"github.com/gii-cache/consistenthash"
	"io/ioutil"
	"log"
	"net/http"
	url2 "net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_giicache/"
	defaultReplicas = 50
)

var (
	// 通过编译器来检测httpGetter是否实现了PeerGetter接口
	_ PeerGetter = (*httpGetter)(nil)
	_ PeerPicker = (*HTTPPool)(nil)
)

type httpGetter struct {
	baseURL string
}

func (h *httpGetter) Get(groupName, key string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s/%s", h.baseURL, url2.QueryEscape(groupName), url2.QueryEscape(key))
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %s", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %s", err)
	}

	return bytes, nil

}

type HTTPPool struct {
	sync.Mutex

	basePath   string
	self       string
	peers      *consistenthash.Map
	httpGetter map[string]*httpGetter // keyed by e.g. "http://10.0.0.2:8008"
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		basePath: defaultBasePath,
		self:     self,
	}
}

func (h *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", h.self, fmt.Sprintf(format, v...))
}

func (h HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, h.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	h.Log("%s %s", r.Method, r.URL.Path)

	path := r.URL.Path[len(h.basePath):]
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}

func (h *HTTPPool) Set(peers ...string) {
	h.Lock()
	defer h.Unlock()

	h.peers = consistenthash.New(defaultReplicas, nil)
	h.peers.Add(peers...)
	h.httpGetter = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		h.httpGetter[peer] = &httpGetter{baseURL: peer + h.basePath}
	}
}

func (h *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	h.Lock()
	defer h.Unlock()

	if peer := h.peers.Get(key); peer != "" && peer != h.self {
		h.Log("Pick peer %s", peer)
		return h.httpGetter[peer], true
	}

	return nil, false
}
