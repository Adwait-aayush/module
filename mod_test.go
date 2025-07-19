package module

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestMod_GenRandomString(t *testing.T) {
	var testmod Module

	s := testmod.GenRandomString(10)
	if len(s) != 10 {
		t.Error("Wrong length of random string returned")
	}
}

var test = []struct {
	name         string
	allowedTypes []string
	rename       bool
	error        bool
}{
	{"Test1", []string{"application/pdf", "image/png"}, false, false},
	{"Test2", []string{"application/pdf", "image/png"}, true, false},
	{"Test1", []string{"image/png"}, false, true},
}

func TestMod_UploadFiles(t *testing.T) {
	for _, e := range test {
		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)
		wg := sync.WaitGroup{}
		wg.Add(1)

		go func() {
			defer writer.Close()
			defer wg.Done()

			part, err := writer.CreateFormFile("file", "5bftXbBRDRD7RH3V5HD9HhLdL.pdf")
			if err != nil {
				t.Error("Error creating form file:", err)
				return
			}

			f, err := os.Open("./testdata/5bftXbBRDRD7RH3V5HD9HhLdL.pdf")
			if err != nil {
				t.Error("Error opening test PDF file:", err)
				return
			}
			defer f.Close()

			// Just copy the file to the form part instead of decoding/encoding PDF
			_, err = io.Copy(part, f)
			if err != nil {
				t.Error("Error copying PDF content to part:", err)
				return
			}
		}()

		request := httptest.NewRequest("POST", "/", pr)
		request.Header.Add("Content-Type", writer.FormDataContentType())

		var testmod Module
		testmod.AllowedFileTypes = e.allowedTypes
		testmod.MaxFileSize = 10 * 1024 * 1024 // Ensure max file size is set

		uploadedFiles, err := testmod.UploadFiles(request, "./testdata/uploads/", e.rename)

		// Check expected vs actual error
		if err != nil && !e.error {
			t.Error("Unexpected error:", err)
		}
		if err == nil && e.error {
			t.Error("Expected error but got none:", e.name)
		}

		if err == nil && !e.error {
			uploadedFilePath := filepath.Join("./testdata/uploads/", uploadedFiles[0].NewFileName)
			if _, statErr := os.Stat(uploadedFilePath); os.IsNotExist(statErr) {
				t.Error("File not uploaded")
			} else {
				_ = os.Remove(uploadedFilePath)
			}
		}

		wg.Wait()
	}
}

// var uploadOneTests = []struct {
// 	name          string
// 	uploadDir     string
// 	errorExpected bool
// }{
// 	{name: "valid", uploadDir: "./testdata/uploads/", errorExpected: false},
// 	{name: "invalid", uploadDir: "//", errorExpected: true},
// }

func TestTools_UploadOneFile(t *testing.T) {
	var uploadOneTests = []struct {
		name          string
		uploadDir     string
		errorExpected bool
	}{
		{name: "valid", uploadDir: "./testdata/uploads/", errorExpected: false},
		{name: "invalid", uploadDir: "//", errorExpected: true},
	}

	for _, e := range uploadOneTests {
		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)

		wg := sync.WaitGroup{}
		wg.Add(1)

		go func() {
			defer writer.Close()
			defer wg.Done()

			// create a form file with field name "file" and file name "sample.pdf"
			part, err := writer.CreateFormFile("file", "5bftXbBRDRD7RH3V5HD9HhLdL.pdf")
			if err != nil {
				t.Error("error creating form file:", err)
				return
			}

			// open the sample test PDF file
			f, err := os.Open("./testdata/5bftXbBRDRD7RH3V5HD9HhLdL.pdf")
			if err != nil {
				t.Error("error opening PDF:", err)
				return
			}
			defer f.Close()

			// copy the PDF content to the form field
			_, err = io.Copy(part, f)
			if err != nil {
				t.Error("error copying PDF to part:", err)
				return
			}
		}()

		request := httptest.NewRequest("POST", "/", pr)
		request.Header.Add("Content-Type", writer.FormDataContentType())

		var testTools Module
		testTools.AllowedFileTypes = []string{"application/pdf"}
		testTools.MaxFileSize = 5 * 1024 * 1024

		uploadedFile, err := testTools.UploadFile(request, e.uploadDir, true)

		if e.errorExpected && err == nil {
			t.Errorf("%s: error expected, but none received", e.name)
		}

		if !e.errorExpected {
			if err != nil {
				t.Errorf("%s: unexpected error: %v", e.name, err)
				continue
			}

			// check if uploaded file exists
			filePath := filepath.Join(e.uploadDir, uploadedFile.NewFileName)
			if _, statErr := os.Stat(filePath); os.IsNotExist(statErr) {
				t.Errorf("%s: expected uploaded file to exist: %v", e.name, statErr)
			} else {
				// clean up uploaded file
				_ = os.Remove(filePath)
			}
		}

		wg.Wait()
	}
}

