package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

type listing struct {
	Id          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Price       string `json:"price"`
	City        string `json:"city"`
	CreatedAt   string `json:"created_at"`
}

type ListingHandler struct {
	db *sql.DB
}

func NewListingHandler(db *sql.DB) *ListingHandler {
	return &ListingHandler{
		db: db,
	}
}

// attach this method to the listing handler struct
// concept is called method receiver -> need to check when we use reference pointer and when we use the value
func (lh ListingHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := lh.db.Query(`SELECT id, title, description, price, city, created_at
		FROM listings
		ORDER BY created_at DESC
		LIMIT 100`)
	if err != nil {
		log.Printf("listing.query: %v", err)
		http.Error(w, "internal erro", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	listings := []listing{}
	for rows.Next() {
		var l listing
		if err := rows.Scan(&l.Id, &l.Title, &l.Description, &l.Price, &l.City, &l.CreatedAt); err != nil {
			log.Printf("rows.scan: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		listings = append(listings, l)
	}
	if err := rows.Err(); err != nil {
		log.Printf("rows.err: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(listings)

}

//// Dependency injection via closure factory
// this is bad idea so we use the above pattern

// func DeleteListings(db *sql.DB) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		id := r.PathValue("id")
// 		if id == "" {
// 			http.Error(w, "missing id", http.StatusBadRequest)
// 			return
// 		}
// 		query := "DELETE FROM listings WHERE id=$1"
// 		result, err := db.Exec(query, id)
// 		if err != nil {
// 			log.Print("delete: %w", err)
// 			http.Error(w, "failed to delete user", http.StatusInternalServerError)
// 			return
// 		}
// 		rowsAffected, _ := result.RowsAffected()
// 		if rowsAffected == 0 {
// 			http.Error(w, "no user found", http.StatusBadRequest)
// 			return
// 		}
// 		w.WriteHeader(http.StatusNoContent)
// 	}
// }

func (lh ListingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	query := "DELETE FROM listings WHERE id=$1"
	result, err := lh.db.Exec(query, id)
	if err != nil {
		log.Print("delete: %w", err)
		http.Error(w, "failed to delete user", http.StatusInternalServerError)
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "no user found", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
