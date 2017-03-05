package main

import (
	"crypto/aes"
	"crypto/cipher"
    "encoding/base64"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

func getHeader(r *http.Request, header string) string{
    return r.Header[header][0]
}

func decryptReader(token string, reader io.Reader) io.Reader{
	key := []byte(token)
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	var iv [aes.BlockSize]byte
	stream := cipher.NewOFB(block, iv[:])
	return &cipher.StreamReader{S: stream, R: reader}
}

func encryptWriter(token string, writer io.Writer) io.Writer{
	key := []byte(token)
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	var iv [aes.BlockSize]byte
	stream := cipher.NewOFB(block, iv[:])
	return &cipher.StreamWriter{S: stream, W: writer}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dir := vars["dir"]
	filename := vars["filename"]
    encoded_filename:= base64.StdEncoding.EncodeToString([]byte(filename))
    path := filepath.Join(".", dir, encoded_filename)
	token := getHeader(r, "Token")

    // only accept long token
	if len(token) != 32 {
		http.Error(w, "Invalid Token", 400)
		return
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		http.Error(w, "directory not found", 404)
		return
	}
	defer f.Close()

    writer := encryptWriter(token, f)
	if _, err := io.Copy(writer, r.Body); err != nil {
		http.Error(w, "Error occurred copying to output stream", 500)
		return
	}
	fmt.Fprintf(w, "%s succesfully uploaded and encrypted\n", filename)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dir := vars["dir"]
	filename := vars["filename"]
    encoded_filename:= base64.StdEncoding.EncodeToString([]byte(filename))
    path := filepath.Join(".", dir, encoded_filename)
	token := getHeader(r, "Token")

	if len(token) != 32 {
		http.Error(w, "Invalid Token", 400)
		return
	}

	fi, err := os.Lstat(path)
	if err != nil {
		http.Error(w, "file not found", 404)
		return
	}

	f, err := os.Open(path)
	if err != nil {
		http.Error(w, "file not found", 404)
		return
	}
	defer f.Close()

	contentLength := uint64(fi.Size())
	contentType := mime.TypeByExtension(filepath.Ext(filename))
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.FormatUint(contentLength, 10))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Connection", "close")

    reader := decryptReader(token, f)
	if _, err = io.Copy(w, reader); err != nil {
		http.Error(w, "Error occurred copying to output stream", 500)
		return
	}
}

func downloadLatestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dir := vars["dir"]
	token := getHeader(r, "Token")

	if len(token) != 32 {
		http.Error(w, "Invalid Token", 400)
		return
	}

    files, _ := ioutil.ReadDir(filepath.Join("./", dir))
    var max_mod_time int64
    filename := ""

    for _, file := range files {
        fmt.Println(file.Name(), file.ModTime().Unix())
        if file.ModTime().Unix() > max_mod_time {
            filename = file.Name()
            max_mod_time = file.ModTime().Unix()
        }
    }

    decoded_filename, err := base64.StdEncoding.DecodeString(filename)
    if err != nil {
        http.Error(w, "error decoding filename", 500)
        return
    }

	f, err := os.Open(filepath.Join("./", dir, filename))
	if err != nil {
		http.Error(w, "file not found", 404)
		return
	}
	defer f.Close()

	fi, err := os.Lstat(filepath.Join("./", dir, filename))
	if err != nil {
		http.Error(w, "file not found", 404)
		return
	}

	contentLength := uint64(fi.Size())
	contentType := mime.TypeByExtension(filepath.Ext(string(decoded_filename)))
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.FormatUint(contentLength, 10))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", decoded_filename))
	w.Header().Set("Connection", "close")

    reader := decryptReader(token, f)
	if _, err = io.Copy(w, reader); err != nil {
		http.Error(w, "Error occurred copying to output stream", 500)
		return
	}
}
func main() {
	r := mux.NewRouter()
	r.HandleFunc("/{dir}/{filename}", uploadHandler).Methods("PUT")
	r.HandleFunc("/{dir}/latest", downloadLatestHandler).Methods("GET")
	r.HandleFunc("/{dir}/{filename}", downloadHandler).Methods("GET")
	err := http.ListenAndServe(":9090", r)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
