package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/thedevsaddam/renderer"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
)

var rnd *renderer.Render

func UploadFiles(res http.ResponseWriter, req *http.Request) {
	fmt.Println("Uploading new file")
	defer func() {
		fmt.Println("Finished uploading")
	}()

	// 1 << 20 = 10 MB
	if err := req.ParseMultipartForm(10 << 20); err != nil {
		log.Println(err)
		return
	}

	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, handler, err := req.FormFile("myFile")
	if err != nil {
		log.Println(err)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Println(err)
		}
	}()

	ext := filepath.Ext(handler.Filename)
	name := handler.Filename[:len(handler.Filename)-len(ext)]
	tempFile, err := ioutil.TempFile("./shared", name+"_*"+ext)
	if err != nil {
		log.Println(err)
	}
	defer func() {
		if err := tempFile.Close(); err != nil {
			log.Println(err)
		}
	}()

	// read all of the contents of our uploaded file into a byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println(err)
	}
	// write this byte array to our temporary file
	if nrBytes, err := tempFile.Write(fileBytes); err != nil {
		log.Println("Wrote", nrBytes, "bytes")
		log.Println(err)
		return
	}
	// return that we have successfully uploaded our file!
	if err := rnd.HTML(res, http.StatusOK, "go_to_upload", nil); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func Index(res http.ResponseWriter, req *http.Request) {
	if err := rnd.HTML(res, http.StatusOK, "main", nil); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func init() {
	opts := renderer.Options{
		ParseGlobPattern: "./*.html",
	}

	rnd = renderer.New(opts)
}

func main() {
	router := mux.NewRouter()

	port := flag.String("p", "8080", "port")
	directory := flag.String("d", "./shared", "static directory to host")
	flag.Parse()

	// Shared folder
	router.PathPrefix("/shared/").Handler(http.StripPrefix("/shared/", http.FileServer(http.Dir(*directory))))
	// Other
	router.HandleFunc("/", Index)
	router.HandleFunc("/upload", UploadFiles)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", *port), router))
}
