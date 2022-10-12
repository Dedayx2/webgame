package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"unicode"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

// Product data type for export
type Product struct {
	ID          int
	Name        string
	Price       float32
	Img         string
	Description string
}

var tpl *template.Template

var db *sql.DB

func main() {
	tpl, _ = template.ParseGlob("templates/*.html")
	var err error
	// Never use _, := db.Open(), resources need to be released with db.Close
	db, err = sql.Open("mysql", "root:Lostlakex2@tcp(localhost:3306)/testdb")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	fs := http.FileServer(http.Dir("assets"))
	http.Handle("/assets/", http.StripPrefix("/assets", fs))
	http.HandleFunc("/login", loginHandler)
	// http.HandleFunc("/loginauth", loginAuthHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/insert", insertHandler)
	http.HandleFunc("/browse", browseHandler)
	http.HandleFunc("/update/", updateHandler)
	http.HandleFunc("/updateresult/", updateResultHandler)
	http.HandleFunc("/delete/", deleteHandler)
	http.HandleFunc("/productsearch", productSearchHandler)
	http.HandleFunc("/", homePageHandler)
	http.ListenAndServe("localhost:8080", nil)
}

func browseHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("*****browseHandler running*****")
	stmt := "SELECT * FROM products"
	// func (db *DB) Query(query string, args ...interface{}) (*Rows, error)
	rows, err := db.Query(stmt)
	if err != nil {
		panic(err)
	}
	// type sql.Row does not have a .Close() Method but sql.Rows does and must be run
	defer rows.Close()
	var products []Product
	// func (rs *Rows) Next() bool
	for rows.Next() {
		var p Product
		// func (rs *Rows) Scan(dest ...interface{}) error
		err = rows.Scan(&p.ID, &p.Name, &p.Price, &p.Img, &p.Description)
		if err != nil {
			panic(err)
		}
		// func append(slice []Type, elems ...Type) []Type
		products = append(products, p)
	}
	tpl.ExecuteTemplate(w, "select.html", products)
}

func insertHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("*****insertHandler running*****")
	if r.Method == "GET" {
		tpl.ExecuteTemplate(w, "insert.html", nil)
		return
	}
	r.ParseForm()
	// func (r *Request) FormValue(key string) string
	name := r.FormValue("nameName")
	price := r.FormValue("priceName")
	img := r.FormValue("imgName")
	descr := r.FormValue("descrName")
	var err error
	if name == "" || price == "" || img == "" || descr == "" {
		fmt.Println("Error inserting row:", err)
		tpl.ExecuteTemplate(w, "insert.html", "Error inserting data, please check all fields.")
		return
	}
	var ins *sql.Stmt
	// don't use _, err := db.Query()
	// func (db *DB) Prepare(query string) (*Stmt, error)
	ins, err = db.Prepare("INSERT INTO `testdb`.`products` (`name`, `price`,`img`,`description`) VALUES (?, ?, ?,?);")
	if err != nil {
		panic(err)
	}
	defer ins.Close()
	// func (s *Stmt) Exec(args ...interface{}) (Result, error)
	res, err := ins.Exec(name, price, img, descr)

	// check rows affectect???????
	rowsAffec, _ := res.RowsAffected()
	if err != nil || rowsAffec != 1 {
		fmt.Println("Error inserting row:", err)
		tpl.ExecuteTemplate(w, "insert.html", "Error inserting data, please check all fields.")
		return
	}
	lastInserted, _ := res.LastInsertId()
	rowsAffected, _ := res.RowsAffected()
	fmt.Println("ID of last row inserted:", lastInserted)
	fmt.Println("number of rows affected:", rowsAffected)
	tpl.ExecuteTemplate(w, "insert.html", "Product Successfully Inserted")
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("*****updateHandler running*****")
	r.ParseForm()
	id := r.FormValue("idgame")
	row := db.QueryRow("SELECT * FROM testdb.products WHERE idgame = ?;", id)
	var p Product
	// func (r *Row) Scan(dest ...interface{}) error
	err := row.Scan(&p.ID, &p.Name, &p.Price, &p.Img, &p.Description)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/browse", 307)
		return
	}
	tpl.ExecuteTemplate(w, "update.html", p)
}

func updateResultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("*****updateResultHandler running*****")
	r.ParseForm()
	id := r.FormValue("idgame")
	name := r.FormValue("nameName")
	price := r.FormValue("priceName")
	img := r.FormValue("imgName")
	description := r.FormValue("descrName")
	// ins, err = db.Prepare("INSERT INTO `testdb`.`products` (`name`, `price`,`img`,`description`) VALUES (?, ?, ?,?);")
	upStmt := "UPDATE `testdb`.`products` SET `name` = ?, `price` = ?, `img` = ?,`description` = ? WHERE (`idgame` = ?);"
	// func (db *DB) Prepare(query string) (*Stmt, error)
	stmt, err := db.Prepare(upStmt)
	if err != nil {
		fmt.Println("error preparing stmt")
		panic(err)
	}
	fmt.Println("db.Prepare err:", err)
	fmt.Println("db.Prepare stmt:", stmt)
	defer stmt.Close()
	var res sql.Result
	// func (s *Stmt) Exec(args ...interface{}) (Result, error)
	res, err = stmt.Exec(id, name, price, img, description)
	rowsAff, _ := res.RowsAffected()
	if err != nil || rowsAff != 1 {
		fmt.Println(err)
		tpl.ExecuteTemplate(w, "result.html", "There was a problem updating the product")
		return
	}
	tpl.ExecuteTemplate(w, "result.html", "Product was Successfully Updated")
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("*****deleteHandler running*****")
	r.ParseForm()
	id := r.FormValue("idgame")
	//  func (db *DB) Prepare(query string) (*Stmt, error)
	del, err := db.Prepare("DELETE FROM `testdb`.`products` WHERE (`idgame` = ?);")
	if err != nil {
		panic(err)
	}
	defer del.Close()
	var res sql.Result
	res, err = del.Exec(id)
	rowsAff, _ := res.RowsAffected()
	fmt.Println("rowsAff:", rowsAff)

	if err != nil || rowsAff != 1 {
		fmt.Fprint(w, "Error deleting product")
		return
	}
	/*
		if err != nil {
			fmt.Fprint(w, "Error deleting product")
			return
		}
	*/
	fmt.Println("err:", err)
	tpl.ExecuteTemplate(w, "result.html", "Product was Successfully Deleted")
}

func productSearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tpl.ExecuteTemplate(w, "productsearch.html", nil)
		return
	}
	r.ParseForm()
	var P Product
	name := r.FormValue("productName")
	fmt.Println("name:", name)
	// stmt := `SELECT * FROM products WHERE name = '` + name + "';"
	stmt := "SELECT * FROM products WHERE name = ?;"
	// func (db *DB) QueryRow(query string, args ...interface{}) *Row
	row := db.QueryRow(stmt, name)
	// func (r *Row) Scan(dest ...interface{}) error
	err := row.Scan(&P.ID, &P.Name, &P.Price, &P.Img, &P.Description)
	if err != nil {
		panic(err)
	}
	tpl.ExecuteTemplate(w, "productsearch.html", P)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("*****loginHandler running*****")
	if r.Method == "GET" {
		tpl.ExecuteTemplate(w, "login.html", nil)
		return
	}
	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")
	fmt.Println("username:", username, "password:", password)
	// retrieve password from db to compare (hash) with user supplied password's hash
	var hash string
	stmt := "SELECT Hash FROM bcrypt WHERE Username = ?"
	row := db.QueryRow(stmt, username)
	err := row.Scan(&hash)
	fmt.Println("hash from db:", hash)
	if err != nil {
		fmt.Println("error selecting Hash in db by Username")
		tpl.ExecuteTemplate(w, "login.html", "check username and password")
		return
	}
	// func CompareHashAndPassword(hashedPassword, password []byte) error
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	// returns nill on succcess
	if err == nil {
		tpl.ExecuteTemplate(w, "login.html", nil)
		return
	}
	fmt.Println("incorrect password")
	tpl.ExecuteTemplate(w, "login.html", "check username and password")
}

// loginAuthHandler authenticates user login
// func loginAuthHandler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Println("*****loginAuthHandler running*****")
// 	r.ParseForm()
// 	username := r.FormValue("username")
// 	password := r.FormValue("password")
// 	fmt.Println("username:", username, "password:", password)
// 	// retrieve password from db to compare (hash) with user supplied password's hash
// 	var hash string
// 	stmt := "SELECT Hash FROM bcrypt WHERE Username = ?"
// 	row := db.QueryRow(stmt, username)
// 	err := row.Scan(&hash)
// 	fmt.Println("hash from db:", hash)
// 	if err != nil {
// 		fmt.Println("error selecting Hash in db by Username")
// 		tpl.ExecuteTemplate(w, "login.html", "check username and password")
// 		return
// 	}
// 	// func CompareHashAndPassword(hashedPassword, password []byte) error
// 	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
// 	// returns nill on succcess
// 	if err == nil {
// 		fmt.Fprint(w, "You have successfully logged in :)")
// 		return
// 	}
// 	fmt.Println("incorrect password")
// 	tpl.ExecuteTemplate(w, "login.html", "check username and password")
// }

