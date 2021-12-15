package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type StudentHttp struct {
	// add deps here
}

func NewStudentHttp() *StudentHttp {
	return &StudentHttp{}
}

var _ Http = StudentHttp{}

func (s StudentHttp) MapRoutes(r *mux.Router) {
	r.Path("/schools/{school_id}/students").Methods(http.MethodPost).HandlerFunc(s.create)
	r.Path("/schools/{school_id}/students/{student_id}").Methods(http.MethodGet).HandlerFunc(s.get)
}

func (s StudentHttp) create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		Message string `json:"message"`
	}{
		Message: fmt.Sprintf("Hello mister from POST (school: %s)",
			mux.Vars(r)["school_id"]),
	})
}

func (s StudentHttp) get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		Message string `json:"message"`
	}{
		Message: fmt.Sprintf("Hello mister from GET (school: %s, student: %s)",
			mux.Vars(r)["school_id"], mux.Vars(r)["student_id"]),
	})
}
