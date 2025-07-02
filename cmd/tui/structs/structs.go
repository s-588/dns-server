package structs

import "time"

type RR struct {
	ID     int32
	Domain string
	Data   string
	Type   string
	Class  string
	TTL    int32
}
type User struct {
	ID        int32
	Login     string
	FirstName string
	LastName  string
	Role      string
	Password  string
}
type Log struct {
	Time  time.Time
	Level string
	Msg   string
}
