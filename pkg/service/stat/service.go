package stat

import (
	"context"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Service interface {
	CollectStat(stat *Stat)
	InitServer()
}

type serviceImpl struct {
	conn     driver.Conn
	logger   *logrus.Entry
	rwMutex  *sync.RWMutex
	statChan chan *Stat
}

func (s serviceImpl) InitServer() {
	if s.conn == nil {
		return
	}
	timer := time.NewTicker(time.Second * 2)
	go func() {
		stats := make([]*Stat, 0)
		for {
			select {
			case st, open := <-s.statChan:
				if open {
					stats = append(stats, st)
					if len(stats) >= 100 {
						_ = s.sendStat(stats)
						stats = make([]*Stat, 0)
					}
				} else {
					_ = s.sendStat(stats)
					for _, stat := range stats {
						Store(stat)
					}
					return
				}
			case <-timer.C:
				_ = s.sendStat(stats)
				for _, stat := range stats {
					Store(stat)
				}
				stats = make([]*Stat, 0)
				break
			}
		}
	}()
}

func (s serviceImpl) sendStat(stats []*Stat) error {
	if s.conn == nil {
		return nil
	}
	ctx := context.Background()
	batch, err := s.conn.PrepareBatch(ctx, "INSERT INTO stat")
	if err != nil {
		return err
	}
	for _, stat := range stats {
		if e := batch.AppendStruct(stat); e != nil {
			err = e
			s.logger.Error(e)
		}
	}
	if err = batch.Send(); err != nil {
		return err
	}
	return batch.Flush()
}

func (s serviceImpl) CollectStat(stat *Stat) {
	if s.conn == nil {
		return
	}
	s.statChan <- stat
}

func NewStatService(conn driver.Conn, logger *logrus.Entry) Service {
	return &serviceImpl{
		conn:     conn,
		logger:   logger.WithField("search_index", "stat_service"),
		rwMutex:  &sync.RWMutex{},
		statChan: make(chan *Stat, 100),
	}
}
