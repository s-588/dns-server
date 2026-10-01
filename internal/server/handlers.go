package server

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"log/slog"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/prionis/dns-server/internal/database"
	"github.com/prionis/dns-server/internal/dns"
	"github.com/prionis/dns-server/proto/genproto/crudpb"
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
			rr, err := parseRR(a)
			if err != nil {
				slog.Error("can't parse resource record", "error", err)
				continue
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

	l := len(req.Questions)
	if l < 0 || l > math.MaxUint16 {
		slog.Error("too many questions, setting to max", "count", l)
		l = math.MaxUint16
	}
	resp.Header.QDCount = uint16(l)

	l = len(req.Answers)
	if l < 0 || l > math.MaxUint16 {
		slog.Error("too many answers, setting to max", "count", l)
		l = math.MaxUint16
	}
	resp.Header.ANCount = uint16(l)

	out, err := resp.MarshalBinary()
	if err != nil {
		slog.Error("marshal response", "error", err)
		return nil
	}
	return out
}

func parseRR(a database.ResourceRecord) (dns.RR, error) {
	t, ok := dns.ParseType(a.Type)
	if !ok {
		slog.Error("uknown type", "domain", a.Domain, "type", a.Type)
		return dns.RR{}, fmt.Errorf("unknown type: %s", a.Type)
	}
	c, ok := dns.ParseClass(a.Class)
	if !ok {
		slog.Error("uknown class", "domain", a.Domain, "class", a.Class)
		return dns.RR{}, fmt.Errorf("unknown class: %s", a.Class)
	}
	rdata, err := dns.ParseRData(t, a.Data)
	if err != nil {
		slog.Error("can't parse RDATA", "domain", a.Domain, "type", a.Type, "data", a.Data)
		return dns.RR{}, fmt.Errorf("can't parse RDATA: %w", err)
	}
	respData, err := rdata.MarshalBinary()
	if err != nil {
		slog.Error("can't marshal RDATA", "domain", a.Domain, "type", a.Type, "rdata", rdata)
		return dns.RR{}, fmt.Errorf("can't marshal RDATA: %w", err)
	}
	l := len(respData)
	if l < 0 || l > math.MaxUint16 {
		slog.Error("length of RDATA is too big", "domain", a.Domain, "type", a.Type, "length", l)
		return dns.RR{}, fmt.Errorf("length of RDATA is too big: %d", l)
	}

	rr := dns.RR{
		Name:     a.Domain,
		Type:     t,
		Class:    c,
		TTL:      a.TTL,
		RDLength: uint16(l),
		RData:    rdata,
	}
	return rr, nil
}

func (s Server) parseLoginRequest(r *http.Request) (*crudpb.Login, error) {
	if r.Header.Get("Content-Type") != "application/protobuf" {
		return nil, fmt.Errorf("unsupported content type")
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("can't read request body: %w", err)
	}
	defer func() {
		err := r.Body.Close()
		if err != nil {
			slog.Error("can't close request body", "addr", r.RemoteAddr)
		}
	}()

	credentials := &crudpb.Login{}
	err = proto.Unmarshal(body, credentials)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshal body: %w", err)
	}

	return credentials, nil
}

func (s Server) authenticate(ctx context.Context, credentials *crudpb.Login) (database.User, error) {
	user, err := s.db.CheckUserPassword(ctx, credentials.Username, credentials.Password)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			s.metrics.LoginAttemptsTotal.WithLabelValues("invalid_user").Inc()
			return database.User{}, fmt.Errorf("user not found")
		}

		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			s.metrics.LoginAttemptsTotal.WithLabelValues("wrong_password").Inc()
			return database.User{}, fmt.Errorf("incorrect password")
		}

		return database.User{}, fmt.Errorf("something went wrong")
	}

	return user, nil
}

// setJWTToken generates a JWT token for the authenticated user and sets it as a cookie in the response.
func (s Server) setJWTToken(w http.ResponseWriter, user database.User) error {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":         user.ID,
		"login":      user.Login,
		"role":       user.Role,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
	})

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return fmt.Errorf("JWT_SECRET environment variable is not set")
	}

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return fmt.Errorf("can't sign new JWT token")
	}

	// gosec is disabled because it gives warning over https being false.
	cookie := http.Cookie{ // #nosec G124
		Name:     "jwt",
		Value:    tokenString,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   14 * 24 * 60 * 60,
		SameSite: http.SameSiteStrictMode,
		Secure:   s.https,
	}
	http.SetCookie(w, &cookie)
	return nil
}

