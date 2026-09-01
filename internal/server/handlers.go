package server

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/prionis/dns-server/internal/database"
	"github.com/prionis/dns-server/internal/dns"
	"github.com/prionis/dns-server/proto/crud/genproto/crudpb"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// dnsHandler method is called by the UDP/TCP listeners and
// provide processing of the raw msg.
func (s Server) dnsHandler(msg []byte) []byte {
	start := time.Now()
	var respCode dns.RCode
	var req dns.Message
	if err := req.UnmarshalBinary(msg); err != nil {
		slog.Error("unmarshal message: %w", "error", err)
		respCode = dns.RCodeFormatError
	}
	resp := dns.Message{
		Header: dns.Header{
			ID: req.Header.ID,
		},
	}
	resp.Header.SetFlag(dns.FlagQR)
	resp.Questions = append(resp.Questions, req.Questions...)
	for _, q := range req.Questions {
		answers, err := s.db.FindRecords(context.Background(),
			q.Name,
			q.Type.String())
		if err != nil {
			slog.Error("can't get resource records from database", "error", err.Error())
			continue
		}
		if len(answers) == 0 {
			slog.Warn("domain not found", "name", q.Name, "type", q.Type)
		}
		for _, a := range answers {
			t, ok := dns.ParseType(a.Type)
			if !ok {
				slog.Error("uknown type", "domain", a.Domain, "type", a.Type)
				continue
			}
			c, ok := dns.ParseClass(a.Class)
			if !ok {
				slog.Error("uknown class", "domain", a.Domain, "class", a.Class)
				continue
			}
			rdata, err := dns.ParseRData(t, a.Data)
			if err != nil {
				slog.Error("can't parse RDATA", "domain", a.Domain, "type", a.Type, "data", a.Data)
				continue
			}
			respData, err := rdata.MarshalBinary()
			if err != nil {
				slog.Error("can't marshal RDATA", "domain", a.Domain, "type", a.Type, "rdata", rdata)
				continue
			}

			rr := dns.RR{
				Name:     a.Domain,
				Type:     t,
				Class:    c,
				TTL:      uint32(a.TTL),
				RDLength: uint16(len(respData)),
				RData:    rdata,
			}
			slog.Info("found answer", "RR", rr)
			req.Answers = append(req.Answers, rr)
		}
		s.metrics.DNSQueriesTotal.WithLabelValues(
			q.Type.String(),
			respCode.String()).Inc()
		s.metrics.DNSRecordsFound.
			WithLabelValues(q.Type.String()).
			Add(float64(len(req.Answers)))
		s.metrics.DNSQueryDuration.
			WithLabelValues(q.Type.String()).
			Observe(time.Since(start).Seconds())
	}

	resp.Header.QDCount = uint16(len(resp.Questions))
	resp.Header.ANCount = uint16(len(resp.Answers))

	out, err := resp.MarshalBinary()
	if err != nil {
		slog.Error("marshal response", "error", err)
		return nil
	}
	return out
}

