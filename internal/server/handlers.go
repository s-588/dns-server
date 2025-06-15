package server

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/miekg/dns"
	"github.com/prionis/dns-server/internal/database"
	"github.com/prionis/dns-server/proto/crud/genproto/crudpb"
	"google.golang.org/protobuf/proto"
)

// dnsHandler it is the tcp/udp handler for dns questions.
func (s Server) dnsHandler(w dns.ResponseWriter, msg *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(msg)
	for _, question := range m.Question {
		answers, err := s.db.FindRecords(context.Background(), question.Name, dns.TypeToString[question.Qtype])
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

// loginHandler handle login requests, accept user credentials, process and add jwt token to the response.
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
		http.Error(w, "Incorrect message format", http.StatusBadRequest)
		return
	}

	user, err := s.db.CheckUserPassword(r.Context(), credentials.Username, credentials.Password)
	if err != nil {
		s.logger.Error("invalid login attempt for user " + credentials.GetUsername() + ": " + err.Error())
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login":      user.Login,
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
		Id:        user.ID,
		Login:     user.Login,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
	}
	b, err := proto.Marshal(u)

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/protobuf")
	w.Write(b)
	s.logger.Info(fmt.Sprintf("Login for %s %s %s(%s) handled",
		user.Role, user.FirstName, user.LastName, user.Login))
}

// registerHandler handle add user requests and return created user.
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
		http.Error(w, "Incorrect message format", http.StatusBadRequest)
		return
	}

	user, err := s.db.AddUser(r.Context(),
		database.User{
			Login:     credentials.Login,
			FirstName: credentials.FirstName,
			LastName:  credentials.LastName,
			Role:      credentials.Role,
		}, credentials.Password)
	if err != nil {
		s.logger.Error("can't register new user: " +
			credentials.GetFirstName() + " " +
			credentials.GetLastName())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	u := &crudpb.User{
		Login:     user.Login,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
	}
	b, err := proto.Marshal(u)

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/protobuf")
	w.Write(b)
	s.logger.Info(fmt.Sprintf("Register new user: %s %s %s(%s)",
		user.Role, user.FirstName, user.LastName, user.Login))
}

