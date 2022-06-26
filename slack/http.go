package slack

import (
	"encoding/json"
	"log"
	"net/http"
)

type Server struct {
	tracker *Tracker
}

func NewServer(tracker *Tracker) *Server {
	return &Server{
		tracker: tracker,
	}
}

func (s *Server) HandleReadinessResponse() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {

		if r.Method == "POST" {
			if r.FormValue("payload") != "" {
				log.Println(r.FormValue("payload"))
				r.ParseForm()
				payloadForm := r.FormValue("payload")
				var payload MessageResponse
				if err := json.Unmarshal([]byte(payloadForm), &payload); err != nil {
					log.Println(err)
				}
				if err := s.tracker.Update(payload.CallbackId, payload.Actions[0].Value == "OK"); err != nil {
					log.Println(err)
				}
			}
		}
		rw.Write([]byte("🫡"))
	}
}
