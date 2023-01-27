package main

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const MAX_UPLOAD_SIZE = 5 * 1024 * 1024 // 5MB

func upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, MAX_UPLOAD_SIZE)
	if err := r.ParseMultipartForm(MAX_UPLOAD_SIZE); err != nil {
		http.Error(w, "file too big. must be under 5MB", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// make buffer used to detect content type
	buff := make([]byte, 512)
	if _, err := file.Read(buff); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filetype := http.DetectContentType(buff)
	if filetype != "image/jpeg" && filetype != "image/png" {
		http.Error(w, "invalid format. only jpeg and png are accepted", http.StatusBadRequest)
		return
	}

	// seek back to beginning to get whole file
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dst, err := os.Create(filepath.Join(".", "uploads", fileHeader.Filename))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
