package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"

	"github.com/devkekops/scanme/internal/app/domainfinder"
	"github.com/devkekops/scanme/internal/app/scanner"
)

type BaseHandler struct {
	*chi.Mux
	fs http.Handler
	df domainfinder.DomainFinder
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

	bh.Get("/", bh.getIndex())
	bh.Get("/api/getDomains", bh.getDomains())
	bh.Get("/api/getTemplates", bh.getTemplates())
	bh.Post("/api/search", bh.searchSubdomains())
	bh.Post("/api/scan", bh.scan())

	return bh
}

func (bh *BaseHandler) getIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bh.fs.ServeHTTP(w, r)
	}
}

func (bh *BaseHandler) getDomains() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var domains = []string{"sbermarket.ru", "sbermarket.tech"}

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

		home, _ := os.UserHomeDir()
		catalog := path.Join(home, "nuclei-templates")
		files, err := ioutil.ReadDir(catalog)
		if err != nil {
			log.Println(err)
		}

		for _, f := range files {
			templates = append(templates, f.Name())
		}
		fmt.Println(templates)

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

func (bh *BaseHandler) searchSubdomains() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var domains []string
		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&domains)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		//subdomains := bh.df.Search(domains)
		subdomains := []string{"https://mx3.sbermarket.ru", "https://www.gift.sbermarket.ru", "https://exponea-gw.sbermarket.ru", "https://adfs.sbermarket.ru", "https://retailers.sbermarket.ru", "https://www.job.sbermarket.ru", "https://calculator.sbermarket.ru", "https://sendcrm.sbermarket.ru", "https://retailers-gw.sbermarket.ru", "https://shp-gw.sbermarket.ru", "https://imgproxy.sbermarket.ru", "https://cdn.sbermarket.ru", "https://happymars.sbermarket.ru", "https://store-plan.sbermarket.ru", "https://www.nedelyadobroty.sbermarket.ru", "https://read.sendcrm.sbermarket.ru", "https://naumen-wfm.sbermarket.ru", "https://api-sberapp.sbermarket.ru", "https://admin-bs.sbermarket.ru", "https://api-deliveryclub.sbermarket.ru", "https://admin-gw.sbermarket.ru", "https://sm-university.sbermarket.ru"}
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
		var domains []string
		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&domains)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		scanner.Scan(domains)
		w.WriteHeader(http.StatusOK)
	}
}
