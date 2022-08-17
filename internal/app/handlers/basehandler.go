package handlers

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/gorilla/websocket"
	"github.com/projectdiscovery/nuclei/v2/pkg/output"

	"github.com/devkekops/scanme/internal/app/domainfinder"
	"github.com/devkekops/scanme/internal/app/scanner"
)

type BaseHandler struct {
	*chi.Mux
	fs http.Handler
	df domainfinder.DomainFinder
}

type Message struct {
	Event string          `json:"event"`
	Msg   json.RawMessage `json:"msg"`
}

type Scan struct {
	Domains   []string `json:"domains"`
	Templates []string `json:"templates"`
}

func NewBaseHandler(df domainfinder.DomainFinder) *BaseHandler {
	root := "./internal/app/static"
	fs := http.FileServer(http.Dir(root))

	bh := &BaseHandler{
		Mux: chi.NewMux(),
		fs:  fs,
		df:  df,
	}

	bh.Use(middleware.Logger)

	bh.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	  }))

	//bh.Get("/", bh.getIndex())
	bh.Handle("/*", fs)
	bh.Get("/api/getDomains", bh.getDomains())
	bh.Get("/api/getTemplates", bh.getTemplates())
	//bh.Post("/api/search", bh.searchSubdomains())
	//bh.Post("/api/scan", bh.scan())
	bh.Handle("/ws", bh.wsEndpoint())

	return bh
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (bh *BaseHandler) wsEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		connection, _ := upgrader.Upgrade(w, r, nil)

		for {
			_, message, _ := connection.ReadMessage()

			//connection.WriteMessage(websocket.TextMessage, message)
			go messageHandler(message, bh, connection)
		}
	}
}

func messageHandler(message []byte, bh *BaseHandler, conn *websocket.Conn) {
	if len(message) > 0 {
		//fmt.Println(string(message))
		var inMsg Message
		err := json.Unmarshal(message, &inMsg)
		if err != nil {
			return
		}
		switch inMsg.Event {
		case "search":
			buf, _ := json.Marshal("start")
			outMsg := Message{Event: "search", Msg: buf}
			conn.WriteJSON(outMsg)

			var domains []string
			json.Unmarshal(inMsg.Msg, &domains)
			//fmt.Println(domains)
			subdomainCh := make(chan []string)
			go bh.df.Search(domains, subdomainCh)
			for {
				subdomains, ok := <-subdomainCh
				if !ok {
					break
				}
				buf, _ := json.Marshal(subdomains)
				outMsg := Message{Event: "search", Msg: buf}
				conn.WriteJSON(outMsg)
			}

			buf, _ = json.Marshal("finish")
			outMsg = Message{Event: "search", Msg: buf}
			conn.WriteJSON(outMsg)

		case "scan":
			buf, _ := json.Marshal("start")
			outMsg := Message{Event: "scan", Msg: buf}
			conn.WriteJSON(outMsg)

			var newScan Scan
			json.Unmarshal(inMsg.Msg, &newScan)

			resultCh := make(chan *output.ResultEvent)
			go scanner.Scan(newScan.Domains, newScan.Templates, resultCh)
			for {
				result, ok := <-resultCh
				if !ok {
					break
				}
				buf, _ := json.Marshal(result)
				outMsg := Message{Event: "scan", Msg: buf}
				conn.WriteJSON(outMsg)
			}

			buf, _ = json.Marshal("finish")
			outMsg = Message{Event: "scan", Msg: buf}
			conn.WriteJSON(outMsg)
		}
	}
}

/*func (bh *BaseHandler) getIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bh.fs.ServeHTTP(w, r)
	}
}*/

func (bh *BaseHandler) getDomains() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var domains = []string{"sbermarket.ru", "sbermarket.tech", "instamart.ru", "sbmt.io"}

		buf, err := json.Marshal(domains)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(buf)
		if err != nil {
			log.Println(err)
		}
	}
}

func (bh *BaseHandler) getTemplates() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var templates []string

		//home, _ := os.UserHomeDir()
		//catalog := path.Join(home, "nuclei-templates")
		cwd, _ := os.Getwd()
		catalog := filepath.Join(cwd, "/templates")

		files, err := ioutil.ReadDir(catalog)
		if err != nil {
			log.Println(err)
		}

		for _, f := range files {
			templates = append(templates, f.Name())
		}
		//fmt.Println(templates)

		buf, err := json.Marshal(templates)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(buf)
		if err != nil {
			log.Println(err)
		}
	}
}

/*func (bh *BaseHandler) searchSubdomains() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var domains []string
		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&domains)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		subdomains := bh.df.Search(domains)
		//subdomains := []string{"https://mx3.sbermarket.ru", "https://www.gift.sbermarket.ru", "https://exponea-gw.sbermarket.ru"}
		buf, err := json.Marshal(subdomains)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(buf)
		if err != nil {
			log.Println(err)
		}
	}
}

func (bh *BaseHandler) scan() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var newScan Scan

		if err := json.NewDecoder(req.Body).Decode(&newScan); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Println(err)
			return
		}

		results := scanner.Scan(newScan.Domains, newScan.Templates)
		if len(results) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		buf, err := json.Marshal(results)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(buf)
		if err != nil {
			log.Println(err)
		}
	}
}*/
