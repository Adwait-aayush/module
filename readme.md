# 📦 go-utils-module

A simple, reusable Go module providing essential utilities for building modern web backends. This module helps you handle:

- ✅ File uploads with size/type validation  
- 🔐 JWT Authentication (generation & validation) *(in progress)*  
- 🔁 Middleware support (auth, logging, etc.) *(in progress)*  
- 📂 Directory creation *(in progress)*  
- 🔢 Random string generation  
- 📄 JSON parsing and writing *(in progress)*  

Perfect for developers building REST APIs or microservices in Go.

🧪 **Fully written with unit tests in Go to ensure reliability and maintainability.**

---

## ✨ Features

### 📁 File Uploads  
Upload single or multiple files with automatic MIME-type and size validation. Supports custom file types and max-size rules.

### 🔐 JWT Authentication *(in progress)*  
Generate and validate JWTs using a secret key. Built for secure route protection and session handling.

### 🧱 Middleware Support *(in progress)*  
Plug-and-play middleware utilities like auth guards, request logging, or your own custom logic.

### 🔢 Random String Generator  
Create secure, random, URL-safe strings. Great for file renaming, access codes, and unique IDs.

### 📄 JSON Utilities *(in progress)*  
Safely decode JSON from HTTP requests and encode structured responses with proper error handling.

###    SlugGeneration
Generate slug from string.
---

## 🚀 Installation

```bash
go get github.com/Ayush/module
```

---

## 📚 Usage

### File Uploads

```go
import "github.com/Ayush/module"

func uploadHandler(w http.ResponseWriter, r *http.Request) {
    m := module.Module{
        MaxFileSize:      10 * 1024 * 1024, // 10 MB
        AllowedFileTypes: []string{"image/jpeg", "image/png", "application/pdf"},
    }
    files, err := m.UploadFiles(r, "./uploads")
    if err != nil {
        http.Error(w, "Upload error: "+err.Error(), http.StatusInternalServerError)
        return
    }
    for _, file := range files {
        fmt.Fprintf(w, "Uploaded: %s as %s (%d bytes)\n", file.FileName, file.NewFileName, file.FileSize)
    }
}
```

### Upload a Single File

```go
import "github.com/Ayush/module"

func uploadOneHandler(w http.ResponseWriter, r *http.Request) {
    m := module.Module{
        MaxFileSize:      10 * 1024 * 1024,
        AllowedFileTypes: []string{"application/pdf"},
    }
    file, err := m.UploadFile(r, "./uploads")
    if err != nil {
        http.Error(w, "Upload error: "+err.Error(), http.StatusInternalServerError)
        return
    }
    fmt.Fprintf(w, "Uploaded: %s as %s (%d bytes)\n", file.FileName, file.NewFileName, file.FileSize)
}
```

### Generate a Random String

```go
import "github.com/Ayush/module"

func main() {
    m := module.Module{}
    randomStr := m.GenRandomString(16)
    fmt.Println("Random string:", randomStr)
}
```

### Generate Slug

```go
import "github.com/Ayush/module"

func main() {
    m := module.Module{}
    str:="Random String"
    Str := m.MakeSlug(str)
    fmt.Println("Slug:",Str)
}

```


## 🤝 Contributing

Contributions, issues, and feature requests are welcome!  
Feel free to check [issues](https://github.com/Ayush/module/issues) or submit a pull request.

---

## 📄 License

MIT License. See [LICENSE](LICENSE) for details.