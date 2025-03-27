package sqlite

// This package define a wrapper to transform sqlite ResourceRecord's to
// protocol.ResourceRecord's

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/prionis/dns-server/protocol"
	"github.com/prionis/dns-server/sqlite/query"
)

type DB struct {
	queries *query.Queries
}

func NewDB() (DB, error) {
	conn, err := sql.Open("sqlite", "resourceRecords")
	if err = conn.Ping(); err != nil {
		return DB{}, err
	}

	queries := query.New(conn)
	return DB{
		queries: queries,
	}, nil
}

func (db DB) GetResourceRecord(name string) ([]protocol.ResourceRecord, error) {
	resourceRecords := make([]protocol.ResourceRecord, 0)

	rrs, err := db.queries.GetResourceRecord(context.Background(), name)
	if err != nil {
		return resourceRecords, err
	}

	for _, record := range rrs {

		t, ok := protocol.Types[record.Type]
		if !ok {
			return resourceRecords, fmt.Errorf("uknown type %s", record.Type)
		}

		c, ok := protocol.Classes[record.Class]
		if !ok {
			return resourceRecords, fmt.Errorf("uknown class %s", record.Type)
		}

		resourceRecords = append(resourceRecords, protocol.ResourceRecord{
			Domain:     record.Domain,
			Type:       t,
			Class:      c,
			TimeToLive: uint32(record.Ttl.Int64),
			DataLen:    uint16(len(record.Result)),
			Data:       []byte(record.Result),
		})
	}
	return resourceRecords, nil
}
