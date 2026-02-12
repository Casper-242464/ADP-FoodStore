package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"foodstore/internal/models"
	"foodstore/internal/services"
)

type ProductHandler struct {
	service     *services.ProductService
	userService *services.UserService
	uploadDir   string
}

const (
	maxUploadFileBytes = int64(8 << 20) // 8MB
	maxFormMemoryBytes = int64(16 << 20)
	defaultUploadDir   = "frontend/uploads"
)

func NewProductHandler(ps *services.ProductService, us *services.UserService, uploadDir string) *ProductHandler {
	dir := strings.TrimSpace(uploadDir)
	if dir == "" {
		dir = defaultUploadDir
	}
	return &ProductHandler{service: ps, userService: us, uploadDir: dir}
}

func (ph *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		var (
			products []models.Product
			err      error
		)

		if r.URL.Query().Get("mine") == "1" {
			userID, userErr := getUserIDFromHeader(r)
			if userErr != nil {
				writeJSONError(w, http.StatusUnauthorized, "missing or invalid user id")
				return
			}
			products, err = ph.service.ListProductsBySellerID(userID)
		} else {
			products, err = ph.service.ListProducts()
		}
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, products)
	case http.MethodPost:
		userID, err := getUserIDFromHeader(r)
		if err != nil {
			writeJSONError(w, http.StatusUnauthorized, "missing or invalid user id")
			return
		}

		reqBody, err := parseProductMultipart(r, true, ph.uploadDir)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}

		if reqBody.Name == "" || reqBody.Description == "" || reqBody.Category == "" {
			writeJSONError(w, http.StatusBadRequest, "name, description, category are required")
			return
		}
		if reqBody.Price < 0 || reqBody.Stock < 0 {
			writeJSONError(w, http.StatusBadRequest, "price and stock must be >= 0")
			return
		}

		id, err := ph.service.CreateProduct(models.Product{
			Name:        reqBody.Name,
			Description: reqBody.Description,
			ImageURL:    reqBody.ImageURL,
			Price:       reqBody.Price,
			Stock:       reqBody.Stock,
			Category:    reqBody.Category,
			Unit:        reqBody.Unit,
		}, userID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"id":        id,
			"image_url": reqBody.ImageURL,
		})
	case http.MethodPut:
		userID, err := getUserIDFromHeader(r)
		if err != nil {
			writeJSONError(w, http.StatusUnauthorized, "missing or invalid user id")
			return
		}
		isAdmin, err := ph.isAdministrator(userID)
		if err != nil {
			writeJSONError(w, http.StatusUnauthorized, "user not found")
			return
		}

		reqBody, err := parseProductMultipart(r, false, ph.uploadDir)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		if reqBody.ID <= 0 {
			writeJSONError(w, http.StatusBadRequest, "id is required")
			return
		}
		if reqBody.Name == "" || reqBody.Description == "" || reqBody.Category == "" {
			writeJSONError(w, http.StatusBadRequest, "name, description, category are required")
			return
		}
		if reqBody.Price < 0 || reqBody.Stock < 0 {
			writeJSONError(w, http.StatusBadRequest, "price and stock must be >= 0")
			return
		}

		existing, err := ph.service.GetProductByID(reqBody.ID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeJSONError(w, http.StatusNotFound, "product not found")
				return
			}
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if !isAdmin && existing.SellerID != userID {
			writeJSONError(w, http.StatusForbidden, "you can edit only your own products")
			return
		}

		imageURL := existing.ImageURL
		if reqBody.HasImage {
			imageURL = reqBody.ImageURL
		}

		productToUpdate := models.Product{
			ID:          reqBody.ID,
			Name:        reqBody.Name,
			Description: reqBody.Description,
			ImageURL:    imageURL,
			Price:       reqBody.Price,
			Stock:       reqBody.Stock,
			Category:    reqBody.Category,
			Unit:        reqBody.Unit,
		}
		var updated bool
		if isAdmin {
			updated, err = ph.service.UpdateProductAsAdmin(productToUpdate)
		} else {
			updated, err = ph.service.UpdateProduct(productToUpdate, userID)
		}
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if !updated {
			writeJSONError(w, http.StatusNotFound, "product not found")
			return
		}

		if reqBody.HasImage {
			deleteLocalProductImage(existing.ImageURL, ph.uploadDir)
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
	case http.MethodDelete:
		userID, err := getUserIDFromHeader(r)
		if err != nil {
			writeJSONError(w, http.StatusUnauthorized, "missing or invalid user id")
			return
		}
		isAdmin, err := ph.isAdministrator(userID)
		if err != nil {
			writeJSONError(w, http.StatusUnauthorized, "user not found")
			return
		}

		idStr := r.URL.Query().Get("id")
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid id")
			return
		}

		existing, err := ph.service.GetProductByID(id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeJSONError(w, http.StatusNotFound, "product not found")
				return
			}
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if !isAdmin && existing.SellerID != userID {
			writeJSONError(w, http.StatusForbidden, "you can delete only your own products")
			return
		}

		var deleted bool
		if isAdmin {
			deleted, err = ph.service.DeleteProductAsAdmin(id)
		} else {
			deleted, err = ph.service.DeleteProduct(id, userID)
		}
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if !deleted {
			writeJSONError(w, http.StatusNotFound, "product not found")
			return
		}

		deleteLocalProductImage(existing.ImageURL, ph.uploadDir)
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (ph *ProductHandler) isAdministrator(userID int) (bool, error) {
	user, err := ph.userService.GetUserByID(userID)
	if err != nil {
		return false, err
	}
	return user.Role == "administrator", nil
}

