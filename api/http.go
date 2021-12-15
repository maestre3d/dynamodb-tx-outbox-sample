package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/maestre3d/dynamodb-tx-outbox/controller"
)

func NewHttpApi() *http.Server {
	router := mux.NewRouter()
	mapRoutes(router)
	return &http.Server{
		Addr:    ":8081",
		Handler: router,
	}
}

func mapRoutes(r *mux.Router) {
	ctrl := controller.NewStudentHttp()
	ctrl.MapRoutes(r)
}
