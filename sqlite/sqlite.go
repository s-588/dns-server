package sqlite

// This package define a wrapper to transform sqlite ResourceRecord's to
// protocol.ResourceRecord's

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"

	_ "github.com/glebarez/go-sqlite"

	"github.com/prionis/dns-server/database"
	"github.com/prionis/dns-server/protocol"
	"github.com/prionis/dns-server/sqlite/query"
)

type DB struct {
	queries *query.Queries
}

func NewDB() (DB, error) {
	conn, err := sql.Open("sqlite", "RRs.db")
	if err != nil {
		return DB{}, err
	}

	if err = conn.Ping(); err != nil {
		return DB{}, err
	}

	queries := query.New(conn)
	return DB{
		queries: queries,
	}, nil
}

func (db DB) AddRR(t, class string, domain string, ip net.IP, ttl int64) error {
	_, err := db.queries.CreateResourceRecord(context.Background(), query.CreateResourceRecordParams{
		Domain: domain,
		Type:   t,
		Class:  class,
		Data:   ip.String(),
		Ttl:    sql.NullInt64{ttl, true},
	})
	if err != nil {
		return err
	}
	return nil
}

func (db DB) DelRR(id int64) (*protocol.RR, error) {
	rr, err := db.queries.DeleteResourceRecord(context.Background(), id)
	if err != nil {
		return nil, err
	}

	rrType, err := db.queries.GetTypeName(context.Background(), rr.Typeid)
	rrClass, err := db.queries.GetClassName(context.Background(), rr.Classid)

	return &protocol.RR{
		Domain:     rr.Domain,
		Type:       protocol.Types[rrType],
		TimeToLive: uint32(rr.Ttl.Int64),
		Class:      protocol.Classes[rrClass],
		DataLen:    4,
		Data:       net.ParseIP(rr.Data),
	}, nil
}

func (db DB) GetAllRRs() ([]*database.DBRR, error) {
	resourceRecords := make([]*database.DBRR, 0)

	rrs, err := db.queries.GetResourceRecordRecursive(context.Background())
	if err != nil {
		return resourceRecords, err
	}

	for _, record := range rrs {

		t, err := db.queries.GetTypeName(context.Background(), record.Typeid)
		if err != nil {
			return resourceRecords, fmt.Errorf("uknown type ID %s in resource record ID %d", record.Typeid, record.ID)
		}

		c, err := db.queries.GetClassName(context.Background(), record.Classid)
		if err != nil {
			return resourceRecords, fmt.Errorf("uknown class ID %s in resource record ID %d", record.Classid, record.ID)
		}

		resourceRecords = append(resourceRecords, &database.DBRR{
			ID: record.ID,
			RR: protocol.RR{
				Domain:     record.Domain,
				Type:       protocol.Types[t],
				Class:      protocol.Classes[c],
				TimeToLive: uint32(record.Ttl.Int64),
				DataLen:    4,
				Data:       net.ParseIP(record.Data).To4(),
			},
		})
	}
	if len(resourceRecords) <= 0 {
		return resourceRecords, errors.New("not found")
	}
	return resourceRecords, nil
}

func (db DB) GetRRs(name string) ([]*protocol.RR, error) {
	resourceRecords := make([]*protocol.RR, 0)

	rrs, err := db.queries.GetResourceRecord(context.Background(), name)
	if err != nil {
		return resourceRecords, err
	}

	for _, record := range rrs {

		t, err := db.queries.GetTypeName(context.Background(), record.Typeid)
		if err != nil {
			return resourceRecords, fmt.Errorf("uknown type ID %s in resource record ID %d", record.Typeid, record.ID)
		}

		c, err := db.queries.GetClassName(context.Background(), record.Classid)
		if err != nil {
			return resourceRecords, fmt.Errorf("uknown class ID %s in resource record ID %d", record.Classid, record.ID)
		}

		rrType, ok := protocol.Types[t]
		if !ok {
			return resourceRecords, fmt.Errorf("uknown type ID %s", record.Typeid)
		}

		rrClass, ok := protocol.Classes[c]
		if !ok {
			return resourceRecords, fmt.Errorf("uknown class %s", record.Typeid)
		}

		resourceRecords = append(resourceRecords, &protocol.RR{
			Domain:     record.Domain,
			Type:       rrType,
			Class:      rrClass,
			TimeToLive: uint32(record.Ttl.Int64),
			DataLen:    4,
			Data:       net.ParseIP(record.Data).To4(),
		})
	}
	if len(resourceRecords) <= 0 {
		return resourceRecords, errors.New("not found")
	}
	return resourceRecords, nil
}