// loginHandler handle login requests, accept user credentials, process and add jwt token to the response.
func (s Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	credentials := &crudpb.Login{}

	if r.Header.Get("Content-Type") != "application/protobuf" {
		slog.Error("Content-Type header is set to " + r.Header.Get("Content-Type"))
		http.Error(w, "Accept only application/protobuf Content-Type", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("can't read request body from " + r.RemoteAddr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	err = proto.Unmarshal(body, credentials)
	if err != nil {
		slog.Error("can't unmarshal body from " + r.RemoteAddr)
		http.Error(w, "Incorrect message format", http.StatusBadRequest)
		return
	}

	user, err := s.db.CheckUserPassword(r.Context(), credentials.Username, credentials.Password)
	if err != nil {
		slog.Error("invalid login attempt for user " + credentials.GetUsername() + ": " + err.Error())

		var pgErr *pgconn.PgError
		var errStr string
		if errors.As(err, &pgErr) {
			errStr = "User not found"
			s.metrics.LoginAttemptsTotal.WithLabelValues("invalid_user").Inc()
			http.Error(w, errStr, http.StatusForbidden)
			return
		}

		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			errStr = "Incorrect password"
			s.metrics.LoginAttemptsTotal.WithLabelValues("wrong_password").Inc()
			http.Error(w, errStr, http.StatusForbidden)
			return
		}

		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":         user.ID,
		"login":      user.Login,
		"role":       user.Role,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
	})

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		slog.Error("JWT_SECRET environment variable is not set")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		slog.Error("can't sign new JWT token for user " + user.Login + ": " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	cookie := http.Cookie{
		Name:     "jwt",
		Value:    tokenString,
		Path:     "/",
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
	slog.Info(fmt.Sprintf("Login for %s %s %s(%s) handled",
		user.Role, user.FirstName, user.LastName, user.Login))
	s.metrics.LoginAttemptsTotal.WithLabelValues("success").Inc()
}

// registerHandler handle add user requests and return created user.
func (s Server) registerHandler(w http.ResponseWriter, r *http.Request) {
	credentials := &crudpb.Register{}

	if r.Header.Get("Content-Type") != "application/protobuf" {
		slog.Error("Content-Type header is set to " + r.Header.Get("Content-Type"))
		http.Error(w, "Accept only application/protobuf Content-Type", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("can't read request body from " + r.RemoteAddr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	err = proto.Unmarshal(body, credentials)
	if err != nil {
		slog.Error("can't unmarshal body from " + r.RemoteAddr)
		http.Error(w, "Incorrect message format", http.StatusBadRequest)
		return
	}

	id, err := s.db.AddUser(r.Context(),
		database.User{
			Login:     credentials.Login,
			FirstName: credentials.FirstName,
			LastName:  credentials.LastName,
			Role:      credentials.Role,
		}, credentials.Password)
	if err != nil {
		slog.Error("can't register new user " +
			credentials.Login + ": " +
			err.Error())
		var pgErr *pgconn.PgError
		var errStr string
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23502", "23503": // not_null_violation
				errStr = "Uknown role"

			case "23505": // unique_violation
				errStr = "Already exist"

			default:
				errStr = "Can't update user"
			}
		}
		http.Error(w, errStr, http.StatusInternalServerError)
		return
	}

	u := &crudpb.User{
		Id:        id,
		Login:     credentials.Login,
		FirstName: credentials.FirstName,
		LastName:  credentials.LastName,
		Role:      credentials.Role,
	}
	b, err := proto.Marshal(u)

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/protobuf")
	w.Write(b)
	slog.Info(fmt.Sprintf("Register new user: %s %s %s(%s)",
		credentials.Role, credentials.FirstName, credentials.LastName, credentials.Login))
}

// getRecordHandler handle get request for resource records.
func (s Server) getRecordHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		slog.Error("id is not specified in the path")
		http.Error(w, "Id of the record is not specified in the path", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		slog.Error("can't parse id(" + idStr + ")")
	}

	rr, err := s.db.GetRecord(r.Context(), int32(id))
	if err != nil {
		slog.Error("can't get user: " + err.Error())
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
		slog.Error("can't marshal resource record message: " + err.Error())
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
	slog.Info("GET resource record %s %d %s %s %s",
		rr.Domain, rr.TTL, rr.Class, rr.Type, rr.Data,
	)
}

// getAllRecordsHandler handle get requests for resource records.
func (s Server) getAllRecordsHandler(w http.ResponseWriter, r *http.Request) {
	rrs, err := s.db.GetAllRecords(r.Context())
	if err != nil {
		slog.Error("can't get records from databas: " + err.Error())
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
		slog.Error("can't get records form databas: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/protobuf")
	w.Write(resp)
	slog.Info("GET all resource records, returned " +
		strconv.FormatInt(int64(len(rrs)), 10) + " records")
}

// getUserHandler handle get user requests and return user with provided ID.
func (s Server) getUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		slog.Error("login is not specified in the path")
		http.Error(w, "Login of the user is not specified in the path", http.StatusBadRequest)
		return
	}

	user, err := s.db.GetUser(r.Context(), id)
	if err != nil {
		slog.Error("can't get user: " + err.Error())
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
	slog.Info(fmt.Sprintf("GET user %s %s %s(%s)",
		user.Role, user.FirstName, user.LastName, user.Login))
}

// getAllUsersHandler handle get requests and return all users.
func (s Server) getAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := s.db.GetAllUsers(r.Context())
	if err != nil {
		slog.Error("can't get records form databas: " + err.Error())
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
		slog.Error("can't get records from databas: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/protobu")
	w.Write(resp)
	slog.Info("GET all users, " +
		strconv.FormatInt(int64(len(users)), 10) + " users returned")
}

// websocketHandler handle websocket connection for logs.
func (s Server) websocketHandler(ws *WebSocket) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.Error("can't upgrade connection " + err.Error())
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		ws.AddConn(conn)
		slog.Info("new websocket connection with " + conn.RemoteAddr().String() + "established")
	}
}

// deleteUserHandler handle delete request of users.
func (s Server) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	pathID := r.PathValue("id")
	if pathID == "" {
		slog.Error("id not specified in the path")
		http.Error(w, "ID of the user to delete is not specified in the path", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(pathID, 10, 64)
	if err != nil {
		slog.Error("can't parse id to delete: " + err.Error())
		http.Error(w, "Incorrect user id", http.StatusBadRequest)
		return
	}

	err = s.db.DeleteUser(r.Context(), int32(id))
	if err != nil {
		slog.Error("can't delete user: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User with id " + pathID + "successfull deleted"))
	slog.Info(fmt.Sprintf("DELETE user, user with id %d was deleted", id))
}

// patchUserHandler handle requests for updating of the user.
func (s Server) patchUserHandler(w http.ResponseWriter, r *http.Request) {
	user := &crudpb.User{}

	if r.Header.Get("Content-Type") != "application/protobuf" {
		slog.Error("Content-Type header is set to " + r.Header.Get("Content-Type"))
		http.Error(w, "Accept only application/protobuf Content-Type", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("can't read request body from " + r.RemoteAddr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	err = proto.Unmarshal(body, user)
	if err != nil {
		slog.Error("can't unmarshal body from " + r.RemoteAddr)
		http.Error(w, "Incorrect message format", http.StatusBadRequest)
		return
	}

	err = s.db.UpdateUser(r.Context(), database.User{
		ID:        user.Id,
		Login:     user.Login,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
	}, user.Password)
	if err != nil {
		slog.Error("can't update user: " + err.Error())
		var pgErr *pgconn.PgError
		var errStr string
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23502", "23503": // not_null_violation
				errStr = "Uknown role"

			case "23505": // unique_violation
				errStr = "Already exist"

			default:
				errStr = "Can't update user"
			}
		}
		http.Error(w, errStr, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	slog.Info(fmt.Sprintf("PATCH user, user %d was updated: %s %s %s(%s)",
		user.Id, user.Role, user.FirstName, user.LastName, user.Login))
}

// deleteRRHandler handle delete requests of the resource records.
func (s Server) deleteRRHandler(w http.ResponseWriter, r *http.Request) {
	pathID := r.PathValue("id")
	if pathID == "" {
		slog.Error("id not specified in the path")
		http.Error(w, "ID of the resource record to delete is not specified in the path", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(pathID, 10, 32)
	if err != nil {
		slog.Error("can't parse id to delete: " + err.Error())
		http.Error(w, "Incorrect id", http.StatusBadRequest)
		return
	}

	err = s.db.DeleteRecord(r.Context(), int32(id))
	if err != nil {
		slog.Error("can't delete resource record: " + err.Error())
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
		slog.Error("Content-Type header is set to " + r.Header.Get("Content-Type"))
		http.Error(w, "Accept only application/protobuf Content-Type", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("can't read request body from " + r.RemoteAddr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	rr := &crudpb.ResourceRecord{}
	err = proto.Unmarshal(body, rr)
	if err != nil {
		slog.Error("can't unmarshal body from " + r.RemoteAddr)
		http.Error(w, "Incorrect message format", http.StatusBadRequest)
		return
	}

	id, err := s.db.AddRecord(r.Context(),
		database.ResourceRecord{
			Domain: rr.Domain,
			Data:   rr.Data,
			Type:   rr.Type,
			Class:  rr.Class,
			TTL:    rr.TimeToLive,
		})
	if err != nil {
		slog.Error("can't add resource record: " + err.Error())
		var pgErr *pgconn.PgError
		var errStr string
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23502", "23503": // not_null_violation
				errStr = "Uknown type or class"

			case "23505": // unique_violation
				errStr = "Already exist"

			default:
				errStr = "Can't update"
			}
		}
		http.Error(w, errStr, http.StatusInternalServerError)
		return
	}

	protoRR := &crudpb.ResourceRecord{
		Id:         id,
		Domain:     rr.Domain,
		Data:       rr.Data,
		Type:       rr.Type,
		Class:      rr.Class,
		TimeToLive: rr.TimeToLive,
	}

	result, err := proto.Marshal(protoRR)

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/protobuf")
	w.Write(result)
	slog.Info(fmt.Sprintf("POST resource record: %s %d %s %s %s",
		rr.Domain, rr.TimeToLive, rr.Class, rr.Type, rr.Data,
	))
}

// patchRRHandler handle update of the resource record requests.
func (s Server) patchRRHandler(w http.ResponseWriter, r *http.Request) {
	rr := &crudpb.ResourceRecord{}

	if r.Header.Get("Content-Type") != "application/protobuf" {
		slog.Error("Content-Type header is set to " + r.Header.Get("Content-Type"))
		http.Error(w, "Accept only application/protobuf Content-Type", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("can't read request body from " + r.RemoteAddr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	err = proto.Unmarshal(body, rr)
	if err != nil {
		slog.Error("can't unmarshal body from " + r.RemoteAddr)
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
		slog.Error("can't update user: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	slog.Info(fmt.Sprintf("PATCH resource record, "+
		"resource record with id %d was updated: %s %d %s %s %s",
		rr.Id, rr.Domain, rr.TimeToLive, rr.Class, rr.Type, rr.Data))
}

func (s Server) getAllLogsHandler(w http.ResponseWriter, r *http.Request) {
	result := &crudpb.LogCollection{}
	file, err := os.Open("DNSServer.log")
	if err != nil {
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		log := make(map[string]any)
		err := json.Unmarshal([]byte(line), &log)
		if err != nil {
			continue
		}
		t, err := time.Parse(time.RFC3339, log["time"].(string))
		if err != nil {
			t = time.Now()
		}
		result.Logs = append(result.Logs, &crudpb.Log{
			Time:  timestamppb.New(t),
			Level: log["level"].(string),
			Msg:   log["msg"].(string),
		})
	}
	resp, err := proto.Marshal(result)
	if err != nil {
		slog.Error("can't get records from database: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/protobuf")
	w.Write(resp)
	slog.Info("GET all logs, " +
		strconv.FormatInt(int64(len(result.Logs)), 10) + " logs returned")
}