// getRecordHandler handle get request for resource records.
func (s Server) getRecordHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		s.logger.Error("id is not specified in the path")
		http.Error(w, "Id of the record is not specified in the path", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		s.logger.Error("can't parse id(" + idStr + ")")
	}

	rr, err := s.db.GetRecord(r.Context(), int32(id))
	if err != nil {
		s.logger.Error("can't get user: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	protoRR := &crudpb.ResourceRecord{
		Id:         rr.ID,
		Domain:     rr.Domain,
		Data:       rr.Data,
		Type:       rr.Type,
		Class:      rr.Class,
		TimeToLive: rr.TTL,
	}
	body, err := proto.Marshal(protoRR)
	if err != nil {
		s.logger.Error("can't marshal resource record message: " + err.Error())
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
	s.logger.Info("GET resource record %s %d %s %s %s",
		rr.Domain, rr.TTL, rr.Class, rr.Type, rr.Data,
	)
}

// getAllRecordsHandler handle get requests for resource records.
func (s Server) getAllRecordsHandler(w http.ResponseWriter, r *http.Request) {
	rrs, err := s.db.GetAllRecords(r.Context())
	if err != nil {
		s.logger.Error("can't get records form databas: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	records := &crudpb.ResourceRecordCollection{}
	for _, rr := range rrs {
		records.Records = append(records.Records, &crudpb.ResourceRecord{
			Id:         rr.ID,
			Domain:     rr.Domain,
			Data:       rr.Data,
			Class:      rr.Class,
			Type:       rr.Type,
			TimeToLive: int32(rr.TTL),
		})
	}

	resp, err := proto.Marshal(records)
	if err != nil {
		s.logger.Error("can't get records form databas: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/protobuf")
	w.Write(resp)
	s.logger.Info("GET all resource records, returned " +
		strconv.FormatInt(int64(len(rrs)), 10) + " records")
}

// getUserHandler handle get user requests and return user with provided ID.
func (s Server) getUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		s.logger.Error("login is not specified in the path")
		http.Error(w, "Login of the user is not specified in the path", http.StatusBadRequest)
		return
	}

	user, err := s.db.GetUser(r.Context(), id)
	if err != nil {
		s.logger.Error("can't get user: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	u := &crudpb.User{
		Id:        user.ID,
		Login:     user.Login,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
	}
	b, err := proto.Marshal(u)

	w.WriteHeader(http.StatusOK)
	w.Write(b)
	s.logger.Info(fmt.Sprintf("GET user %s %s %s(%s)",
		user.Role, user.FirstName, user.LastName, user.Login))
}

// getAllUsersHandler handle get requests and return all users.
func (s Server) getAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := s.db.GetAllUsers(r.Context())
	if err != nil {
		s.logger.Error("can't get records form databas: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	u := &crudpb.UserCollection{}
	for _, user := range users {
		u.Users = append(u.Users, &crudpb.User{
			Id:        user.ID,
			Login:     user.Login,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      user.Role,
		})
	}

	resp, err := proto.Marshal(u)
	if err != nil {
		s.logger.Error("can't get records form databas: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/protobu")
	w.Write(resp)
	s.logger.Info("GET all users, " +
		strconv.FormatInt(int64(len(users)), 10) + " users returned")
}

// websocketHandler handle websocket connection for logs.
func (s Server) websocketHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("can't upgrade connection " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	s.wsConns = append(s.wsConns, conn)
	s.logger.Info("new websocket connection with " + conn.RemoteAddr().String() + "established")
}

// deleteUserHandler handle delete request of users.
func (s Server) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	pathID := r.PathValue("id")
	if pathID == "" {
		s.logger.Error("id not specified in the path")
		http.Error(w, "ID of the user to delete is not specified in the path", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(pathID, 10, 64)
	if err != nil {
		s.logger.Error("can't parse id to delete: " + err.Error())
		http.Error(w, "Incorrect user id", http.StatusBadRequest)
		return
	}

	err = s.db.DeleteUser(r.Context(), int32(id))
	if err != nil {
		s.logger.Error("can't delete user: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User with id " + pathID + "successfull deleted"))
	s.logger.Info(fmt.Sprintf("DELETE user, user with id %d was deleted", id))
}

// patchUserHandler handle requests for updating of the user.
func (s Server) patchUserHandler(w http.ResponseWriter, r *http.Request) {
	user := &crudpb.User{}

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

	err = proto.Unmarshal(body, user)
	if err != nil {
		s.logger.Error("can't unmarshal body from " + r.RemoteAddr)
		http.Error(w, "Incorrect message format", http.StatusBadRequest)
		return
	}

	err = s.db.UpdateUser(r.Context(), database.User{
		ID:        user.Id,
		Login:     user.Login,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
	})
	if err != nil {
		s.logger.Error("can't update user: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	s.logger.Info(fmt.Sprintf("PATCH user, user %d was updated: %s %s %s(%s)",
		user.Id, user.Role, user.FirstName, user.LastName, user.Login))
}

// deleteRRHandler handle delete requests of the resource records.
func (s Server) deleteRRHandler(w http.ResponseWriter, r *http.Request) {
	pathID := r.PathValue("id")
	if pathID == "" {
		s.logger.Error("id not specified in the path")
		http.Error(w, "ID of the resource record to delete is not specified in the path", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(pathID, 10, 32)
	if err != nil {
		s.logger.Error("can't parse id to delete: " + err.Error())
		http.Error(w, "Incorrect id", http.StatusBadRequest)
		return
	}

	err = s.db.DeleteRecord(r.Context(), int32(id))
	if err != nil {
		s.logger.Error("can't delete resource record: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Resource record with id " + pathID + "successfull deleted"))
	slog.Info("DELETE resource record " + pathID)
}

// postRRHandler handle create of resource records requests.
func (s Server) postRRHandler(w http.ResponseWriter, r *http.Request) {
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

	rr := &crudpb.ResourceRecord{}
	err = proto.Unmarshal(body, rr)
	if err != nil {
		s.logger.Error("can't unmarshal body from " + r.RemoteAddr)
		http.Error(w, "Incorrect message format", http.StatusBadRequest)
		return
	}

	newRR, err := s.db.AddRecord(r.Context(),
		database.ResourceRecord{
			Domain: rr.Domain,
			Data:   rr.Data,
			Type:   rr.Type,
			Class:  rr.Class,
			TTL:    rr.TimeToLive,
		})
	if err != nil {
		s.logger.Error("can't add resource record: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	protoRR := &crudpb.ResourceRecord{
		Id:         newRR.ID,
		Domain:     newRR.Domain,
		Data:       newRR.Data,
		Type:       newRR.Type,
		Class:      newRR.Class,
		TimeToLive: newRR.TTL,
	}

	result, err := proto.Marshal(protoRR)

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/protobuf")
	w.Write(result)
	s.logger.Info(fmt.Sprintf("POST resource record: %s %d %s %s %s",
		newRR.Domain, newRR.TTL, newRR.Class, newRR.Type, newRR.Data,
	))
}

// patchRRHandler handle update of the resource record requests.
func (s Server) patchRRHandler(w http.ResponseWriter, r *http.Request) {
	rr := &crudpb.ResourceRecord{}

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

	err = proto.Unmarshal(body, rr)
	if err != nil {
		s.logger.Error("can't unmarshal body from " + r.RemoteAddr)
		http.Error(w, "Incorrect message format", http.StatusBadRequest)
		return
	}

	err = s.db.UpdateRecord(r.Context(),
		database.ResourceRecord{
			ID:     rr.Id,
			Domain: rr.Domain,
			Data:   rr.Data,
			Type:   rr.Type,
			Class:  rr.Class,
			TTL:    rr.TimeToLive,
		})
	if err != nil {
		s.logger.Error("can't update user: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	s.logger.Info(fmt.Sprintf("PATCH resource record, "+
		"resource record with id %d was updated: %s %d %s %s %s",
		rr.Id, rr.Domain, rr.TimeToLive, rr.Class, rr.Type, rr.Data))
}
