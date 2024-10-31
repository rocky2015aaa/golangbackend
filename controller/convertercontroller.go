package controller

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"golangbackend/service"

	"github.com/gorilla/mux"
)

type converterController struct {
	router  *mux.Router
	service service.ConverterService
}

type ConverterController interface {
	SetUpRouter()
}

type ConversionData struct {
	Width  string
	Height string
}

func NewConverterController(router *mux.Router, imgConverterService service.ConverterService) ConverterController {

	return &converterController{
		router:  router,
		service: imgConverterService,
	}
}

func (app *converterController) SetUpRouter() {
	app.router.
		Methods("POST").
		Path("/image/convert").
		HandlerFunc(app.convertImage)

	app.router.
		Methods("GET").
		Path("/image").
		HandlerFunc(app.getImage)
}

func (app *converterController) getRatio(r *http.Request) ConversionData {
	r.ParseForm()
	r.ParseMultipartForm(512 << 20)

	var formValue = make(map[string]interface{})
	for key, values := range r.Form {
		for _, value := range values {
			formValue[key] = value
		}
	}

	var ratioData ConversionData
	val, _ := json.Marshal(formValue)
	json.Unmarshal(val, &ratioData)
	return ratioData
}

func (app *converterController) convertImage(w http.ResponseWriter, r *http.Request) {
	fileName, mineType, validateErr := app.service.ValidateImage(r, "image")
	if validateErr != nil {
		app.responseError(w, validateErr.Error())
		return
	}

	//writeFileName := strings.Split(fileName, ".")
	//if (mineType != "text/plain; charset=utf-8" && mineType != "application/octet-stream") || (writeFileName[1] != "svg") {
	if (mineType != "text/plain; charset=utf-8" && mineType != "application/octet-stream") {
		app.responseError(w, "unable to convert: unsupportive image type")
		return
	}

	// getting ratio width and height from frontend, but currently, backend is not getting ratio input from frontend yet
	ratioData := app.getRatio(r)
	if ratioData.Width == "" {
		ratioData.Width = "2320" //adding default value
	}
	if ratioData.Height == "" {
		ratioData.Width = "910" //adding default value
	}
	width, err := strconv.Atoi(ratioData.Width)
	if err != nil {
		app.responseError(w, err.Error())
		return
	}
	height, err := strconv.Atoi(ratioData.Height)
	if err != nil {
		app.responseError(w, err.Error())
		return
	}

	writeFile, err := app.service.ConverSvgToEps(fileName, width, height)
	if err != nil {
		app.responseError(w, err.Error())
		return
	}

	app.responseFile(w, r, writeFile)
}

func (app *converterController) getImage(w http.ResponseWriter, r *http.Request) {
	u, _ := url.Parse(r.URL.String())
	vars := u.Query()
	if _, ok := vars["name"]; !ok {
		app.responseError(w, "file name is required")
		return
	}

	fileName := vars["name"][0]
	validateFile := strings.Split(fileName, ".")
	if len(validateFile) < 2 || validateFile[1] != "eps" {
		app.responseError(w, "invalid file request")
		return
	}

	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		app.responseError(w, "file does not exist")
		return
	}

	app.responseFile(w, r, fileName)
}

func (app *converterController) responseFile(w http.ResponseWriter, r *http.Request, writeFile string) {
	w.Header().Set("Content-Disposition", "attachment; filename="+writeFile)
	w.Header().Set("Content-Type", "image/x-eps")
	http.ServeFile(w, r, writeFile)
}

func (app *converterController) responseError(w http.ResponseWriter, message string) {
	payload := map[string]string{"error": message}
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		log.Println(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(422)
	w.Write(response)
}
