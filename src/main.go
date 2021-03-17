package main

import (
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
	r := mux.NewRouter()
	r.HandleFunc("/", homePage)
	r.HandleFunc("/articles", returnAllArticles)
	r.HandleFunc("/article", createNewArticle).Methods("POST")
	r.HandleFunc("/article/{id}", deleteArticle).Methods("DELETE")
	r.HandleFunc("/article/{id}", returnSingleArticle)
	http.Handle("/", &MyServer{r})

	port := ":" + os.Getenv("PORT")
	if port == ":" {
		port = ":8080"
	}
	log.Printf("Serving under port %s\n", port)

	if os.Getenv("TOKEN") == "" {
		log.Printf("Authentication is disabled! Please set TOKEN environment variable to enable.")
	}
	http.ListenAndServe(port, nil)
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

type MyServer struct {
	r *mux.Router
}

func (s *MyServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if origin := req.Header.Get("Origin"); origin != "" {
		rw.Header().Set("Access-Control-Allow-Origin", origin)
		rw.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		rw.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding,X-Session-Token, X-CSRF-Token, Authorization")
	}
	// Stop here if its Preflighted OPTIONS request
	if req.Method == "OPTIONS" {
		return
	}
	// Lets Gorilla work
	s.r.ServeHTTP(rw, req)
}
