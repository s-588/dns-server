package server

import (
	"fmt"
	"log/slog"

	"github.com/miekg/dns"
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