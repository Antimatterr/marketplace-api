package handlers

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/Antimatterr/marketplace-api/internal/httpx"
	"github.com/Antimatterr/marketplace-api/internal/middleware"
)

type listing struct {
	Id          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       int64     `json:"price"`
	City        string    `json:"city"`
	CreatedAt   time.Time `json:"created_at"`
}

type ListingHandler struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewListingHandler(db *sql.DB, logger *slog.Logger) *ListingHandler {
	return &ListingHandler{
		db:     db,
		logger: logger,
	}
}

// attach this method to the listing handler struct
// concept is called method receiver -> need to check when we use reference pointer and when we use the value
func (lh ListingHandler) List(w http.ResponseWriter, r *http.Request) {
	//request scoped context
	ctx := r.Context()
	rows, err := lh.db.QueryContext(ctx, `SELECT id, title, description, price, city, created_at
		FROM listings
		ORDER BY created_at DESC
		LIMIT 100`)
	if err != nil {
		lh.logger.Error("listing.query failed", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "internal server error", httpx.CodeInternalError)
		return
	}
	defer rows.Close()
	listings := []listing{}
	for rows.Next() {
		var l listing
		if err := rows.Scan(&l.Id, &l.Title, &l.Description, &l.Price, &l.City, &l.CreatedAt); err != nil {
			lh.logger.Error("rows.scan failed", "error", err)
			httpx.Error(w, http.StatusInternalServerError, "internal server error", httpx.CodeInternalError)
			return
		}
		listings = append(listings, l)
	}
	if err := rows.Err(); err != nil {
		lh.logger.Error("rows.err failer", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "internal server error", httpx.CodeInternalError)
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
	ctx := r.Context()
	requestId := middleware.GetRequestIdFromContext(ctx)
	id := r.PathValue("id")

	query := "DELETE FROM listings WHERE id=$1"
	result, err := lh.db.ExecContext(ctx, query, id)
	if err != nil {
		lh.logger.Error("delete failed", "listing_id", id, "requestId", requestId, "error", err)
		httpx.Error(w, http.StatusInternalServerError, "Internal Error", httpx.CodeInternalError)
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		lh.logger.Info("not found", "listing_id", id)
		httpx.Error(w, http.StatusNotFound, "Not found", httpx.CodeNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (lh ListingHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateListingRequest
	ctx := r.Context()
	requestId := middleware.GetRequestIdFromContext(ctx)
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		lh.logger.Error("request parsing failed", "request_id", requestId, "error", err)
		httpx.Error(w, http.StatusBadRequest, "invalid body", httpx.CodeMalformedJSON)
		return
	}

	row := lh.db.QueryRowContext(ctx, `INSERT INTO listings (title,description,price,city) VALUES($1, $2, $3, $4) RETURNING id, title, created_at`,
		req.Title, req.Description, req.Price, req.City)

	var listing CreatedListingResponse
	err = row.Scan(&listing.Id, &listing.Title, &listing.CreatedAt)
	if err != nil {
		lh.logger.Error("failed to insert", "request_id", requestId, "error", err)
		httpx.Error(w, http.StatusInternalServerError, "Something went wrong", httpx.CodeInternalError)
		return
	}

	lh.logger.Info("listing created", "request_id", requestId, "listing_id", listing.Id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(listing)

}
