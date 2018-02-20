package main

import (
	"database/sql"
	"fmt"
)

func CreateHallList() error {
	db, err := connectDB(adminDB)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS HallList (
			 SL int(100) PRIMARY KEY AUTO_INCREMENT,
			 Name varchar(100) NOT NULL
			)`)

	fmt.Println(query)

	_, err = db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func GetHallList(db *sql.DB) ([]string, error) {
	var HallList []string
	var Hall string

	rows, err := db.Query("SELECT Name FROM HallList")
	if err != nil {
		return HallList, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&Hall)
		if err != nil {
			return HallList, err
		} else {
			HallList = append(HallList, Hall)
		}
	}
	err = rows.Err()
	if err != nil {
		return HallList, err
	}

	return HallList, nil

}

func AddToHallList(hallname string) error {
	db, err := connectDB(adminDB)
	if err != nil {
		return err
	}
	defer db.Close()

	query := fmt.Sprintf("INSERT HallList SET Name=?")
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(hallname)
	if err != nil {
		return err
	}

	return nil

}

func CreatePurchaseTable(hallName string) error {
	db, err := connectDB(purchaseDB)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS ` + "`" + hallName + "`" + ` (
  SL int(100) PRIMARY KEY AUTO_INCREMENT,
  Name varchar(100) NOT NULL,
  Phone varchar(100) NOT NULL,
  Email varchar(100) NOT NULL,
  Movie varchar(100) NOT NULL,
  Category varchar(100) NOT NULL,
  Price varchar(50) NOT NULL,
  Time varchar(50) NOT NULL,
  Date varchar(50) NOT NULL,
  Quantity int(50) NOT NULL,
  OrderID varchar(100) NOT NULL,
  ShowID varchar(100) NOT NULL,
  BookingTime timestamp NULL DEFAULT NULL
)`)

	fmt.Println(query)

	_, err = db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func CreateCategoryTable() error {
	db, err := connectDB(adminDB)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS Categories (
	 ID varchar(100) NOT NULL,
     HallName varchar(100) NOT NULL,
	 Name varchar(100) NOT NULL,
	 Price int NOT NULL,
	 TotalSeats int NOT NULL,
	 RemSeats int NOT NULL
)`)

	fmt.Println(query)

	_, err = db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func CreateHallTable(hallName string) error {
	db, err := connectDB(adminDB)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS ` + "`" + hallName + "`" + ` (
	 ID varchar(100) PRIMARY KEY,
     SL int(100) UNIQUE AUTO_INCREMENT,
     Name varchar(100) NOT NULL,
     Description varchar(100) NOT NULL,
     Trailer varchar(100) NOT NULL,
     Poster varchar(100) NOT NULL,
     Time time NOT NULL,
     Date date NOT NULL
)`)

	fmt.Println(query)

	_, err = db.Exec(query)
	if err != nil {
		return err
	}

	err = AddToHallList(hallName)
	if err != nil {
		return err
	}

	return nil
}
