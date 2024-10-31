package main

import (
	"golangbackend/controller"
	"golangbackend/service"

	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {

	router := mux.NewRouter().StrictSlash(true).PathPrefix("/api/v1").Subrouter()
	converter := service.NewConverterService()
	imgController := controller.NewConverterController(router, converter)

	imgController.SetUpRouter()

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})

	log.Printf("public server Listening at http://%s:%s", "127.0.0.1", ":5001")
	log.Fatal(http.ListenAndServe(":5001", handlers.CORS(headersOk, methodsOk, originsOk)(router)))

}