type productMultipartRequest struct {
	ID          int
	Name        string
	Description string
	ImageURL    string
	Price       float64
	Stock       int
	Category    string
	Unit        string
	HasImage    bool
}

func parseProductMultipart(r *http.Request, imageRequired bool, uploadDir string) (productMultipartRequest, error) {
	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		return productMultipartRequest{}, errors.New("use multipart/form-data")
	}

	if err := r.ParseMultipartForm(maxFormMemoryBytes); err != nil {
		return productMultipartRequest{}, errors.New("invalid multipart form")
	}

	price, err := strconv.ParseFloat(strings.TrimSpace(r.FormValue("price")), 64)
	if err != nil {
		return productMultipartRequest{}, errors.New("invalid price")
	}
	stock, err := strconv.Atoi(strings.TrimSpace(r.FormValue("stock")))
	if err != nil {
		return productMultipartRequest{}, errors.New("invalid stock")
	}
	unit, err := normalizeProductUnit(r.FormValue("unit"))
	if err != nil {
		return productMultipartRequest{}, err
	}

	id := 0
	if idStr := strings.TrimSpace(r.FormValue("id")); idStr != "" {
		parsedID, parseErr := strconv.Atoi(idStr)
		if parseErr != nil {
			return productMultipartRequest{}, errors.New("invalid id")
		}
		id = parsedID
	}

	imageURL, hasImage, err := saveUploadedImage(r, "image", imageRequired, uploadDir)
	if err != nil {
		return productMultipartRequest{}, err
	}

	return productMultipartRequest{
		ID:          id,
		Name:        strings.TrimSpace(r.FormValue("name")),
		Description: strings.TrimSpace(r.FormValue("description")),
		ImageURL:    imageURL,
		Price:       price,
		Stock:       stock,
		Category:    strings.TrimSpace(r.FormValue("category")),
		Unit:        unit,
		HasImage:    hasImage,
	}, nil
}

func normalizeProductUnit(raw string) (string, error) {
	v := strings.TrimSpace(strings.ToLower(raw))
	switch v {
	case "kg":
		return "kg", nil
	case "piece", "pieces", "pcs", "pc", "shtuk", "sht", "штук", "шт":
		return "piece", nil
	case "pack", "pachka", "пачка":
		return "pack", nil
	case "":
		return "piece", nil
	default:
		return "", errors.New("invalid unit (use: kg, piece, pack)")
	}
}

func saveUploadedImage(r *http.Request, fieldName string, required bool, uploadDir string) (string, bool, error) {
	file, header, err := r.FormFile(fieldName)
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			if required {
				return "", false, errors.New("image file is required")
			}
			return "", false, nil
		}
		return "", false, errors.New("failed to read uploaded file")
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
	default:
		return "", false, errors.New("allowed image formats: .jpg, .jpeg, .png, .gif, .webp")
	}

	dir := strings.TrimSpace(uploadDir)
	if dir == "" {
		dir = defaultUploadDir
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", false, errors.New("failed to prepare upload directory")
	}

	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	localPath := filepath.Join(dir, filename)
	dst, err := os.Create(localPath)
	if err != nil {
		return "", false, errors.New("failed to save image file")
	}
	defer dst.Close()

	written, err := io.Copy(dst, io.LimitReader(file, maxUploadFileBytes+1))
	if err != nil {
		return "", false, errors.New("failed to write image file")
	}
	if written > maxUploadFileBytes {
		_ = os.Remove(localPath)
		return "", false, errors.New("image file is too large (max 8MB)")
	}

	return "/uploads/" + filename, true, nil
}

func deleteLocalProductImage(imageURL string, uploadDir string) {
	if !strings.HasPrefix(imageURL, "/uploads/") {
		return
	}
	baseName := filepath.Base(imageURL)
	if baseName == "." || baseName == "/" || baseName == "" {
		return
	}
	dir := strings.TrimSpace(uploadDir)
	if dir == "" {
		dir = defaultUploadDir
	}
	_ = os.Remove(filepath.Join(dir, baseName))
}

func getUserIDFromHeader(r *http.Request) (int, error) {
	userIDStr := r.Header.Get("X-User-Id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		return 0, errors.New("invalid user id")
	}
	return userID, nil
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
