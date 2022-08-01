package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

var a App

func TestMain(m *testing.M) {
	a.InitApp(
		GoDotEnvVariable("TEST_HOST"),
		GoDotEnvVariable("DB_USERNAME"),
		GoDotEnvVariable("DB_PASSWORD"),
		GoDotEnvVariable("DB_NAME"))
	ensureTableExists()
	code := m.Run()
	clearTable()
	os.Exit(code)
}

const tableCreationQuery = `CREATE TABLE IF NOT EXISTS products (
    id SERIAL,
    name TEXT NOT NULL,
    price NUMERIC(10, 2) NOT NULL DEFAULT 0.00,
    CONSTRAINT products_pkey PRIMARY KEY (id)
)`

func ensureTableExists() {
	if _, err := a.DB.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func clearTable() {
	a.DB.Exec("DELETE FROM products")
	a.DB.Exec("ALTER SEQUENCE products_id_seq RESTART with 1")
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)
	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func addProducts(numProducts int) {
	if numProducts < 1 {
		numProducts = 1
	}
	for i := 0; i < numProducts; i++ {
		a.DB.Exec("INSERT INTO products(name, price) values ($1, $2)", "Product "+strconv.Itoa(i), (i+1.0)*10)
	}
}

// Tests
func TestEmptyTable(t *testing.T) {
	clearTable()
	req, _ := http.NewRequest("GET", "/products", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array, but got %s\n", body)
	}
}

func TestGetNonExistentProduct(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/product/11", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["error"] != "Product not found" {
		t.Errorf("Expected error 'Product not found' but got %s\n ", m["error"])
	}
}

func TestCreateProduct(t *testing.T) {
	clearTable()

	productJson := []byte(`{"name": "test product", "price": 11.22 }`)
	req, _ := http.NewRequest("POST", "/product", bytes.NewBuffer(productJson))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != "test product" {
		t.Errorf("Expected product name to be 'test product' but got '%v'", m["name"])
	}

	if m["price"] != 11.22 {
		t.Errorf("Expected product price to be '11.22' but got '%v'", m["price"])
	}

	if m["id"] != 1.0 {
		t.Errorf("Expected product id to be '1' but got '%v'", m["id"])
	}
}

func TestDeleteProduct(t *testing.T) {
	clearTable()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/product/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/product/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestUpdateProduct(t *testing.T) {
	clearTable()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)
	var originalProduct map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalProduct)

	var productJson = []byte(`{"name": "test product - updated name", "price": 11.22}`)
	req, _ = http.NewRequest("PUT", "/product/1", bytes.NewBuffer(productJson))
	req.Header.Set("Content-Type", "application/json")
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["id"] != originalProduct["id"] {
		t.Errorf("Expected id to remain the same (%v) but got '%v'\n", originalProduct["id"], m["id"])
	}

	if m["name"] == originalProduct["name"] {
		t.Errorf("Expected product name to change from '%v' to '%v'. But got '%v'", originalProduct["name"], m["name"], m["name"])
	}

	if m["price"] == originalProduct["price"] {
		t.Errorf("Expected product price to change from '%v' to '%v'. But got '%v'", originalProduct["price"], m["price"], m["price"])
	}

}
