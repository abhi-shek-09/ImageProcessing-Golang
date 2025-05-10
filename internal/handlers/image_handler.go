package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"image"
	"image-processing/internal/database"
	"image-processing/internal/middleware"
	"image-processing/internal/models"
	"image-processing/internal/service"
	"io"
	"log"
	"strings"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func UploadImage(w http.ResponseWriter, r *http.Request) {
	// verify auth
	userID := r.Context().Value(middleware.UserIDKey)
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	uID, ok := userID.(int)
	if !ok {
		http.Error(w, "Invalid User ID", http.StatusUnauthorized)
		return
	}

	// get the file content and header content from here
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Image file required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// generate a filename
	safeFilename := filepath.Base(header.Filename)
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), safeFilename)
	uploadPath := filepath.Join("uploads", filename)

	// make the folder
	if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
		http.Error(w, "Unable to create directory", http.StatusInternalServerError)
		return
	}

	// create the output file
	out, err := os.Create(uploadPath)
	if err != nil {
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	// copy file contents into output file
	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	query := "INSERT INTO images (user_id, original_url, transformed_url, metadata) VALUES ($1, $2, $3, $4) RETURNING id, created_at"
	originalURL := fmt.Sprintf("/uploads/%s", filename)
	metadata := map[string]interface{}{
		"filename": filename,
		"size":     header.Size,
		"type":     header.Header.Get("Content-Type"),
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		http.Error(w, "Error encoding metadata", http.StatusInternalServerError)
		return
	}

	var image models.Images
	image.UserID = uID
	image.OriginalURL = originalURL
	image.TransformedURL = ""
	image.Metadata = metadataJSON

	err = database.DB.QueryRow(query, image.UserID, image.OriginalURL, image.TransformedURL, image.Metadata).Scan(&image.ID, &image.CreatedAt)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(image)
}

func GetImageByID(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey)
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	uID, ok := userID.(int)
	if !ok {
		http.Error(w, "Invalid User ID", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid image ID", http.StatusBadRequest)
		return
	}

	var image models.Images
	var metadataJSON []byte

	query := `SELECT id, user_id, original_url, transformed_url, metadata, created_at FROM images WHERE id = $1`
	err = database.DB.QueryRow(query, id).Scan(
		&image.ID,
		&image.UserID,
		&image.OriginalURL,
		&image.TransformedURL,
		&metadataJSON,
		&image.CreatedAt,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("UploadImage DB error: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if uID != image.UserID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = json.Unmarshal(metadataJSON, &image.Metadata)
	if err != nil {
		http.Error(w, "Error decoding metadata", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(image)
}

func GetImages(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey)
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	uID, ok := userID.(int)
	if !ok {
		http.Error(w, "Invalid User ID", http.StatusUnauthorized)
		return
	}

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1 // Default to page 1
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit < 1 {
		limit = 10 // Default to 10 items per page
	}

	offset := (page - 1) * limit // determines where to start fetching records
	rows, err := database.DB.Query("SELECT id, user_id, original_url, transformed_url, metadata, created_at FROM images WHERE user_id = $1 LIMIT $2 OFFSET $3", uID, limit, offset)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var images []models.Images
	for rows.Next() {
		var image models.Images
		var metadataJSON []byte

		if err := rows.Scan(&image.ID, &image.UserID, &image.OriginalURL, &image.TransformedURL, &metadataJSON, &image.CreatedAt); err != nil {
			http.Error(w, "Error scanning results", http.StatusInternalServerError)
			return
		}

		if err := json.Unmarshal(metadataJSON, &image.Metadata); err != nil {
			http.Error(w, "Error decoding metadata", http.StatusInternalServerError)
			return
		}

		images = append(images, image)
	}

	var total int
	err = database.DB.QueryRow("SELECT COUNT(*) FROM images WHERE user_id=$1", uID).Scan(&total)
	if err != nil {
		http.Error(w, "Failed to get total count", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":  images,
		"page":  page,
		"limit": limit,
		"total": total,
	})
}

func TransformImage(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey)
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	uID, ok := userID.(int)
	if !ok {
		http.Error(w, "Invalid User ID", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid image ID", http.StatusBadRequest)
		return
	}

	var img models.Images
	var metadataJSON []byte

	query := `SELECT id, user_id, original_url, transformed_url, metadata, created_at FROM images WHERE id = $1`
	err = database.DB.QueryRow(query, id).Scan(
		&img.ID,
		&img.UserID,
		&img.OriginalURL,
		&img.TransformedURL,
		&metadataJSON,
		&img.CreatedAt,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if uID != img.UserID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = json.Unmarshal(metadataJSON, &img.Metadata)
	if err != nil {
		http.Error(w, "Error decoding metadata", http.StatusInternalServerError)
		return
	}

	var transformReq models.TransformationRequest
	if err := json.NewDecoder(r.Body).Decode(&transformReq); err != nil {
		http.Error(w, "Invalid transformation request", http.StatusBadRequest)
		return
	}

	log.Println(img.OriginalURL)
	originalPath := filepath.Join(".", img.OriginalURL) // resolves to ./uploads/174...
	file, err := os.Open(originalPath)
	if err != nil {
		http.Error(w, "Failed to open original image", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Step 2: Decode image
	srcImg, _, err := image.Decode(file)
	if err != nil {
		http.Error(w, "Failed to decode image", http.StatusInternalServerError)
		return
	}

	// Step 3: Apply transformations and save new image
	newFullPath, err := service.TransformAndSaveImage(srcImg, &transformReq, originalPath)
	if err != nil {
		http.Error(w, "Transformation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	transformedURL := strings.TrimPrefix(newFullPath, ".") // remove the leading dot
	transformedURL = filepath.ToSlash(transformedURL) 
	updateQuery := `UPDATE images SET transformed_url = $1 WHERE id = $2`
	
	_, err = database.DB.Exec(updateQuery, transformedURL, id)
	if err != nil {
		http.Error(w, "Failed to update image path in DB", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message":         "Image transformed successfully",
		"transformed_url": newFullPath,
	})
}
