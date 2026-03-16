package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Book struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	ISBN      string             `json:"isbn" bson:"isbn"`
	Title     string             `json:"title" bson:"title"`
	Author    Author             `json:"author" bson:"author"`
	CreatedAt time.Time          `json:"created_at,omitempty" bson:"created_at,omitempty"`
	UpdatedAt time.Time          `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
}

type Author struct {
	FirstName string `json:"first_name" bson:"first_name"`
	LastName  string `json:"last_name" bson:"last_name"`
}

type bookRequest struct {
	ISBN   string `json:"isbn"`
	Title  string `json:"title"`
	Author Author `json:"author"`
}

var bookCollection *mongo.Collection

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("cannot load .env, using system environment")
	}

	collection, err := connectMongo()
	if err != nil {
		log.Fatal(err)
	}
	bookCollection = collection

	router := mux.NewRouter()
	router.HandleFunc("/api/books", getBooks).Methods(http.MethodGet)
	router.HandleFunc("/api/books/{id}", getBook).Methods(http.MethodGet)
	router.HandleFunc("/api/books", createBook).Methods(http.MethodPost)
	router.HandleFunc("/api/books/{id}", updateBook).Methods(http.MethodPut)
	router.HandleFunc("/api/books/{id}", deleteBook).Methods(http.MethodDelete)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("server started on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func connectMongo() (*mongo.Collection, error) {
	uri := os.Getenv("CONNECT_DB")
	if uri == "" {
		uri = os.Getenv("connectDB")
	}
	if uri == "" {
		return nil, &appError{StatusCode: http.StatusInternalServerError, Message: "missing CONNECT_DB in environment"}
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "crud_demo"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return client.Database(dbName).Collection("books"), nil
}

func getBooks(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	cursor, err := bookCollection.Find(ctx, bson.M{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer cursor.Close(ctx)

	var books []Book
	if err := cursor.All(ctx, &books); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, books)
}

func getBook(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid book id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var book Book
	err = bookCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&book)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			writeError(w, http.StatusNotFound, "book not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, book)
}

func createBook(w http.ResponseWriter, r *http.Request) {
	var payload bookRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if payload.ISBN == "" || payload.Title == "" || payload.Author.FirstName == "" || payload.Author.LastName == "" {
		writeError(w, http.StatusBadRequest, "isbn, title, author.first_name, author.last_name are required")
		return
	}

	now := time.Now()
	book := Book{
		ID:        primitive.NewObjectID(),
		ISBN:      payload.ISBN,
		Title:     payload.Title,
		Author:    payload.Author,
		CreatedAt: now,
		UpdatedAt: now,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if _, err := bookCollection.InsertOne(ctx, book); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, book)
}

func updateBook(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid book id")
		return
	}

	var payload bookRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updateFields := bson.M{}
	if payload.ISBN != "" {
		updateFields["isbn"] = payload.ISBN
	}
	if payload.Title != "" {
		updateFields["title"] = payload.Title
	}
	if payload.Author.FirstName != "" || payload.Author.LastName != "" {
		author := bson.M{}
		if payload.Author.FirstName != "" {
			author["first_name"] = payload.Author.FirstName
		}
		if payload.Author.LastName != "" {
			author["last_name"] = payload.Author.LastName
		}
		updateFields["author"] = author
	}
	if len(updateFields) == 0 {
		writeError(w, http.StatusBadRequest, "empty update payload")
		return
	}
	updateFields["updated_at"] = time.Now()

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedBook Book
	err = bookCollection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": updateFields},
		opts,
	).Decode(&updatedBook)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			writeError(w, http.StatusNotFound, "book not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, updatedBook)
}

func deleteBook(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(mux.Vars(r)["id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid book id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	result, err := bookCollection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if result.DeletedCount == 0 {
		writeError(w, http.StatusNotFound, "book not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "book deleted"})
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

type appError struct {
	StatusCode int
	Message    string
}

func (e *appError) Error() string {
	return e.Message
}
