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
		m := make(map[string]string)

		m["name"] = template.HTMLEscapeString(r.URL.Query().Get("name"))
		m["email"] = template.HTMLEscapeString(r.URL.Query().Get("email"))
		m["phone"] = template.HTMLEscapeString(r.URL.Query().Get("phone"))
		m["price"] = template.HTMLEscapeString(r.URL.Query().Get("price")) // const that we can set ourselves

		m["movie"] = template.HTMLEscapeString(r.URL.Query().Get("movie")) // *
		m["hall"] = template.HTMLEscapeString(r.URL.Query().Get("hall"))
		m["showid"] = template.HTMLEscapeString(r.URL.Query().Get("showid")) // *
		m["time"] = template.HTMLEscapeString(r.URL.Query().Get("time"))
		m["date"] = template.HTMLEscapeString(r.URL.Query().Get("date"))
		m["quantity"] = template.HTMLEscapeString(r.URL.Query().Get("quantity"))
		m["category"] = template.HTMLEscapeString(r.URL.Query().Get("category"))

		m["orderid"] = template.HTMLEscapeString(r.URL.Query().Get("orderid"))

		quantity, err := strconv.Atoi(m["quantity"])
		if err != nil {
			log.Println("Error while converting quantity", err)
		}

		db, err := connectDB(adminDB)
		if err != nil {
			log.Println("Error while connecting to db : ", err)
		}
		defer db.Close()

		RemSeats, err := LookupRemSeats(db, m["showid"], m["category"])
		if err != nil {
			log.Println("Edit seats err : ", err)
		}
		if RemSeats-quantity < 0 {
			fmt.Fprintf(w, "No seats remaining")
			return
		}

		err = EditSeats(db, -1*quantity, m["showid"], m["category"])
		if err != nil {
			log.Println("Edit seats err : ", err)
		}
		params := GenerateParams(m)
		v := url.Values{}

		for key := range params {
			v.Set(key, params[key])
		}
		resp, err := http.PostForm(gateway, v)
		if err != nil {
			log.Println(err)
		}
		defer resp.Body.Close()
		io.Copy(w, resp.Body)

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
		panic(err)
	}
	fmt.Println(url.Query())

}

func PaymentResponse(w http.ResponseWriter, r *http.Request) {
	/*
	   This is the URL traknpay will callback after payment.
	   If payment is succesful, twilio messages to car provider, customer, etc.
	*/

	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	params := "https://xyz/?" + string(b)

	url, err := url.Parse(params)
	if err != nil {
		panic(err)
	}

	m := url.Query()
	quantity, err := strconv.Atoi(m["quantity"][0])
	if err != nil {
		log.Println("Error while converting quantity", err)
	}

	db, err := connectDB(adminDB)
	if err != nil {
		log.Println("Error while connecting to db :", err)
	}
	defer db.Close()

	if m["response_message"][0] != "Transaction successful" {
		err = EditSeats(db, quantity, m["showid"][0], m["category"][0])
		if err != nil {
			log.Println("Edit seats err : ", err)
		}
	} else {
		amount, err := strconv.Atoi(m["amount"][0])
		if err != nil {
			log.Println("Error while converting price", err)
		}
		moviename, hallname, err := LookupHallMovie(db, m["showid"][0])
		orderid := fmt.Sprintf("%d", time.Now().Unix())
		if err != nil {
			panic(err)
		}

		p := &Purchase{Name: m["name"][0], Phone: m["phone"][0], Email: m["email"][0], BookingTime: m["payment_datetime"][0], Quantity: quantity, OrderID: orderid,
			Show: Show{ID: m["showid"][0], HallName: hallname,
				Movie: Movie{Name: moviename, Time: m["time"][0], Date: m["date"][0]}, Category: Category{Name: m["category"][0]}}}

		db2, err := connectDB(purchaseDB)
		if err != nil {
			log.Println("Error while connecting to db :", err)
		}
		defer db2.Close()

		err = p.AddPurchase(db2, amount)
		if err != nil {
			log.Println(err)
		}

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
		fmt.Println("GOT STUFF")
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
