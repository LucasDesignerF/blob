package routes

import (
	"blob/src/controllers"
	multipart "blob/src/controllers/multipart"
	"blob/src/functions"
	"blob/src/middleware"
	"net/http"
	"strings"
)

func RegisterRoutes(mux *http.ServeMux, limiter *middleware.RateLimiter) {

	// helper para forçar método HTTP
	methodHandler := func(method string, h http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Method != method {
				functions.WriteJSONMethodNotAllowed(w)
				return
			}
			h.ServeHTTP(w, r)
		}
	}

	// GET / (public)
	mux.HandleFunc("/", GETHandler)

	// GET /health (public)
	mux.Handle(
		"/health",
		methodHandler("GET", limiter.Middleware(http.HandlerFunc(HealthHandler))),
	)

	// POST /blob/initiate (private)
	mux.Handle(
		"/blob/initiate",
		methodHandler("POST", limiter.Middleware(
			middleware.AuthMiddleware(
				http.HandlerFunc(multipart.InitiateUpload),
			),
		)),
	)

	// /blob (private) - supports GET (list) and PUT (upload)
	mux.Handle("/blob", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "PUT":
			limiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(controllers.UploadBlobController))).ServeHTTP(w, r)
			return
		case "GET":
			limiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(controllers.ListBlobsController))).ServeHTTP(w, r)
			return
		default:
			functions.WriteJSONMethodNotAllowed(w)
			return
		}
	}))

	// Unified handler for dynamic /blob/* routes (download, view, get/edit/delete, multipart)
	mux.HandleFunc("/blob/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// GET /blob/{id}/download (public)
		if strings.HasSuffix(path, "/download") || strings.HasSuffix(path, "/download/") {
			controllers.DownloadBlobController(w, r)
			return
		}

		// GET /blob/{id}/view (public)
		if strings.HasSuffix(path, "/view") || strings.HasSuffix(path, "/view/") {
			controllers.ViewBlobController(w, r)
			return
		}

		parts := strings.Split(strings.Trim(path, "/"), "/")
		if len(parts) < 2 || parts[0] != "blob" {
			functions.WriteJSONMethodNotAllowed(w)
			return
		}

		// multipart routes: /blob/{uploadId}/chunk, /blob/{uploadId}/complete, /blob/{uploadId}/status
		if len(parts) >= 3 {
			uploadAction := parts[2]
			switch uploadAction {
			case "chunk":
				if r.Method == "PUT" {
					limiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(multipart.UploadChunk))).ServeHTTP(w, r)
					return
				}
			case "complete":
				if r.Method == "POST" {
					limiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(multipart.CompleteUpload))).ServeHTTP(w, r)
					return
				}
			case "status":
				if r.Method == "GET" {
					limiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(multipart.UploadStatus))).ServeHTTP(w, r)
					return
				}
			}
		}

		// id-based operations: /blob/{id}
		if len(parts) >= 2 {
			if len(parts) == 2 || (len(parts) == 3 && parts[2] == "") {
				switch r.Method {
				case "GET":
					limiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(controllers.GetBlobController))).ServeHTTP(w, r)
					return
				case "POST":
					limiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(controllers.EditBlobController))).ServeHTTP(w, r)
					return
				case "DELETE":
					limiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(controllers.DeleteBlobController))).ServeHTTP(w, r)
					return
				}
			}
		}

		// fallback: method not allowed or not found
		functions.WriteJSONMethodNotAllowed(w)
	})

	// GET /metrics (private)
	mux.Handle(
		"/metrics",
		methodHandler("GET", limiter.Middleware(middleware.AuthMiddleware(http.HandlerFunc(controllers.BlobMetricsController)))),
	)

}
