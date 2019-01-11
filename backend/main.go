package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
)

var (
	addr        string
	storagePath string
)

func init() {
	port, has := os.LookupEnv("port")
	if !has {
		port = "8877"
	}
	addr = fmt.Sprintf(":%s", port)

	path, has := os.LookupEnv("STORAGE_PATH")
	if !has {
		path = "/tmp/audio"
	}
	storagePath = path
}

func main() {
	http.HandleFunc("/", handleRequest)
	log.Println(addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func handleRequest(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Content-Type", "application/json")
	if req.Method == http.MethodOptions {
		return
	}
	switch req.URL.Path {
	case "/ping":
		handlePing(w, req)
	case "/upload":
		handleUpload(w, req)
	case "/convert":
		handleConvert(w, req)
	case "/download":
		handleDownload(w, req)
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write(respError("api not found"))
	}
}

func handlePing(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("pong"))
}

func handleDownload(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write(respError("Method not allowed"))
		return
	}
	hash := req.URL.Query().Get("hash")
	if hash == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(respError("hash is required"))
		return
	}
	content, err := ioutil.ReadFile(fmt.Sprintf("%s/%s.m4r", storagePath, hash))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(respError(fmt.Sprintf("read file failed: %s", err)))
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename=ring.m4r")
	w.Header().Set("Content-Type", "audio/x-m4r")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
	w.Write(content)
}

func handleConvert(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write(respError("Method not allowed"))
		return
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(respError(fmt.Sprintf("read body failed: %s", err)))
		return
	}
	type Param struct {
		Hash     string  `json:"hash"`
		Start    float64 `json:"start"`    // seconds
		Duration float64 `json:"duration"` // seconds
		Fade     bool    `json:"fade"`
	}

	var param Param
	if err := json.Unmarshal(body, &param); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(respError(fmt.Sprintf("decode body failed: %s", err)))
		return
	}

	// rm m4r first
	if err := os.RemoveAll(fmt.Sprintf("%s/%s.m4r", storagePath, param.Hash)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(respError(fmt.Sprintf("delete file failed: %s", err)))
		return
	}

	// cmd
	cmdArgs := []string{
		// cut
		"-ss", fmt.Sprintf("%.2f", param.Start), "-t", fmt.Sprintf("%.2f", param.Duration),
		// input
		"-i", fmt.Sprintf("%s/%s.mp3", storagePath, param.Hash),
		// convert
		"-c:a", "libfdk_aac", "-c:v", "copy", "-f", "ipod", "-b:a", "96k",
	}
	if param.Fade {
		// append fade
		cmdArgs = append(cmdArgs, "-af", "afade=t=in:ss=0:d=1.5")
	}
	// append output
	cmdArgs = append(cmdArgs, fmt.Sprintf("%s/%s.m4r", storagePath, param.Hash))

	// exec cmd
	if _, err := exec.Command("ffmpeg", cmdArgs...).CombinedOutput(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(respError(fmt.Sprintf("convert failed")))
		return
	}
	w.Write(respSuccess(param))
}

func handleUpload(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write(respError("Method not allowed"))
		return
	}

	_, fileHeader, err := req.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(respError(fmt.Sprintf("upload file failed: %s", err)))
		return
	}
	if mime := fileHeader.Header.Get("Content-Type"); mime != "audio/mp3" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(respError(fmt.Sprintf("invalid file type: %s", mime)))
		return
	}

	// handle file
	f, err := fileHeader.Open()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(respError(fmt.Sprintf("open file failed: %s", err)))
		return
	}
	buf := bytes.NewBuffer(nil)
	buf.ReadFrom(f)
	hash := fmt.Sprintf("%x", md5.Sum(buf.Bytes()))
	// save
	if err := os.MkdirAll(storagePath, os.ModePerm); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(respError(fmt.Sprintf("mkdir failed: %s", err)))
		return
	}
	file, err := os.Create(fmt.Sprintf("/%s/%s.mp3", storagePath, hash))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(respError(fmt.Sprintf("create file failed: %s", err)))
		return
	}
	if _, err := file.Write(buf.Bytes()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(respError(fmt.Sprintf("write file failed: %s", err)))
		return
	}

	w.Write(respSuccess(map[string]interface{}{
		"hash": fmt.Sprintf("%s", hash),
	}))
}

func respError(err string) []byte {
	b, _ := json.Marshal(map[string]interface{}{
		"success": false,
		"error":   err,
	})
	return b
}

func respSuccess(data interface{}) []byte {
	b, _ := json.Marshal(map[string]interface{}{
		"success": true,
		"data":    data,
		"error":   nil,
	})
	return b
}
