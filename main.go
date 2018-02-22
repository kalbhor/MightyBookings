package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	port = ":8000"
)

var (
	loc, _ = time.LoadLocation("Asia/Kolkata")
)

func MakePayment(w http.ResponseWriter, r *http.Request) {
	/*
		expects `name`, `amount`, `email` & `phone` params in the request.
		Makes the payment request to trakNPay

	*/

	if r.Method == "GET" { // Change to POST

		db, err := connectDB(adminDB) // Make connection to admin db
		if err != nil {
			log.Println("Error while connecting to db : ", err)
		}
		defer db.Close()

		db2, err := connectDB(purchaseDB)
		if err != nil {
			log.Println("Error while connecting to db :", err)
		}
		defer db2.Close()

		m := make(map[string]string)

		for key, val := range r.URL.Query() {
			if len(val) != 0 {
				m[key] = template.HTMLEscapeString(val[0])
			}
		}

		fmt.Println(m)
		quantity, err := strconv.Atoi(m["quantity"])
		if err != nil {
			log.Println("Could not convert quantity to int : ", err)
			http.Error(w, err.Error(), 500)
			return
		}
		price, err := strconv.Atoi(m["price"])
		if err != nil {
			log.Println("Could not convert amount to int ", err)
			http.Error(w, err.Error(), 500)
			return
		}
		t := time.Now().In(loc)
		orderid := fmt.Sprintf("%d", t.Unix())
		bookingtime := fmt.Sprintf("%v-%v-%v %v:%v:%v", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
		m["order_id"] = orderid
		m["bookingtime"] = bookingtime

		p := &Purchase{Status: 0, Name: m["name"], Phone: m["phone"], Email: m["email"], BookingTime: bookingtime, Quantity: quantity, OrderID: orderid,
			Show: Show{ID: m["showid"], HallName: m["hall"],
				Movie: Movie{Name: m["movie"], Time: m["time"], Date: m["date"]}}}

		p.Show.Categories = make([]Category, 1)
		p.Show.Categories[0] = Category{Name: m["category"], Price: price}
		err = p.AddPurchase(db2)
		if err != nil {
			log.Println(err)
		}

		remSeats, err := RemainingSeats(db, m["showid"], m["category"])
		if err != nil {
			log.Println("Could not lookup remaining seats : ", err)
			http.Error(w, err.Error(), 500)
			return
		}

		if remSeats-quantity < 0 {
			log.Println("Quantity greater than remaining seats")
			http.Error(w, err.Error(), 500)
			return
		}

		err = EditSeats(db, -1*quantity, m["showid"], m["category"])
		if err != nil {
			log.Println("Could not deduct seats : ", err)
			http.Error(w, err.Error(), 500)
			return
		}

		resp, err := CallPayment(m)
		if err != nil {
			log.Println("Could not connect to payment page : ", err)
			http.Error(w, err.Error(), 500)
			return
		}
		defer resp.Body.Close()
		io.Copy(w, resp.Body)

		fmt.Println("Added to purchase table : ", p)

	}
}

func FailedPayment(w http.ResponseWriter, r *http.Request) {

	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	params := "https://xyz/?" + string(b)

	url, err := url.Parse(params)
	if err != nil {
		log.Println("Could not parse params : ", err)
	}
	fmt.Println(url)

}

func PaymentResponse(w http.ResponseWriter, r *http.Request) {
	/*
	   This is the URL traknpay will callback after payment.
	   If payment is succesful, twilio messages to car provider, customer, etc.
	*/
	db2, err := connectDB(purchaseDB)
	if err != nil {
		log.Println("Error while connecting to db :", err)
	}
	defer db2.Close()

	db, err := connectDB(adminDB)
	if err != nil {
		log.Println("Error while connecting to db :", err)
	}
	defer db.Close()

	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		fmt.Println("Error while reading request body from trakNPay : ", err)
		http.Error(w, err.Error(), 500)
		return
	}

	params := "https://xyz/?" + string(b)

	url, err := url.Parse(params)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	m := url.Query()
	fmt.Println("Got from tracknpay : ", m)
	quantity, err := strconv.Atoi(m["quantity"][0])
	if err != nil {
		log.Println("Error while converting quantity", err)
	}
	fmt.Println("1")

	if m["response_message"][0] != "Transaction successful" {
		err = EditSeats(db, quantity, m["showid"][0], m["category"][0])
		if err != nil {
			log.Println("Could not add seats back : ", err)
		}
	} else {
		fmt.Println("2")
		p := Purchase{Status: 1, OrderID: m["order_id"][0], Show: Show{HallName: m["description"][0]}}
		fmt.Println("3")
		err := p.Success(db2)
		fmt.Println("4")
		if err != nil {
			log.Println("Could not confirm seats : ", err)
		}
		fmt.Println("5")
	}

}

