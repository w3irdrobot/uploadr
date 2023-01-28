package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
	shasum := sha256.Sum256(buf.Bytes())

	pubkey := r.FormValue("pubkey")
	signature := r.FormValue("signature")
	logrus.WithFields(logrus.Fields{"pubkey": pubkey, "signature": signature}).Debug("form values")

	validSig, err := checkSignature(pubkey, signature, shasum[:])
	if err != nil {
		logrus.WithError(err).Error("error checking the signature")
		http.Error(w, "error checking signature", http.StatusBadRequest)
		return
	}
	if !validSig {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	// create directory path to put file
	directory := createAndGetDirectory(shasum[:])
	fullDirectory := filepath.Join("uploads", directory)
	if err := os.MkdirAll(fullDirectory, os.ModePerm); err != nil {
		logrus.WithError(err).Error("error creating the directory on filesystem")
		http.Error(w, "error preparing the file", http.StatusInternalServerError)
		return
	}

	// get name and paths to file
	extension := filepath.Ext(fileHeader.Filename)
	name := fmt.Sprintf("%s%s", hex.EncodeToString(shasum[:]), extension)
	path := filepath.Join(directory, name)
	fullPath, err := filepath.Abs(filepath.Join(fullDirectory, name))
	if err != nil {
		logrus.WithError(err).Error("error calculating the local path on filesystem")
		http.Error(w, "error opening the file on filesystem", http.StatusInternalServerError)
		return
	}

	// save file to the filesystem
	dst, err := os.Create(fullPath)
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

	// respond with file data
	url := url.URL{
		Scheme: "http",
		Host:   "localhost:8080",
		Path:   path,
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
