package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	persist "sfhb/src/persist"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/acme/autocert"
)

// Article - Our struct for all articles
type Article struct {
	Id      uuid.UUID `json:"id"`
	Title   string    `json:"title"`
	Type    string    `json:"type"`
	Content string    `json:"content"`
	Created time.Time `json:"created"`
}

var Articles []Article

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "sfhb api v.0.1")
}

func returnAllArticles(w http.ResponseWriter, r *http.Request) {
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding,X-Session-Token, X-CSRF-Token, Authorization")
	}
	if r.Method == "OPTIONS" {
		return
	}
	if _, err := os.Stat("./data.json"); err == nil {
		if err := persist.Load("./data.json", &Articles); err != nil {
			log.Fatalln(err)
		}
	}
	// sort newest first
	sort.Slice(Articles[:], func(i, j int) bool {
		return Articles[i].Created.Unix() > Articles[j].Created.Unix()
	})
	json.NewEncoder(w).Encode(Articles)
}

func returnSingleArticle(w http.ResponseWriter, r *http.Request) {
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding,X-Session-Token, X-CSRF-Token, Authorization")
	}
	if r.Method == "OPTIONS" {
		return
	}

	if _, err := os.Stat("./data.json"); err == nil {
		if err := persist.Load("./data.json", &Articles); err != nil {
			log.Fatalln(err)
		}
	}

	vars := mux.Vars(r)
	key := vars["id"]

	for _, article := range Articles {
		u, err := uuid.Parse(key)
		if err != nil {
			log.Fatalln(err)
		}

		if article.Id == u {
			json.NewEncoder(w).Encode(article)
		}
	}
}

func createNewArticle(w http.ResponseWriter, r *http.Request) {
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding,X-Session-Token, X-CSRF-Token, Authorization")
	}
	if r.Method == "OPTIONS" {
		return
	}

	token := r.Header.Get("X-Session-Token")

	if token == os.Getenv("TOKEN") {
		// get the body of our POST request
		// unmarshal this into a new Article struct
		// append this to our Articles array.
		reqBody, _ := ioutil.ReadAll(r.Body)
		var article Article
		json.Unmarshal(reqBody, &article)
		// set creation timestamp to now
		article.Created = time.Now()
		article.Id = uuid.New()
		// update our global Articles array to include
		// our new Article
		Articles = append(Articles, article)

		json.NewEncoder(w).Encode(article)

		if err := persist.Save("./data.json", Articles); err != nil {
			log.Fatalln(err)
		}
	} else {
		// Write an error and stop the handler chain
		http.Error(w, "Forbidden", http.StatusForbidden)
	}

}

func deleteArticle(w http.ResponseWriter, r *http.Request) {
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding,X-Session-Token, X-CSRF-Token, Authorization")
	}
	if r.Method == "OPTIONS" {
		return
	}
	token := r.Header.Get("X-Session-Token")

	if token == os.Getenv("TOKEN") {
		vars := mux.Vars(r)
		id := vars["id"]

		for index, article := range Articles {
			u, err := uuid.Parse(id)
			if err != nil {
				log.Fatalln(err)
			}

			if article.Id == u {
				Articles = append(Articles[:index], Articles[index+1:]...)
			}
		}

		if err := persist.Save("./data.json", Articles); err != nil {
			log.Fatalln(err)
		}
	} else {
		// Write an error and stop the handler chain
		http.Error(w, "Forbidden", http.StatusForbidden)
	}

}

func handleRequests() {

	certManager := autocert.Manager{
		Prompt: autocert.AcceptTOS,
		//HostPolicy: autocert.HostWhitelist("example.com"),
		Cache: autocert.DirCache("./certs"), //Folder for storing certificates
	}

	r := mux.NewRouter()
	r.HandleFunc("/", homePage)
	r.HandleFunc("/articles", returnAllArticles).Methods("GET", "OPTIONS")
	r.HandleFunc("/article", createNewArticle).Methods("POST", "OPTIONS")
	r.HandleFunc("/article/{id}", deleteArticle).Methods("DELETE", "OPTIONS")
	r.HandleFunc("/article/{id}", returnSingleArticle).Methods("GET", "OPTIONS")

	server := &http.Server{
		Addr:    ":https",
		Handler: r,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
	}

	go http.ListenAndServe(":http", certManager.HTTPHandler(nil))

	port := ":" + os.Getenv("PORT")
	if port == ":" {
		port = ":8080"
	}
	log.Printf("Serving under port %s\n", port)

	if os.Getenv("TOKEN") == "" {
		log.Printf("Authentication is disabled! Please set TOKEN environment variable to enable.")
	}

	log.Fatal(server.ListenAndServeTLS("", "")) //Key and cert are coming from Let's Encrypt

	// http.ListenAndServe(port, nil)
}

func main() {

	// load data at startup
	if _, err := os.Stat("./data.json"); err == nil {
		// data.json should exist
		if err := persist.Load("./data.json", &Articles); err != nil {
			log.Fatalln(err)
		}
	}
	// start http server
	handleRequests()
}
