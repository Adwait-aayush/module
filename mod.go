package module

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const source = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_+"

type Module struct {
	MaxFileSize int64
	AllowedFileTypes []string
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

func (m *Module) UploadFile(r *http.Request, uploadDir string,rename ...bool)(*UploadedFile,error){
renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}
	files,err:=m.UploadFiles(r, uploadDir, renameFile)
	if err!=nil{
		return nil,err
	}
	return files[0],nil
}


func (m *Module) UploadFiles(r *http.Request, uploadDir string, rename ...bool) ([]*UploadedFile, error) {
	
	renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}

	var uploadedFiles []*UploadedFile

	// // Create the upload directory if it does not exist.
	// err := m.CreateDirIfNotExist(uploadDir)
	// if err != nil {
	// 	return nil, err
	// }

	
	if m.MaxFileSize == 0 {
		m.MaxFileSize = 1024 * 1024 * 10 // Default to 10MB
	}

	
	err := r.ParseMultipartForm(int64(m.MaxFileSize))
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