// registerHandler serves form for registring new users
func registerHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("*****registerHandler running*****")

	r.ParseForm()
	username := r.FormValue("username")
	// check username for only alphaNumeric characters
	var nameAlphaNumeric = true
	for _, char := range username {
		// func IsLetter(r rune) bool, func IsNumber(r rune) bool
		// if !unicode.IsLetter(char) && !unicode.IsNumber(char) {
		if unicode.IsLetter(char) == false && unicode.IsNumber(char) == false {
			nameAlphaNumeric = false
		}
	}
	// check username pswdLength
	var nameLength bool
	if 5 <= len(username) && len(username) <= 50 {
		nameLength = true
	}
	// check password criteria
	password := r.FormValue("password")
	fmt.Println("password:", password, "\npswdLength:", len(password))
	// variables that must pass for password creation criteria
	var pswdLowercase, pswdUppercase, pswdNumber, pswdSpecial, pswdLength, pswdNoSpaces bool
	pswdNoSpaces = true
	for _, char := range password {
		switch {
		// func IsLower(r rune) bool
		case unicode.IsLower(char):
			pswdLowercase = true
		// func IsUpper(r rune) bool
		case unicode.IsUpper(char):
			pswdUppercase = true
		// func IsNumber(r rune) bool
		case unicode.IsNumber(char):
			pswdNumber = true
		// func IsPunct(r rune) bool, func IsSymbol(r rune) bool
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			pswdSpecial = true
		// func IsSpace(r rune) bool, type rune = int32
		case unicode.IsSpace(int32(char)):
			pswdNoSpaces = false
		}
	}
	if 11 < len(password) && len(password) < 60 {
		pswdLength = true
	}
	fmt.Println("pswdLowercase:", pswdLowercase, "\npswdUppercase:", pswdUppercase, "\npswdNumber:", pswdNumber, "\npswdSpecial:", pswdSpecial, "\npswdLength:", pswdLength, "\npswdNoSpaces:", pswdNoSpaces, "\nnameAlphaNumeric:", nameAlphaNumeric, "\nnameLength:", nameLength)
	if !pswdLowercase || !pswdUppercase || !pswdNumber || !pswdSpecial || !pswdLength || !pswdNoSpaces || !nameAlphaNumeric || !nameLength {
		tpl.ExecuteTemplate(w, "register.html", "please check username and password criteria")
		return
	}
	// check if username already exists for availability
	stmt := "SELECT UserID FROM bcrypt WHERE username = ?"
	row := db.QueryRow(stmt, username)
	var uID string
	err := row.Scan(&uID)
	if err != sql.ErrNoRows {
		fmt.Println("username already exists, err:", err)
		tpl.ExecuteTemplate(w, "register.html", "username already taken")
		return
	}
	// create hash from password
	var hash []byte
	// func GenerateFromPassword(password []byte, cost int) ([]byte, error)
	hash, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("bcrypt err:", err)
		tpl.ExecuteTemplate(w, "register.html", "there was a problem registering account")
		return
	}
	fmt.Println("hash:", hash)
	fmt.Println("string(hash):", string(hash))
	// func (db *DB) Prepare(query string) (*Stmt, error)
	var insertStmt *sql.Stmt
	insertStmt, err = db.Prepare("INSERT INTO bcrypt (Username, Hash) VALUES (?, ?);")
	if err != nil {
		fmt.Println("error preparing statement:", err)
		tpl.ExecuteTemplate(w, "register.html", "there was a problem registering account")
		return
	}
	defer insertStmt.Close()
	var result sql.Result
	//  func (s *Stmt) Exec(args ...interface{}) (Result, error)
	result, err = insertStmt.Exec(username, hash)
	rowsAff, _ := result.RowsAffected()
	lastIns, _ := result.LastInsertId()
	fmt.Println("rowsAff:", rowsAff)
	fmt.Println("lastIns:", lastIns)
	fmt.Println("err:", err)
	if err != nil {
		fmt.Println("error inserting new user")
		tpl.ExecuteTemplate(w, "register.html", "there was a problem registering account")
		return
	}
	fmt.Fprint(w, "login,html", "congrats, your account has been successfully created")
}

// registerAuthHandler creates new user in database

func homePageHandler(w http.ResponseWriter, r *http.Request) {

}