func GetData(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" { // Change to post
		shows := make([]*Show, 0)
		var s *Show
		db, err := connectDB(adminDB)
		if err != nil {
			log.Println("Error while connecting to db :", err)
		}
		defer db.Close()

		HallList, err := GetHallList(db)

		if err != nil {
			log.Println(err)
		}
		for _, Hall := range HallList {
			fmt.Println(Hall)
			s, err = GetInfo(db, Hall)
			if err != nil || s.HallName == "" {
				fmt.Println(err)
			} else {
				datetime := s.Date + "T" + s.Time + "+05:30"
				t, err := time.Parse(time.RFC3339, datetime)
				if err != nil {
					fmt.Println(err)
				}

				now := time.Now().In(loc)
				now.Add(time.Hour * time.Duration(2))
				if now.Before(t) {
					fmt.Println(s)
					shows = append(shows, s)
				}

			}

		}

		b, err := json.Marshal(shows)
		if err != nil {
			fmt.Println(err)
			fmt.Fprintf(w, fmt.Sprintf("%v", err))
		} else {
			fmt.Fprintf(w, fmt.Sprintf("%v", string(b)))
		}
	}
}

func AddData(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, _ := template.ParseFiles("admin.gtpl")
		t.Execute(w, nil)
	} else {
		if err := r.ParseForm(); err != nil {
			fmt.Println(err)
		}

		var CategoryNames []string
		var CategorySeats []string
		var CategoryPrices []string

		for key, val := range r.PostForm {
			if strings.Contains(key, "categoryname") {
				CategoryNames = append(CategoryNames, val[0])
			} else if strings.Contains(key, "seats") {
				CategorySeats = append(CategorySeats, val[0])
			} else if strings.Contains(key, "price") {
				CategoryPrices = append(CategoryPrices, val[0])
			}

			fmt.Fprintf(w, fmt.Sprintf("%v : %v", key, val))
		}

		hallName := template.HTMLEscapeString(r.PostFormValue("hallname"))
		movieName := template.HTMLEscapeString(r.PostFormValue("moviename"))
		movieDescription := template.HTMLEscapeString(r.PostFormValue("description"))
		movieTrailer := template.HTMLEscapeString(r.PostFormValue("trailer"))
		moviePoster := template.HTMLEscapeString(r.PostFormValue("poster"))
		movieTime := template.HTMLEscapeString(r.PostFormValue("time"))
		movieDate := template.HTMLEscapeString(r.PostFormValue("date"))

		h := &Show{
			HallName: hallName,
			ID:       strconv.FormatInt(time.Now().Unix(), 10),
			Movie: Movie{
				Name:        movieName,
				Description: movieDescription,
				Trailer:     movieTrailer,
				Poster:      moviePoster,
				Time:        movieTime,
				Date:        movieDate,
			},
		}
		db, err := connectDB(adminDB)

		if err != nil {
			log.Println("Error while connecting to db :", err)
		}
		defer db.Close()

		err = h.AddShow(db)
		if err != nil {
			log.Println("Add Show err : ", err)
		}

		c := &Category{}
		for i := 0; i < len(CategoryNames); i++ {
			seats, err := strconv.Atoi(CategorySeats[i])
			if err != nil {
				log.Println("Err : ", err)
			}
			price, err := strconv.Atoi(CategoryPrices[i])
			if err != nil {
				log.Println("Err : ", err)
			}
			c = &Category{Name: string(CategoryNames[i]), Seats: seats, Price: price}
			err = c.AddCategory(db, h.HallName, h.ID)
			if err != nil {
				log.Println("Add Category err : ", err)
			}

		}
	}
}

func main() {
	/*
		CreateCategoryTable()
		CreateHallList()

		CreateHallTable("Imperial Hall")
		CreatePurchaseTable("Imperial Hall")

		CreateHallTable("Denzong Hall")
		CreatePurchaseTable("Denzong Hall")
	*/

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/getdata", GetData)
	http.HandleFunc("/makepayment", MakePayment)
	http.HandleFunc("/paymentresponse", PaymentResponse)
	http.HandleFunc("/adddata", AddData)
	http.HandleFunc("/failure", FailedPayment)
	http.ListenAndServe(port, nil)
}
