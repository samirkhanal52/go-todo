package route

import (
	"log"
	"net/http"

	"github.com/samirkhanal52/go-todo/middleware"
)

func HandleIndex(res http.ResponseWriter, req *http.Request) {
	reqUrl := "." + req.URL.Path

	log.Println("Fetching..", reqUrl, req.Method)

	if reqUrl == "./" {
		reqUrl = "./static/templates/index.html"

		http.ServeFile(res, req, reqUrl)

	} else if reqUrl == "./todo" {
		switch req.Method {
		case "GET":
			middleware.HandleFetchToDo(res, req)
		case "POST":
			middleware.HandleAddToDo(res, req)
		case "DELETE":
			middleware.HandleDeleteToDo(res, req)
		case "PUT":
			middleware.HandleUpdateToDo(res, req)
		default:
			http.NotFound(res, req)
		}
	} else {
		http.NotFound(res, req)
	}
}
