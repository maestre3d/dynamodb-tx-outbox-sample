package controller

import "github.com/gorilla/mux"

type Http interface {
	MapRoutes(r *mux.Router)
}
