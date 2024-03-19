package main

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func badrequest(w http.ResponseWriter, message string) {
	log.Println(message)
	http.Error(w, message, http.StatusBadRequest)
}

func failure(w http.ResponseWriter, message string) {
	log.Println(message)
	http.Error(w, message, http.StatusInternalServerError)
}

func forbiden(w http.ResponseWriter, message string) {
	log.Println(message)
	http.Error(w, message, http.StatusForbidden)
}

func notfound(w http.ResponseWriter, message string) {
	log.Println(message)
	http.Error(w, message, http.StatusNotFound)
}

func notallowed(w http.ResponseWriter, message string) {
	log.Println(message)
	http.Error(w, message, http.StatusMethodNotAllowed)
}

func getContentTypeForExt(s string) string {

	if strings.HasSuffix(s, "jpg") {
		return "image/jpeg"
	}

	if strings.HasSuffix(s, "jpeg") {
		return "image/jpeg"
	}

	if strings.HasSuffix(s, "gif") {
		return "image/gif"
	}

	if strings.HasSuffix(s, "png") {
		return "image/png"
	}

	if strings.HasSuffix(s, "webp") {
		return "image/webp"
	}

	if strings.HasSuffix(s, "avif") {
		return "image/avif"
	}

	return "application/octet-stream"
}

func proxyWorker(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		notallowed(w, "method not allowed")
		return
	}

	for i := 0; i < len(configuration.Locations); i++ {

		if configuration.Locations[i].prefix.MatchString(r.URL.Path) {
			// log.Println("Using configuration", i, configuration.Locations[i].Prefix)

			if configuration.Locations[i].DropParams && (len(r.URL.RawQuery) > 0) {
				http.Redirect(w, r, r.URL.Path, http.StatusMovedPermanently)
				return
			}

			URL := r.URL.Path

			if len(configuration.Locations[i].RegExpMatch) > 0 {
				URL = configuration.Locations[i].translation.ReplaceAllString(URL, configuration.Locations[i].RegExpSub)
			}

			// log.Println("Final URL:", URL)

			s3Client, err := minio.New(configuration.Locations[i].StorageEndpoint,
				&minio.Options{
					Creds:  credentials.NewStaticV4(configuration.Locations[i].StorageAccessKey, configuration.Locations[i].StorageSecretAccessKey, ""),
					Secure: configuration.Locations[i].StorageUseSSL})

			if err != nil {
				failure(w, "S3 client failed")
				return
			}

			object, err := s3Client.GetObject(context.Background(), configuration.Locations[i].StorageBucketName, URL, minio.GetObjectOptions{})
			if err != nil {
				failure(w, "S3 client failed")
				return
			}

			buffer := new(bytes.Buffer)
			buffer.ReadFrom(object)

			if len(buffer.Bytes()) == 0 {
				notfound(w, "empty object")
				return
			}

			w.Header().Set("Content-Type", getContentTypeForExt(URL))
			w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))

			if len(configuration.Locations[i].CacheControl) > 0 {
				w.Header().Set("Cache-Control", configuration.Locations[i].CacheControl)
			}

			if len(configuration.Locations[i].Expires) > 0 {
				t := time.Now().Add(configuration.Locations[i].expires)
				w.Header().Set("Expires", t.Format(time.RFC1123))
			}

			if _, err := w.Write(buffer.Bytes()); err != nil {
				log.Println(err)
				return
			}

			return
		}

	}

	http.Error(w, "Not found", http.StatusNotFound)
}
