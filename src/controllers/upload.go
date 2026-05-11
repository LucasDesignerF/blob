package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"blob/src/database"
	"blob/src/functions"
	"blob/src/models"

	"github.com/asaskevich/govalidator"
	"github.com/google/uuid"
)

func UploadBlobController(w http.ResponseWriter, r *http.Request) {

	maxChunkSize := int64(32 << 20)
	if v := os.Getenv("BLOB_MAX_CHUNK_SIZE"); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil && parsed > 0 {
			maxChunkSize = parsed
		}
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxChunkSize)
	bucket := r.FormValue("bucket")
	filename := r.FormValue("filename")
	publicStr := r.FormValue("public")
	expiresAtStr := r.FormValue("expires_at")
	metadata := r.FormValue("metadata")

	validationErrors := make(map[string]string)

	if govalidator.IsNull(bucket) {
		validationErrors["bucket"] = "bucket is required"
	} else if !govalidator.StringLength(bucket, "1", "64") {
		validationErrors["bucket"] = "bucket must be 1-64 chars"
	} else if strings.Contains(bucket, "..") {
		validationErrors["bucket"] = "bucket cannot contain '..'"
	} else if !regexp.MustCompile(`^[a-zA-Z0-9/_-]+$`).MatchString(bucket) {
		validationErrors["bucket"] = "bucket can only contain letters, numbers, '/', '-', and '_'"
	}

	if filename != "" && !govalidator.StringLength(filename, "1", "255") {
		validationErrors["filename"] = "filename must be 1-255 chars"
	}

	if publicStr != "" && !functions.StringInSlice(publicStr, []string{"true", "false", "0", "1"}) {
		validationErrors["public"] = "public must be true, false, 0 or 1"
	}

	if expiresAtStr != "" && !govalidator.IsRFC3339(expiresAtStr) {
		validationErrors["expires_at"] = "expires_at must be RFC3339 date"
	}

	if metadata != "" && !govalidator.IsJSON(metadata) {
		validationErrors["metadata"] = "metadata must be valid JSON"
	}

	if len(validationErrors) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"error":  "Validation failed",
			"fields": validationErrors,
		}); err != nil {
			functions.Error("failed to encode validation error json: %v", err)
		}
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		functions.WriteJSONError(w, "File is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	allowedMimes := os.Getenv("BLOB_ALLOWED_MIME_TYPES")
	maxUploadSize, _ := strconv.ParseInt(os.Getenv("BLOB_MAX_UPLOAD_SIZE_BYTES"), 10, 64)
	maxStorageSize, _ := strconv.ParseInt(os.Getenv("BLOB_MAX_STORAGE_SIZE"), 10, 64)

	storagePath := os.Getenv("BLOB_STORAGE_PATH")
	if storagePath == "" {
		storagePath = "storage/uploads"
	}

	// Detect MIME type real lendo os primeiros 512 bytes
	var mime string
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		functions.WriteJSONError(w, "Failed to seek file", http.StatusInternalServerError)
		return
	}
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	mime = http.DetectContentType(buf[:n])
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		functions.WriteJSONError(w, "Failed to seek file", http.StatusInternalServerError)
		return
	}

	if !functions.IsAllowedMimeType(mime, functions.SplitComma(allowedMimes)) {
		functions.WriteJSONError(w, "MIME type not allowed", http.StatusBadRequest)
		return
	}

	if maxUploadSize > 0 && header.Size > maxUploadSize {
		functions.WriteJSONError(w, "File too large", http.StatusRequestEntityTooLarge)
		return
	}

	if maxStorageSize > 0 {
		total, err := functions.GetTotalStorageSize(storagePath)
		if err == nil && total+header.Size > maxStorageSize {
			functions.WriteJSONError(w, "Storage limit exceeded", http.StatusInsufficientStorage)
			return
		}
	}

	id := uuid.New()

	if filename == "" {
		filename = header.Filename
	}

	if filename != "" && (strings.Contains(filename, "..") || strings.ContainsAny(filename, "/\\")) {
		functions.WriteJSONError(w, "Invalid filename", http.StatusBadRequest)
		return
	}
	bucketPath := filepath.Join(storagePath, bucket)
	if err := os.MkdirAll(bucketPath, 0750); err != nil {
		functions.WriteJSONError(w, "Failed to create bucket directory", http.StatusInternalServerError)
		return
	}

	blobPath := filepath.Join(bucketPath, id.String())
	realBlobPath, err := filepath.Abs(blobPath)
	realBucketPath, err2 := filepath.Abs(bucketPath)
	if err != nil || err2 != nil || !strings.HasPrefix(realBlobPath, realBucketPath) {
		functions.WriteJSONError(w, "Invalid file path", http.StatusBadRequest)
		return
	}
	out, err := os.Create(realBlobPath)
	if err != nil {
		functions.WriteJSONError(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	size, err := io.Copy(out, file)
	if err != nil {
		functions.WriteJSONError(w, "Failed to write file", http.StatusInternalServerError)
		return
	}

	var metaJson map[string]interface{}
	if metadata != "" {
		if err := json.Unmarshal([]byte(metadata), &metaJson); err != nil {
			functions.Error("failed to unmarshal metadata: %v", err)
		}
	}

	var expiresAt *time.Time
	if expiresAtStr != "" {
		t, err := time.Parse(time.RFC3339, expiresAtStr)
		if err != nil {
			functions.WriteJSONError(w, "invalid expires_at format", http.StatusBadRequest)
			return
		}
		if t.Before(time.Now()) {
			functions.WriteJSONError(w, "expires_at cannot be in the past", http.StatusBadRequest)
			return
		}
		expiresAt = &t
	}

	public := true
	if publicStr != "" {
		if publicStr == "false" || publicStr == "0" {
			public = false
		}
	}

	blob := models.Blob{
		ID:        id,
		Bucket:    bucket,
		Filename:  filename,
		Mime:      mime,
		Size:      size,
		Hash:      id.String(),
		Path:      filepath.ToSlash(filepath.Join(bucket, id.String())),
		Public:    &public,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ExpiresAt: expiresAt,
	}

	if metaJson != nil {
		metaBytes, _ := json.Marshal(metaJson)
		blob.Metadata = metaBytes
	}

	if err := database.DB.Create(&blob).Error; err != nil {
		functions.WriteJSONError(w, "Failed to save blob metadata", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(blob); err != nil {
		functions.Error("failed to encode blob JSON: %v", err)
	}
}
