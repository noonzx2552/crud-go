# MongoDB CRUD

This project is a simple CRUD API built with `Gorilla Mux` and `MongoDB`.

## Files

- `main.go` contains the API server
- `.env` contains environment variables
- `go.mod` and `go.sum` manage dependencies

## Requirements

- Go `1.26.1`
- MongoDB Atlas or any reachable MongoDB instance

## Environment

Example `.env` file:

```env
CONNECT_DB="your-mongodb-connection-string"
DB_NAME="crud_demo"
PORT="8000"
```

Notes:
- `CONNECT_DB` is your MongoDB connection string
- `DB_NAME` is the database name
- The code uses the `books` collection

## Install Dependencies

```bash
go mod tidy
```

## Run

```bash
go run .
```

The server will start at:

```text
http://localhost:8000
```

## Test Compile

```bash
go test ./...
```

## API Endpoints

- `GET /api/books`
- `GET /api/books/{id}`
- `POST /api/books`
- `PUT /api/books/{id}`
- `DELETE /api/books/{id}`

## Example Request Body

```json
{
  "isbn": "978-1234567890",
  "title": "Mongo CRUD",
  "author": {
    "first_name": "John",
    "last_name": "Doe"
  }
}
```

## cURL Examples

Get all books:

```bash
curl http://localhost:8000/api/books
```

Create a book:

```bash
curl -X POST http://localhost:8000/api/books ^
  -H "Content-Type: application/json" ^
  -d "{\"isbn\":\"978-1234567890\",\"title\":\"Mongo CRUD\",\"author\":{\"first_name\":\"John\",\"last_name\":\"Doe\"}}"
```

Get a book by ID:

```bash
curl http://localhost:8000/api/books/{id}
```

Update a book:

```bash
curl -X PUT http://localhost:8000/api/books/{id} ^
  -H "Content-Type: application/json" ^
  -d "{\"title\":\"Mongo CRUD Updated\",\"author\":{\"first_name\":\"Jane\",\"last_name\":\"Doe\"}}"
```

Delete a book:

```bash
curl -X DELETE http://localhost:8000/api/books/{id}
```

## Notes About go.mod And go.sum

- `go.mod` is already correct for this project
- `go.sum` should not be edited manually
- If you add or remove packages, run `go mod tidy`