// loginHandler handle login requests, accept user credentials, process and add jwt token to the response.
func (s Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	credentials, err := s.parseLoginRequest(r)
	if err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	user, err := s.authenticate(r.Context(), credentials)
	if err != nil {
		slog.Error("authentication failed", "error", err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	err = s.setJWTToken(w, user)
	if err != nil {
		slog.Error("can't set JWT token", "error", err)
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
	if err != nil {
		slog.Error("can't marshal user message: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/protobuf")
	_, err = w.Write(b)
	if err != nil {
		slog.Error("can't write response for user " + user.Login + ": " + err.Error())
	}
	slog.Info("User login",
		"role", user.Role,
		"firstName", user.FirstName,
		"lastName", user.LastName,
		"login", user.Login)
	s.metrics.LoginAttemptsTotal.WithLabelValues("success").Inc()
}

// registerHandler handle add user requests and return created user.
func (s Server) registerHandler(w http.ResponseWriter, r *http.Request) {
	credentials := &crudpb.Register{}

	if r.Header.Get("Content-Type") != "application/protobuf" {
		slog.Error("Invalid Content-Type header", "value", r.Header.Get("Content-Type"))
		http.Error(w, "Accept only application/protobuf Content-Type", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("can't read request body", "addr", r.RemoteAddr, "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer func() {
		err := r.Body.Close()
		if err != nil {
			slog.Error("can't close request body", "addr", r.RemoteAddr, "error", err)
		}
	}()

	err = proto.Unmarshal(body, credentials)
	if err != nil {
		slog.Error("can't unmarshal body", "addr", r.RemoteAddr, "error", err)
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
		slog.Error("can't register new user", "login", credentials.Login, "error", err)
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
	if err != nil {
		slog.Error("can't marshal user message: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/protobuf")
	_, err = w.Write(b)
	if err != nil {
		slog.Error("can't write response for user", "login", credentials.Login, "error", err)
	}
	slog.Info("Register new user", "role", credentials.Role, "firstName", credentials.FirstName, "lastName", credentials.LastName, "login", credentials.Login)
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
		slog.Error("can't parse id", "id", idStr, "error", err)
	}

	rr, err := s.db.GetRecord(r.Context(), id)
	if err != nil {
		slog.Error("can't get user", "id", id, "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	protoRR := &crudpb.ResourceRecord{
		Id:     rr.ID,
		Domain: rr.Domain,
		Data:   rr.Data,
		Type:   rr.Type,
		Class:  rr.Class,
		Ttl:    rr.TTL,
	}
	body, err := proto.Marshal(protoRR)
	if err != nil {
		slog.Error("can't marshal resource record message", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(body)
	if err != nil {
		slog.Error("can't write response for resource record", "error", err)
	}
	slog.Info("GET resource record",
		"Domain", rr.Domain,
		"TTL", rr.TTL,
		"Class", rr.Class,
		"Type", rr.Type,
		"Data", rr.Data,
	)
}

// getAllRecordsHandler handle get requests for resource records.
func (s Server) getAllRecordsHandler(w http.ResponseWriter, r *http.Request) {
	rrs, err := s.db.GetAllRecords(r.Context())
	if err != nil {
		slog.Error("can't get records from database", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	records := &crudpb.ResourceRecordCollection{}
	for _, rr := range rrs {
		records.Records = append(records.Records, &crudpb.ResourceRecord{
			Id:     rr.ID,
			Domain: rr.Domain,
			Data:   rr.Data,
			Class:  rr.Class,
			Type:   rr.Type,
			Ttl:    rr.TTL,
		})
	}

	resp, err := proto.Marshal(records)
	if err != nil {
		slog.Error("can't get records from database", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/protobuf")
	_, err = w.Write(resp)
	if err != nil {
		slog.Error("can't write response for resource records", "error", err)
	}
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
		slog.Error("can't get user", "id", id, "error", err)
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
	if err != nil {
		slog.Error("can't marshal user message", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(b)
	if err != nil {
		slog.Error("can't write response for user", "login", user.Login, "error", err)
	}
	slog.Info("GET user", "role", user.Role, "firstName", user.FirstName, "lastName", user.LastName, "login", user.Login)
}

// getAllUsersHandler handle get requests and return all users.
func (s Server) getAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := s.db.GetAllUsers(r.Context())
	if err != nil {
		slog.Error("can't get records form databas", "error", err)
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
		slog.Error("can't get records from database", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/protobuf")
	_, err = w.Write(resp)
	if err != nil {
		slog.Error("can't write response for user collection", "error", err)
	}
	slog.Info("GET all users, " +
		strconv.FormatInt(int64(len(users)), 10) + " users returned")
}

// websocketHandler handle websocket connection for logs.
func (s Server) websocketHandler(ws *WebSocket) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.Error("can't upgrade connection", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		ws.AddConn(conn)
		slog.Info("websocket connection established", "remote_addr", conn.RemoteAddr().String())
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

	err = s.db.DeleteUser(r.Context(), id)
	if err != nil {
		slog.Error("can't delete user: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(html.EscapeString("User with id " + pathID + "successful deleted")))
	if err != nil {
		slog.Error("can't write response for user deletion", "error", err)
	}
	slog.Info("Delete user", "id", id)
}

// patchUserHandler handle requests for updating of the user.
func (s Server) patchUserHandler(w http.ResponseWriter, r *http.Request) {
	user := &crudpb.User{}

	if r.Header.Get("Content-Type") != "application/protobuf" {
		slog.Error("Unsupported Content-Type", "content_type", r.Header.Get("Content-Type"))
		http.Error(w, "Accept only application/protobuf Content-Type", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("can't read request body", "remote_addr", r.RemoteAddr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer func() {
		err := r.Body.Close()
		if err != nil {
			slog.Error("can't close request body", "remote_addr", r.RemoteAddr)
		}
	}()

	err = proto.Unmarshal(body, user)
	if err != nil {
		slog.Error("can't unmarshal body", "remote_addr", r.RemoteAddr)
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
		slog.Error("can't update user", "error", err)
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

	err = s.db.DeleteRecord(r.Context(), id)
	if err != nil {
		slog.Error("can't delete resource record: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(html.EscapeString("Resource record with id " + pathID + "successful deleted")))
	if err != nil {
		slog.Error("can't write response for resource record deletion", "error", err.Error())
	}
	slog.Info("Delete resource record", "id", pathID)
}

// postRRHandler handle create of resource records requests.
func (s Server) postRRHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/protobuf" {
		slog.Error("Unsupported Content-Type", "content_type", r.Header.Get("Content-Type"))
		http.Error(w, "Accept only application/protobuf Content-Type", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("can't read request", "addr", r.RemoteAddr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer func() {
		err := r.Body.Close()
		if err != nil {
			slog.Error("can't close request body", "remote_addr", r.RemoteAddr)
		}
	}()

	rr := &crudpb.ResourceRecord{}
	err = proto.Unmarshal(body, rr)
	if err != nil {
		slog.Error("can't unmarshal body", "remote_addr", r.RemoteAddr)
		http.Error(w, "Incorrect message format", http.StatusBadRequest)
		return
	}

	id, err := s.db.AddRecord(r.Context(),
		database.ResourceRecord{
			Domain: rr.Domain,
			Data:   rr.Data,
			Type:   rr.Type,
			Class:  rr.Class,
			TTL:    rr.Ttl,
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
		Id:     id,
		Domain: rr.Domain,
		Data:   rr.Data,
		Type:   rr.Type,
		Class:  rr.Class,
		Ttl:    rr.Ttl,
	}

	result, err := proto.Marshal(protoRR)
	if err != nil {
		slog.Error("can't marshal resource record message: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/protobuf")
	_, err = w.Write(result)
	if err != nil {
		slog.Error("can't write response for resource record creation: " + err.Error())
	}
	slog.Info(fmt.Sprintf("POST resource record: %s %d %s %s %s",
		rr.Domain, rr.Ttl, rr.Class, rr.Type, rr.Data,
	))
}

// patchRRHandler handle update of the resource record requests.
func (s Server) patchRRHandler(w http.ResponseWriter, r *http.Request) {
	rr := &crudpb.ResourceRecord{}

	if r.Header.Get("Content-Type") != "application/protobuf" {
		slog.Error("Unsupported Content-Type", "content_type", r.Header.Get("Content-Type"))
		http.Error(w, "Accept only application/protobuf Content-Type", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("can't read request body", "remote_addr", r.RemoteAddr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer func() {
		err := r.Body.Close()
		if err != nil {
			slog.Error("can't close request body", "remote_addr", r.RemoteAddr)
		}
	}()

	err = proto.Unmarshal(body, rr)
	if err != nil {
		slog.Error("can't unmarshal body", "remote_addr", r.RemoteAddr)
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
			TTL:    rr.Ttl,
		})
	if err != nil {
		slog.Error("can't update user: " + err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	slog.Info(fmt.Sprintf("PATCH resource record, "+
		"resource record with id %d was updated: %s %d %s %s %s",
		rr.Id, rr.Domain, rr.Ttl, rr.Class, rr.Type, rr.Data))
}

func (s Server) getAllLogsHandler(w http.ResponseWriter, r *http.Request) {
	result := &crudpb.LogCollection{}
	file, err := os.Open("DNSServer.log")
	if err != nil {
		slog.Error("can't open log file", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			slog.Error("can't read log file", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		line := scanner.Text()
		log := make(map[string]any)
		err := json.Unmarshal([]byte(line), &log)
		if err != nil {
			slog.Error("can't unmarshal log line", "error", err)
			continue
		}
		s, ok := log["time"].(string)
		if !ok {
			slog.Error("can't get time from log line")
			continue
		}
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			slog.Error("can't parse log time", "error", err)
			t = time.Now()
		}
		l, ok := log["level"].(string)
		if !ok {
			slog.Error("can't get level from log line")
			continue
		}
		m, ok := log["msg"].(string)
		if !ok {
			slog.Error("can't get message from log line")
			continue
		}
		result.Logs = append(result.Logs, &crudpb.Log{
			Time:  timestamppb.New(t),
			Level: l,
			Msg:   m,
		})
	}
	resp, err := proto.Marshal(result)
	if err != nil {
		slog.Error("can't get records from database", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/protobuf")
	_, err = w.Write(resp)
	if err != nil {
		slog.Error("can't write response for logs", "error", err)
	}
	slog.Info("GET all logs, " +
		strconv.FormatInt(int64(len(result.Logs)), 10) + " logs returned")
}
