package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

const (
	adminDB    = "admin"
	purchaseDB = "purchases"
)

var (
	dbPassword = os.Getenv("DB_PASSWORD")
	dbEndpoint = os.Getenv("DB_ENDPOINT")
	dbUser     = os.Getenv("DB_USER")
)

type Purchase struct {
	Name  string
	Phone string
	Email string
	Show
	Quantity    int
	OrderID     string
	BookingTime string
}

type Show struct { //

	HallName string
	ID       string

	Category `json:"category"`
	Movie    `json:"movie"`
}

type Category struct { //
	Price int
	Name  string
	Seats int
}

type Movie struct { //
	Name        string
	Description string
	Trailer     string
	Poster      string
	Date        string
	Time        string
}

func connectDB(dbName string) (*sql.DB, error) {
	dnsStr := fmt.Sprintf("%s:%s@tcp(%s)/%s",
		dbUser, dbPassword, dbEndpoint, dbName,
	)
	db, err := sql.Open("mysql", dnsStr)
	_, err = db.Query("SET time_zone = 'Asia/Kolkata'")
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	return db, nil
}

func (c *Category) AddCategory(db *sql.DB, hallname string, ID string) error {
	query := fmt.Sprintf("INSERT Categories SET HallName=?, Name=?, Price=?, TotalSeats=?, RemSeats=?, ID=?")
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(hallname, c.Name, c.Price, c.Seats, c.Seats, ID)
	if err != nil {
		return err
	}

	return nil
}

func LookupRemSeats(db *sql.DB, ID, Category string) (int, error) {
	rows, err := db.Query("SELECT RemSeats FROM Categories WHERE ID=? AND Name=?", ID, Category)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var RemSeats int
	for rows.Next() {
		err = rows.Scan(&RemSeats)
		if err != nil {
			return 0, err
		}
	}
	err = rows.Err()
	if err != nil {
		return 0, err
	}

	return RemSeats, nil
}

func GetInfo(db *sql.DB, table string) (*Show, error) {
	rows, err := db.Query("SELECT ID, HallName, Name, Price, TotalSeats FROM Categories WHERE HallName=?", table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	s := &Show{}

	for rows.Next() {
		err = rows.Scan(&s.ID, &s.HallName, &s.Category.Name, &s.Category.Price, &s.Category.Seats)
		if err != nil {
			return nil, err
		}
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	rows, err = db.Query("SELECT Name, Description, Trailer, Poster, Time, Date FROM " + "`" + table + "`")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&s.Movie.Name, &s.Movie.Description, &s.Movie.Trailer, &s.Movie.Poster, &s.Movie.Time, &s.Movie.Date)
		if err != nil {
			return nil, err
		}
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return s, nil
}

func LookupHallMovie(db *sql.DB, ID string) (string, string, error) {
	rows, err := db.Query("SELECT HallName FROM Categories WHERE ID=?", ID)
	if err != nil {
		return "", "", err
	}
	defer rows.Close()

	var movie, hall string
	for rows.Next() {
		err = rows.Scan(&hall)
		if err != nil {
			return "", "", err
		}
	}
	err = rows.Err()
	if err != nil {
		return "", "", err
	}

	query := fmt.Sprintf("SELECT Name FROM " + "`" + hall + "`" + " WHERE ID=?")

	rows, err = db.Query(query, ID)
	if err != nil {
		return "", "", err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&movie)
		if err != nil {
			return "", "", err
		}
	}
	err = rows.Err()
	if err != nil {
		return "", "", err
	}
	return movie, hall, nil
}

func EditSeats(db *sql.DB, shifter int, ID string, category string) error {
	rows, err := db.Query("SELECT RemSeats FROM Categories WHERE ID=? AND Name=?", ID, category)
	if err != nil {
		return err
	}
	defer rows.Close()

	var remSeats int
	for rows.Next() {
		err = rows.Scan(&remSeats)

		if err != nil {
			return err
		}
	}
	err = rows.Err()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Remaining seats : ", remSeats)

	stmt, err := db.Prepare("UPDATE Categories SET RemSeats=? WHERE ID=? AND Name=?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(remSeats+shifter, ID, category)
	if err != nil {
		return err
	}

	return nil

}

func (p *Purchase) AddPurchase(db *sql.DB, amount int) error {
	query := fmt.Sprintf("INSERT " + "`" + p.Show.HallName + "`" + "SET Price=?, Name=?, Phone=?, Email=?, BookingTime=?, Quantity=?, OrderID=?, Movie=?, Category=?, Time=?, Date=?, ShowID=?")
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(amount, p.Name, p.Phone, p.Email, p.BookingTime, p.Quantity, p.OrderID, p.Show.Movie.Name, p.Show.Category.Name, p.Show.Movie.Time, p.Show.Movie.Date, p.Show.ID)
	if err != nil {
		return err
	}

	return nil
}

func (h *Show) AddShow(db *sql.DB) error {
	query := fmt.Sprintf("INSERT " + "`" + h.HallName + "`" + "SET Name=?, Description=?, Trailer=?, Poster=?, Time=?, Date=?, ID=?")
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}

	// ID = unix timestamp

	_, err = stmt.Exec(h.Movie.Name, h.Movie.Description, h.Movie.Trailer, h.Movie.Poster, h.Movie.Time, h.Movie.Date, h.ID)
	if err != nil {
		return err
	}

	return nil
}