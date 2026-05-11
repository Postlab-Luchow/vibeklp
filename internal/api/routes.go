package api

import (
	"github.com/gorilla/mux"
	"github.com/musche/klp/internal/storage"
)

// SetupRoutes configures all API routes
func SetupRoutes(store *storage.Storage) *mux.Router {
	router := mux.NewRouter()
	handler := NewHandler(store)

	// API routes
	api := router.PathPrefix("/api").Subrouter()

	// Venues
	api.HandleFunc("/venues", handler.GetVenues).Methods("GET")
	api.HandleFunc("/venues/{id}", handler.GetVenue).Methods("GET")

	// Events
	api.HandleFunc("/events", handler.GetEvents).Methods("GET")
	api.HandleFunc("/events/{id}", handler.GetEvent).Methods("GET")

	// Exhibitions
	api.HandleFunc("/exhibitions", handler.GetExhibitions).Methods("GET")
	api.HandleFunc("/exhibitions/{id}", handler.GetExhibition).Methods("GET")

	// Search
	api.HandleFunc("/search", handler.Search).Methods("GET")

	// Calendar
	api.HandleFunc("/calendar", handler.GetCalendar).Methods("GET")

	// Categories
	api.HandleFunc("/categories", handler.GetCategories).Methods("GET")
	api.HandleFunc("/event-categories", handler.GetEventCategories).Methods("GET")

	// Statistics
	api.HandleFunc("/stats", handler.GetStats).Methods("GET")

	return router
}
