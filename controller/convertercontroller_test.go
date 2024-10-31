package controller

import (
	"bytes"
	"golangbackend/service"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gorilla/mux"
)

func TestConvertImage(t *testing.T) {
	convertController := converterController{
		router:  mux.NewRouter().StrictSlash(true).PathPrefix("/api/v1").Subrouter(),
		service: service.NewConverterService(),
	}

	file, err := os.Open("../test_file.svg")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("image", "../test_file.svg")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(part, file); err != nil {
		t.Fatal(err)
	}
	writer.Close()

	req, err := http.NewRequest("GET", "/api/v1/convert", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()

	convertController.convertImage(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			rr.Code, http.StatusOK)
	}

	epsFilePath := "test_file.eps"

	if _, err := os.Stat(epsFilePath); os.IsNotExist(err) {
		t.Fatalf("EPS file %s does not exist", epsFilePath)
	}

	if ext := filepath.Ext(epsFilePath); ext != ".eps" {
		t.Fatalf("Invalid file extension for EPS file: %s", ext)
	}

	file, err = os.Open(epsFilePath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		t.Fatal(err)
	}
	if fileInfo.Size() == 0 {
		t.Fatal("EPS file is empty")
	}

	t.Logf("EPS file %s is valid", epsFilePath)

	os.Remove("test_file.eps")
	os.Remove("test_file.svg")
}
