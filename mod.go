package module

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const source = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_+"

type Module struct {
	MaxFileSize           int64
	AllowedFileTypes      []string
	MaxJsonSize           int64
	AllowUnknownFileTypes bool
}

func (m *Module) GenRandomString(n int) string {
	s, r := make([]rune, n), []rune(source)
	for i := range s {
		p, _ := rand.Prime(rand.Reader, len(r))
		x, y := p.Uint64(), uint64(len(r))
		s[i] = r[x%y]
	}
	return string(s)
}

type UploadedFile struct {
	FileName    string
	NewFileName string
	FileSize    int64
}

func (m *Module) UploadFile(r *http.Request, uploadDir string, rename ...bool) (*UploadedFile, error) {
	renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}
	files, err := m.UploadFiles(r, uploadDir, renameFile)
	if err != nil {
		return nil, err
	}
	return files[0], nil
}

func (m *Module) UploadFiles(r *http.Request, uploadDir string, rename ...bool) ([]*UploadedFile, error) {

	renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}

	var uploadedFiles []*UploadedFile

	// Create the upload directory if it does not exist.
	err := m.CreateDirIfNotExist(uploadDir)
	if err != nil {
		return nil, err
	}

	if m.MaxFileSize == 0 {
		m.MaxFileSize = 1024 * 1024 * 10 // Default to 10MB
	}

	err = r.ParseMultipartForm(int64(m.MaxFileSize))
	if err != nil {
		return nil, fmt.Errorf("error parsing form data: %v", err)
	}

	for _, fHeaders := range r.MultipartForm.File {
		for _, hdr := range fHeaders {
			uploadedFiles, err = func(uploadedFiles []*UploadedFile) ([]*UploadedFile, error) {
				var uploadedFile UploadedFile
				infile, err := hdr.Open()
				if err != nil {
					return nil, err
				}
				defer infile.Close()

				if hdr.Size > int64(m.MaxFileSize) {
					return nil, fmt.Errorf("the uploaded file is too big, and must be less than %d", m.MaxFileSize)
				}

				buff := make([]byte, 512)
				_, err = infile.Read(buff)
				if err != nil {
					return nil, err
				}

				allowed := false
				filetype := http.DetectContentType(buff)
				if len(m.AllowedFileTypes) > 0 {
					for _, x := range m.AllowedFileTypes {
						if strings.EqualFold(filetype, x) {
							allowed = true
						}
					}
				} else {
					allowed = true
				}

				if !allowed {
					return nil, errors.New("the uploaded file type is not permitted")
				}

				_, err = infile.Seek(0, 0)
				if err != nil {
					fmt.Println(err)
					return nil, err
				}

				if renameFile {
					uploadedFile.NewFileName = fmt.Sprintf("%s%s", m.GenRandomString(25), filepath.Ext(hdr.Filename))
				} else {
					uploadedFile.NewFileName = hdr.Filename
				}
				uploadedFile.FileName = hdr.Filename

				var outfile *os.File
				defer outfile.Close()

				if outfile, err = os.Create(filepath.Join(uploadDir, uploadedFile.NewFileName)); nil != err {
					return nil, err
				}
				fileSize, err := io.Copy(outfile, infile)
				if err != nil {
					return nil, err
				}
				uploadedFile.FileSize = fileSize

				uploadedFiles = append(uploadedFiles, &uploadedFile)

				return uploadedFiles, nil
			}(uploadedFiles)
			if err != nil {
				return uploadedFiles, err
			}
		}
	}
	return uploadedFiles, nil
}

func (m *Module) CreateDirIfNotExist(dir string) error {
	const mode = 0755
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, mode); err != nil {
			return fmt.Errorf("error creating upload directory: %v", err)
		}
	}
	return nil
}

func (m *Module) MakeSlug(s string) (string, error) {
	if s == "" {
		return "", errors.New("slug cannot be empty")
	}
	var re = regexp.MustCompile(`[^a-zA-Z0-9]+`)
	slug := strings.Trim(re.ReplaceAllString(strings.ToLower(s), "-"), "-")
	if len(slug) == 0 {
		return "", errors.New("slug cannot be empty")
	}
	return slug, nil
}

type JSONResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (m *Module) ReadJSON(w http.ResponseWriter, r *http.Request, data interface{}) error {
	maxbyte := 1024 * 1024 // 1MB
	if m.MaxJsonSize != 0 {
		maxbyte = int(m.MaxJsonSize)
	}
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxbyte))
	dec := json.NewDecoder(r.Body)
	if !m.AllowUnknownFileTypes {
		dec.DisallowUnknownFields()
	}
	err := dec.Decode(data)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("badly formed json")
		case errors.As(err, &unmarshalTypeError):
			return fmt.Errorf("invalid type for field %q: %s", unmarshalTypeError.Field, unmarshalTypeError.Value)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		case errors.As(err, &invalidUnmarshalError):
			return fmt.Errorf("invalid unmarshal error: %s", invalidUnmarshalError.Error())
		case errors.Is(err, io.EOF):
			return errors.New("request body must not be empty")
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("unknown field %s", fieldName)

		case err.Error() == "http: request body too large":
			return fmt.Errorf("request body must not be larger than %d bytes", maxbyte)

		default:
			return err
		}
	}
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("request body must only contain a single JSON object")
	}

	return nil
}

func (m *Module) WriteJSON(w http.ResponseWriter, status int, data interface{},headers ...http.Header) error {
	out,err:=json.Marshal(data)
	if err!=nil{
		return fmt.Errorf("error marshalling json: %v", err)
	}
	if len(headers)>0{
		for k, v := range headers[0] {
			w.Header()[k]=v
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_,err=w.Write(out)
	if err!=nil{
		return err
	}
	return nil
}

func (m *Module) ErrorJSON(w http.ResponseWriter,err error,status ...int)error{
	statuscode:=http.StatusBadRequest
	if len(status)>0{
		statuscode=status[0]
	}
	var payload JSONResponse
	payload.Error = true
	payload.Message = err.Error()
	return m.WriteJSON(w, statuscode, payload)
}

