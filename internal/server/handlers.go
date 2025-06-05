package server

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/miekg/dns"
	"github.com/prionis/dns-server/proto/crud/genproto/crudpb"
	"google.golang.org/protobuf/proto"
)

func (s Server) dnsHandler(w dns.ResponseWriter, msg *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(msg)
	for _, question := range m.Question {
		answers, err := s.db.GetRRs(question.Name, dns.TypeToString[question.Qtype])
		if err != nil {
			slog.Error("can't get resource records from database: " + err.Error())
		}
		if len(answers) == 0 {
			slog.Warn("domain '" + question.Name + "' and type '" +
				dns.TypeToString[question.Qtype] + "' not found")
		}
		for _, answer := range answers {
			rr, err := dns.NewRR(fmt.Sprintf("%s %d %s %s %s",
				answer.Domain,
				answer.TTL,
				answer.Class,
				answer.Type,
				answer.Data,
			))
			if err != nil {
				slog.Error("can't parse resource record from database to answer")
			}
			slog.Info("found answer: " + rr.String())
			m.Answer = append(m.Answer, rr)
		}
	}
	slog.Info(m.String())
	w.WriteMsg(m)
}

func (s Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	credentials := &crudpb.Login{}

	if r.Header.Get("Content-Type") != "application/protobuf" {
		s.logger.Error("Content-Type header is set to " + r.Header.Get("Content-Type"))
		http.Error(w, "Accept only application/protobuf Content-Type", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Error("can't read request body from " + r.RemoteAddr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	err = proto.Unmarshal(body, credentials)
	if err != nil {
		s.logger.Error("can't unmarshal body from " + r.RemoteAddr)
		http.Error(w, "Incorect message format", http.StatusBadRequest)
		return
	}

	err = s.db.CheckUserPassword(r.Context(), credentials.GetUsername(), credentials.GetPassword())
	if err != nil {
		s.logger.Error("invalid login attempt for user " + credentials.GetUsername() + ": " + err.Error())
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := s.db.GetUser(r.Context(), credentials.GetUsername())
	if err != nil {
		s.logger.Error("can't retrive user " + credentials.GetUsername() + " from database: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user":       user.Login,
		"role":       user.Role,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
	})

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		s.logger.Error("JWT_SECRET environment variable is not set")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		s.logger.Error("can't sign new JWT token for user " + user.Login + ": " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	cookie := http.Cookie{
		Name:     "jwt",
		Value:    tokenString,
		HttpOnly: true,
		MaxAge:   14 * 24 * 60 * 60,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &cookie)

	u := &crudpb.User{
		Login:     user.Login,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
	}
	b, err := proto.Marshal(u)

	w.WriteHeader(http.StatusOK)
	w.Write(b)
	s.logger.Info("Login request handled")
}

func (s Server) registerHandler(w http.ResponseWriter, r *http.Request) {
	credentials := &crudpb.Register{}

	if r.Header.Get("Content-Type") != "application/protobuf" {
		s.logger.Error("Content-Type header is set to " + r.Header.Get("Content-Type"))
		http.Error(w, "Accept only application/protobuf Content-Type", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Error("can't read request body from " + r.RemoteAddr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	err = proto.Unmarshal(body, credentials)
	if err != nil {
		s.logger.Error("can't unmarshal body from " + r.RemoteAddr)
		http.Error(w, "Incorect message format", http.StatusBadRequest)
		return
	}

	user, err := s.db.RegisterNewUser(r.Context(),
		credentials.GetLogin(),
		credentials.GetFirstName(),
		credentials.GetLastName(),
		credentials.GetPassword(),
		credentials.GetRole(),
	)
	if err != nil {
		s.logger.Error("can't register new user: " +
			credentials.GetFirstName() + " " +
			credentials.GetLastName())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user":       user.Login,
		"role":       user.Role,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
	})

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		s.logger.Error("JWT_SECRET environment variable is not set")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		s.logger.Error("can't sign new JWT token for user" + user.Login + ": " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	cookie := http.Cookie{
		Name:     "jwt",
		Value:    tokenString,
		HttpOnly: true,
		MaxAge:   14 * 24 * 60 * 60,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &cookie)

	u := &crudpb.User{
		Login:     user.Login,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
	}
	b, err := proto.Marshal(u)

	w.WriteHeader(http.StatusOK)
	w.Write(b)
	s.logger.Info("Login request handled")
}

func (s Server) getResourceRecordsHandler(w http.ResponseWriter, r *http.Request) {
}

func (s Server) getUsersHandler(w http.ResponseWriter, r *http.Request) {
}

func (s Server) websocketHandler(w http.ResponseWriter, r *http.Request) {
}

func (s Server) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
}

func (s Server) patchUserHandler(w http.ResponseWriter, r *http.Request) {
}

func (s Server) deleteRRHandler(w http.ResponseWriter, r *http.Request) {
}

func (s Server) postRRHandler(w http.ResponseWriter, r *http.Request) {
}

func (s Server) patchRRHandler(w http.ResponseWriter, r *http.Request) {
}
