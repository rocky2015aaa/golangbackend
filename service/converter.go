package service

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

//const ImageScale = 0.25

type converterService struct {
}

type ConverterService interface {
	ValidateImage(request *http.Request, fileName string) (string, string, error)
	ConverSvgToEps(fileName string, height, width int) (string, error)
}

func NewConverterService() ConverterService {
	return &converterService{}
}

func (process *converterService) ValidateImage(request *http.Request, fileName string) (string, string, error) {
	request.ParseForm()
	request.ParseMultipartForm(512 << 20)
	file, reqFilename, err := request.FormFile(fileName)
	if err != nil {
		return "", "", err
	}

	fileHeader := make([]byte, 512)
	if _, err := file.Read(fileHeader); err != nil {
		return "", "", err
	}
	mimeType := http.DetectContentType(fileHeader)

	//writeFileName := strings.Split(reqFilename.Filename, ".")
	//if (mimeType != "text/plain; charset=utf-8" && mimeType != "application/octet-stream") || (writeFileName[1] != "svg") {
	if (mimeType != "text/plain; charset=utf-8" && mimeType != "application/octet-stream") {
		return "", "", errors.New("unable to convert: unsupportive image type")
	}

	files := request.MultipartForm.File[fileName]
	fileErr := process.processFiles(reqFilename.Filename, files)

	return reqFilename.Filename, mimeType, fileErr
}

func (process *converterService) processFiles(fileName string, files []*multipart.FileHeader) error {
	file, err := files[0].Open()
	if err != nil {
		return err
	}
	defer file.Close()

	uploadedFiles, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	return os.WriteFile(fileName, uploadedFiles, fs.ModePerm)
}

func (proces *converterService) ConverSvgToEps(fileName string, width, height int) (string, error) {
	writeFileName := strings.Split(fileName, ".")
	writeTo := fmt.Sprintf("%s%s", writeFileName[0], ".eps")

	ratio := fmt.Sprintf("%fx%f!", float64(width)*2.628, float64(height)*2.628) //added to adjust ratio
	fmt.Println(ratio)                            //as for now it is not getting ratio input from frontend yet

	conversionCmd := exec.Command("inkscape", fileName, "--export-type", "eps", "-o", writeTo)
	if err := conversionCmd.Run(); err != nil {
		fmt.Println(err)		
    		return "", err
	}

	resize := exec.Command("convert", writeTo, "-resize", ratio, writeTo)
	if _, err := resize.Output(); err != nil {
		fmt.Println(err)		
		return "", err
	}

	return writeTo, nil
}
