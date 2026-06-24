# Go Firebird REST API

A lightweight, high-performance REST API written in Go utilizing the Gin Gonic framework and a pure Go Firebird database driver (`github.com/nakagami/firebirdsql`). 

It connects directly to a Firebird database (`.GDB` or `.FDB`), normalization of Windows file paths, handles null safety, and implements connection pooling, request logging, server-side pagination, and graceful shutdown.

---

## Features

* **Pure Go Firebird Driver**: Native database operations without external DLL/C client library dependencies or CGO enabled.
* **Windows Path Normalization**: Automatically converts backslashes in Windows file paths for compatibility with URI parsing.
* **Connection Pooling**: Configures database connection counts and reuse lifetimes.
* **Query-level Pagination (FIRST / SKIP)**: Incorporates server-side limit and offset query mapping for items, search, and category items.
* **Graceful Shutdown**: Intercepts OS termination signals (`SIGINT`, `SIGTERM`), waiting for running requests to finish before shutting down the server and database connection pool.
* **Standardized JSON API Responses**: Uniform response structure:
  ```json
  {
    "success": true,
    "message": "Data ditemukan",
    "data": [...]
  }
  ```

---

## Project Structure

```text
rest_go_toko/
├── main.go               # App entrypoint & graceful shutdown
├── go.mod                # Go module dependencies
├── .env                  # Configuration file (ignored in git)
├── .env.example          # Sample configuration file
├── config/
│   └── config.go         # Configuration loader via godotenv
├── database/
│   └── database.go       # Connection pool initialization & ping
├── models/
│   ├── category.go       # Category model matching ITEM_CATEGORY
│   ├── item.go           # Item model matching ITEM
│   └── response.go       # API Response types
├── handlers/
│   ├── response.go       # HTTP JSON response wrapper & pagination parser
│   ├── category_handler.go  # Category database handler logic
│   └── item_handler.go      # Item database handler logic
└── routes/
    └── routes.go         # Route configurations
```

---

## Instructions of Use

### 1. Requirements
* Go 1.24+ Installed
* Access to a Firebird database

### 2. Configuration Setup
Create a `.env` file in the root directory (you can copy `.env.example`):
```env
DB_HOST=127.0.0.1
DB_PORT=3051
DB_PATH=H:\AMAN\DBTOKOPINTAR_NEW.GDB
DB_USER=pos
DB_PASSWORD=pos

SERVER_PORT=8080
```

### 3. Installation of Dependencies
Download and clean up required Go packages:
```bash
go mod tidy
```

### 4. Running the Server
Start the API server directly from the source code:
```bash
go run main.go
```

### 5. Building the Executable
Compile the project to a standalone executable binary (e.g., Windows `.exe`):
```bash
go build -o rest_go_toko.exe main.go
```
Then run:
```bash
./rest_go_toko.exe
```

---

## API Endpoints & cURL Examples

Once the server is running on port `8080`, test the endpoints using cURL:

### 1. Get All Categories
```bash
curl -s http://localhost:8080/api/categories
```

### 2. Get Paginated Items (default size: 50, page: 1)
```bash
curl -s "http://localhost:8080/api/items?page=1&limit=50"
```

### 3. Get Item Details (by ITEMNO)
```bash
curl -s http://localhost:8080/api/items/00001
```

### 4. Search Items by Name (case-insensitive LIKE search with pagination)
```bash
curl -s "http://localhost:8080/api/items/search?q=mie&page=1&limit=50"
```

### 5. Get Items by Category ID with Pagination
```bash
curl -s "http://localhost:8080/api/categories/1/items?page=1&limit=50"
```
