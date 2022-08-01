package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"strconv"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

func (app *App) InitApp(host, user, password, dbname string) {
	connectionString := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", host, user, password, dbname)

	var err error
	app.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal("Error connecting to db: ", err)
	}
	log.Println("Connected to DB....")
	app.createDb()
	app.Router = mux.NewRouter()
	app.initRoutes()
}

func (app *App) createDb() {
	tableCreationQuery := `CREATE TABLE IF NOT EXISTS products (
    id SERIAL,
    name TEXT NOT NULL,
    price NUMERIC(10, 2) NOT NULL DEFAULT 0.00,
    CONSTRAINT products_pkey PRIMARY KEY (id)
)`
	if _, err := app.DB.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func (app *App) initRoutes() {
	app.Router.HandleFunc("/products", app.getProducts).Methods("GET")
	app.Router.HandleFunc("/product", app.createProduct).Methods("POST")
	app.Router.HandleFunc("/product/{id:[0-9]+}", app.getProduct).Methods("GET")
	app.Router.HandleFunc("/product/{id:[0-9]+}", app.updateProduct).Methods("PUT")
	app.Router.HandleFunc("/product/{id:[0-9]+}", app.deleteProduct).Methods("DELETE")

}

func (app *App) Run(port string) {
	log.Println("starting server on port: ", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), app.Router))
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJson(w, code, map[string]string{"error": message})
}

func respondWithJson(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// Handlers
func (app *App) getProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid product id")
		return
	}

	product := Product{ID: id}
	if err := product.getProduct(app.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Product not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	respondWithJson(w, http.StatusOK, product)
}

func (app *App) getProducts(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count > 10 || count < 1 {
		count = 10
	}

	if start < 0 {
		start = 0
	}

	products, err := getProducts(app.DB, start, count)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJson(w, http.StatusOK, products)
}

func (app *App) createProduct(w http.ResponseWriter, r *http.Request) {
	var p Product
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	defer r.Body.Close()

	if err := p.createProduct(app.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJson(w, http.StatusCreated, p)
}

func (app *App) updateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	var product Product
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&product); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request")
		return
	}

	defer r.Body.Close()
	product.ID = id

	if err := product.updateProduct(app.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJson(w, http.StatusOK, product)
}

func (app *App) deleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	product := Product{ID: id}

	if err := product.deleteProduct(app.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJson(w, http.StatusOK, map[string]string{"result": "successfully deleted product"})
}
