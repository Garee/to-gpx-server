package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const (
	_          = iota
	GPX string = "gpx"
	TCX        = "gtrnctr"
)

const OUT_DIR = "tmp"

const MAX_UPLOAD_SIZE = 1024 * 1024 // 1MB

func getOutFile(inFile string, outDir string, outType string) string {
	_, file := path.Split(inFile)
	outName := strings.TrimSuffix(file, filepath.Ext(file))
	return outDir + "/" + outName + "." + outType
}

func execGpsBabel(inType string, inFile string, outType string, outFile string) *exec.Cmd {
	return exec.Command("gpsbabel", "-i", inType, "-f", inFile, "-o", outType, "-F", outFile)
}

func runGpsBabel(inType string, inFile string, outType string, outFile string) (bytes.Buffer, bytes.Buffer, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := execGpsBabel(inType, inFile, outType, outFile)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout, stderr, err
}

func convertFileToGpx(inType string, inFile string, outDir string) (string, error) {
	outType := GPX
	outFile := getOutFile(inFile, outDir, outType)

	_, stderr, err := runGpsBabel(inType, inFile, outType, outFile)
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return "", err
	}

	return outFile, err
}

func toGpx(inFile string, outDir string) (string, error) {
	inType := TCX
	return convertFileToGpx(inType, inFile, outDir)
}

func handleConvert(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "The POST HTTP method is required.", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	if fileHeader.Size > MAX_UPLOAD_SIZE {
		http.Error(w, "The uploaded file must not be greater than 1MB.", http.StatusBadRequest)
		return
	}

	buf := make([]byte, 512)
	_, err = file.Read(buf)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fileType := http.DetectContentType(buf)
	if !strings.Contains(fileType, "text/xml") {
		http.Error(w, "A text/xml file type is required.", http.StatusBadRequest)
		return
	}

	// Seek to the first "<" character to handle files that start with whitespace.
	offset := bytes.Index(buf, []byte("<"))
	_, err = file.Seek(int64(offset), io.SeekStart)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = os.MkdirAll(OUT_DIR, os.ModePerm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmp, err := os.MkdirTemp(OUT_DIR, fmt.Sprintf("%d", time.Now().UnixNano()))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tmp)

	dst, err := os.Create(filepath.Join(tmp, fileHeader.Filename))
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

	converted, err := toGpx(dst.Name(), tmp)
	if err != nil {
		http.Error(w, "The conversion was unsuccesful. Is the file type supported?", http.StatusInternalServerError)
		return
	}

	http.ServeFile(w, r, converted)
}

func initRoutes() {
	http.HandleFunc("/convert", handleConvert)
}

func startServer(port int) {
	err := http.ListenAndServe(":"+fmt.Sprint(port), nil)
	if err != nil {
		fmt.Println("Failed to start HTTP server: ", err)
		os.Exit(1)
	}
}

func main() {
	initRoutes()
	startServer(8080)
}
