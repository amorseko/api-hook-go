package main

import (
	"api-hook/app"
	"api-hook/connection"
	"api-hook/controller"
	"api-hook/models"
	"api-hook/utils"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"regexp"
	"strings"
	"time"
)

func main() {
	router := mux.NewRouter()
	//ctx := context.Background()

	router.HandleFunc("/api/v1/extservice", controller.ExternalService).Methods("POST", "OPTIONS")

	router.Use(app.CorsMiddleware)

	port := os.Getenv("port")

	if port == "" {
		port = "6001" //localhost
	}

	fmt.Print("apps run on port : " + port)

	s := &http.Server{Addr: ":" + port, Handler: logHandler(router), ReadTimeout: 60 * time.Second, WriteTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 5}
	err := s.ListenAndServe() //Launch the app, visit localhost:8000/api
	if err != nil {
		fmt.Print(err)
	}

}

func logHandler(fn http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		x, err := httputil.DumpRequest(r, true)

		if err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusBadRequest)
			return
		}

		req := strings.ReplaceAll(fmt.Sprintf("%s", x), " ", "")
		words := strings.Fields(req)
		var reqMeta []interface{}

		for _, v := range words {
			if !strings.HasPrefix(v, "{") {
				req = strings.ReplaceAll(req, v, "")
				reqMeta = append(reqMeta, v)
			} else {
				break
			}
		}

		m1 := regexp.MustCompile("\r?\n")
		rep := m1.ReplaceAllString(req, "")

		in := []byte(rep)
		var data map[string]interface{}
		if err := json.Unmarshal(in, &data); err != nil {
			data = nil
		}

		log.Println("===============> REQUEST")
		fmt.Println(string(in))
		rec := httptest.NewRecorder()
		fn.ServeHTTP(rec, r)
		fmt.Println()

		//fmt.Println("===============> RESPONSE")
		//fmt.Print(rec.Body)
		//fmt.Println("===============> END RESPONSE : " + r.RequestURI)
		//fmt.Println()

		var dataRes interface{}
		errRes := json.Unmarshal(rec.Body.Bytes(), &dataRes)

		if errRes == nil {
			_, er := connection.GetDBMongo().Collection("access_logs").InsertOne(context.Background(), models.Logs{Endpoint: r.RequestURI, Request: data, Response: dataRes, Meta: reqMeta, LogAt: utils.GetDateNow()})
			if er != nil {
				log.Print("LOG INSERT :" + er.Error())
			}
		}

		// this copies the recorded response to the response writer
		for k, v := range rec.Header() {
			w.Header()[k] = v
		}
		w.WriteHeader(rec.Code)
		_, _ = rec.Body.WriteTo(w)

	}
}
