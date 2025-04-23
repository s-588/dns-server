package server

import "github.com/prionis/dns-server/protocol"

type Logger interface {
	Info(msg string, keysAndValues ...any)
	Error(msg string, keysAndValues ...any)
	Debug(msg string, keysAndValues ...any)
}

func rrSliceToLog(rrs []*protocol.RR) []protocol.RR {
	var result []protocol.RR
	for _, rr := range rrs {
		if rr != nil {
			result = append(result, *rr)
		}
	}
	return result
}
