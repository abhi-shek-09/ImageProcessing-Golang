package main 

import (
	"image-processing/internal/database"
	"image-processing/internal/middleware"
	"image-processing/internal/handlers"
	"image-processing/internal/auth"
	"github.com/gorilla/mux"
	"net/http"
	"log"
)

func main(){
	database.ConnectDB()
	defer database.CloseDB()
	
	r := mux.NewRouter()

	r.HandleFunc("/register", auth.RegisterUser).Methods("POST")
	r.HandleFunc("/login", auth.LoginUser).Methods("POST")

	r.Use(middleware.RateLimitMiddleware)
	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.AuthMiddleWare)

	api.HandleFunc("/upload", handlers.UploadImage).Methods("POST")
	api.HandleFunc("/images", handlers.GetImages).Methods("GET")
	api.HandleFunc("/images/{id}", handlers.GetImageByID).Methods("GET")
	api.HandleFunc("/images/{id}/transform", handlers.TransformImage).Methods("POST")

	r.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	log.Println("Server is running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}