func TestMod_CreateDirIfNotExist(t *testing.T) {
	var mod Module
	err := mod.CreateDirIfNotExist("./testdata/mydir")
	if err != nil {
		t.Error(err)
	}

	_ = os.Remove("./testdata/mydir")
}

var slugTest = []struct {
	name     string
	n        string
	expected string
	error    bool
}{
	{"ValidSlug", "Hello World! This is a test", "hello-world-this-is-a-test", false},
	{"InvalidSlug", "Hello World! This is a test", "", true},
}

func TestMod_TestSlug(t *testing.T) {
	var mod Module
	for _, e := range slugTest {
		slug, err := mod.MakeSlug(e.n)
		if err != nil && !e.error {
			t.Errorf("expected error%s for but did not recieve %s", e.name, err.Error())
		}
		if !e.error && slug != e.expected {
			t.Errorf("expected slug %s but got %s", e.expected, slug)
		}

	}
}

var jsonTest = []struct {
	name         string
	json         string
	error        bool
	maxsize      int64
	allowUnknown bool
}{
	{"ValidJson", `{"name": "test", "value": 123}`, false, 1024, false},
	{"InvalidJson", `{"name": "test", "value":}`, true, 1024, false},
	{"incorrectType", `{"name": "test", "value": "not a number"}`, true, 1024, false},
	{"EmptyJson", ``, true, 1024, false},
}

func TestMod_ReadJson(t *testing.T) {
	var mod Module
	for _, e := range jsonTest {
		mod.MaxFileSize = e.maxsize
		mod.AllowUnknownFileTypes = e.allowUnknown

		var decodejson struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		}
		req, err := http.NewRequest("POST", "/", bytes.NewReader([]byte(e.json)))
		if err != nil {
			t.Log("Error creating request:", err)
		}
		rr := httptest.NewRecorder()
		err = mod.ReadJSON(rr, req, &decodejson)
		if e.error && err == nil {
			t.Errorf("expected error for %s but did not receive one", e.name)
		}
		if !e.error && err != nil {
			t.Errorf("did not expect error for %s but received: %s", e.name, err.Error())
		}
		req.Body.Close()
	}
}
func TestMod_WriteJson(t *testing.T) {
	var mod Module
	rr := httptest.NewRecorder()
	response := JSONResponse{
		Error:   false,
		Message: "Success",
	}
	headers := make(http.Header)
	headers.Add("Content-Type", "application/json")
	err := mod.WriteJSON(rr, http.StatusOK, response, headers)
	if err != nil {
		t.Errorf("Failed to write Json %v", err)
	}
}

func TestMod_ErrorJson(t *testing.T) {
	var mod Module
	rr := httptest.NewRecorder()
	err := mod.ErrorJSON(rr, errors.New("error"), http.StatusInternalServerError)
	if err != nil {
		t.Errorf("Failed to write error Json %v", err)
	}
	var payload JSONResponse
	decoder := json.NewDecoder(rr.Body)
	if err := decoder.Decode(&payload); err != nil {
		t.Errorf("Failed to decode error Json %v", err)
	}
	if !payload.Error {
		t.Errorf("Expected error to be true but got %v", payload.Error)
	}
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code to be %d but got %d", http.StatusInternalServerError, rr.Code)
	}
}
