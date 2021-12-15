package controller

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/maestre3d/dynamodb-tx-outbox/application/appservice"
	"github.com/maestre3d/dynamodb-tx-outbox/infrastructure/persistence"
)

func init() {
	appservice.StudentRepository = persistence.NewStudentInMemory()
}

func MapStudentHttpRoutes(r *mux.Router) {
	r.Path("/schools/{school_id}/students").Methods(http.MethodPost).HandlerFunc(createStudent)
	r.Path("/schools/{school_id}/students/{student_id}").Methods(http.MethodGet).HandlerFunc(getStudent)
}

func createStudent(w http.ResponseWriter, r *http.Request) {
	studentID := uuid.NewString()
	if err := appservice.RegisterStudent(r.Context(), appservice.RegisterStudentArgs{
		StudentID: studentID,
		Name:      r.PostFormValue("student_name"),
		SchoolID:  mux.Vars(r)["school_id"],
	}); err != nil {
		respondMessageJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondMessageJSON(w, studentID, 200)
}

func getStudent(w http.ResponseWriter, r *http.Request) {
	student, err := appservice.GetStudentByID(r.Context(), mux.Vars(r)["school_id"], mux.Vars(r)["student_id"])
	if err != nil {
		respondMessageJSON(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondStructJSON(w, *student, 200)
}
