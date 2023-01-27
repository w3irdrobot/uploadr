package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

const MAX_UPLOAD_SIZE = 5 * 1024 * 1024 // 5MB

type response struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

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
		logrus.WithError(err).Error("error getting file from request")
		http.Error(w, "file not found in request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// make buffer used to detect content type
	buff := make([]byte, 512)
	if _, err := file.Read(buff); err != nil {
		logrus.WithError(err).Error("error reading 512 bytes from file in request")
		http.Error(w, "unable to read file for mime detection", http.StatusInternalServerError)
		return
	}

	filetype := http.DetectContentType(buff)
	if filetype != "image/jpeg" && filetype != "image/png" {
		http.Error(w, "invalid format. only jpeg and png are accepted", http.StatusBadRequest)
		return
	}

	// seek back to beginning to get whole file
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		logrus.WithError(err).Error("error seeking to beging of file")
		http.Error(w, "unable to process the file", http.StatusInternalServerError)
		return
	}

	// check signature
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, file); err != nil {
		logrus.WithError(err).Error("error reading file from request into buffer")
		http.Error(w, "unable to read file", http.StatusInternalServerError)
		return
	}

	pubkey := r.FormValue("pubkey")
	signature := r.FormValue("signature")
	logrus.WithFields(logrus.Fields{"pubkey": pubkey, "signature": signature}).Debug("form values")

	validSig, err := checkSignature(pubkey, signature, buf.Bytes())
	if err != nil {
		logrus.WithError(err).Error("error checking the signature")
		http.Error(w, "error checking signature", http.StatusBadRequest)
		return
	}
	if !validSig {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	// save file to the filesystem
	dst, err := os.Create(filepath.Join(".", "uploads", fileHeader.Filename))
	if err != nil {
		logrus.WithError(err).Error("error opening the file on filesystem")
		http.Error(w, "error opening the file on filesystem", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err = io.Copy(dst, &buf); err != nil {
		logrus.WithError(err).Error("error copying the file to filesystem")
		http.Error(w, "error copying the file to filesystem", http.StatusInternalServerError)
		return
	}

	url := url.URL{
		Scheme: "https",
		Host:   "localhost:8080",
		Path:   fileHeader.Filename,
	}
	res := response{
		Name: fileHeader.Filename,
		Url:  url.String(),
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		logrus.WithError(err).Error("error encoding response")
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}
}
