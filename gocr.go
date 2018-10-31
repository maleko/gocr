/*
 * PDF to text: Extract all text for each page of a pdf file.
 *
 * N.B. Only outputs character codes as seen in the content stream.  Need to account for text encoding to get readable
 * text in many cases.
 *
 * Run as: go run pdf_extract_text.go input.pdf
 */

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	pdfcontent "github.com/unidoc/unidoc/pdf/contentstream"
	pdf "github.com/unidoc/unidoc/pdf/model"
)

// Dictionaries Struct which contains
// the base folder and an array of matches
type Dictionary struct {
	BaseFolder string  `json:"base_folder"`
	Matches    []Match `json:"matches"`
}

// Match struct which contains
// a short_word type and
// a folder type
type Match struct {
	ShortWord string `json:"short_word"`
	Folder    string `json:"folder"`
}

func main() {
	jsonFileName := "dictionary.json"
	jsonFile, err := os.Open(jsonFileName)

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var dictionary Dictionary
	json.Unmarshal(byteValue, &dictionary)

	baseFolder := dictionary.BaseFolder
	unprocessedFolder := baseFolder + "Unprocessed/"

	for i := 0; i < len(dictionary.Matches); i++ {

		shortWord := dictionary.Matches[i].ShortWord
		destinationFolder := dictionary.Matches[i].Folder
		fmt.Println("Short Word: " + shortWord)
		fmt.Println("Folder: " + destinationFolder)

		files, err := ioutil.ReadDir(unprocessedFolder)
		if err != nil {
			log.Fatal(err)
		}

		for _, file := range files {
			if strings.Contains(file.Name(), "pdf") {
				currentFile := unprocessedFolder + file.Name()
				found, err := detectString(currentFile, shortWord)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
					break
				}

				if found {
					destinationFile := baseFolder + destinationFolder + file.Name()
					fmt.Println(destinationFile)

					os.Rename(currentFile, destinationFile)
				}
			}
		}

		fmt.Println("===============================================")
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func detectString(inputPath string, searchText string) (bool, error) {
	f, err := os.Open(inputPath)
	if err != nil {
		return false, err
	}

	defer f.Close()

	pdfReader, err := pdf.NewPdfReader(f)
	if err != nil {
		return false, err
	}

	page, err := pdfReader.GetPage(1)
	if err != nil {
		return false, err
	}

	found, err := locateString(page, searchText)
	if err != nil {
		return false, err
	}

	return found, err
}

func locateString(page *pdf.PdfPage, searchText string) (bool, error) {
	found := false
	pageContentStr := ""

	contentStreams, err := page.GetContentStreams()
	if err != nil {
		return false, err
	}

	for _, cstream := range contentStreams {
		pageContentStr += cstream
	}
	if err != nil {
		return false, err
	}

	cstreamParser := pdfcontent.NewContentStreamParser(pageContentStr)
	txt, err := cstreamParser.ExtractText()
	if err != nil {
		return false, err
	}

	if strings.Contains(txt, searchText) {
		found = true
	}

	return found, nil
}
