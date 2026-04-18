package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (app *Application) createUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	// create table if it doesn't exist (id omitted for brevity)
	query := `CREATE TABLE IF NOT EXISTS users (firstname varchar(50), lastname varchar(50));`
	if _, err := app.db.Exec(ctx, query); err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	var now string
	if err := app.db.QueryRow(ctx, "SELECT now()::text").Scan(&now); err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	insertQuery := `Insert into users values ('Amit', 'Patel') returning firstname`

	var fname string
	err := app.db.QueryRow(ctx, insertQuery).Scan(&fname)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	getUsers := `SELECT firstname, lastname FROM users;`

	type User struct {
		Firstname string `json:"firstname"`
		Lastname  string `json:"lastname"`
	}

	var u []User

	res, err := app.db.Query(ctx, getUsers)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer res.Close()

	for res.Next() {
		var user User
		if err := res.Scan(&user.Firstname, &user.Lastname); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		u = append(u, user)
	}

	if err := res.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("users %#v\n", u)

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(u)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
