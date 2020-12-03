package cron

import (
	"container/list"
	"errors"
	"sync"
	"time"

	"github.com/journey-c/rbtree"
)

const (
	// (unidirectional) initial --> running --> done
	statusInitial = 0
	statusRunning = 1
	statusDone    = 2
)

// CmdCron task execution command
type CmdCron func()

// Cron structure
type Cron struct {
	jobRecord     *rbtree.RbTree
	jobRecordLock sync.RWMutex
	// used to store tasks when the service is not started.
	jobList     *list.List
	jobListLock sync.RWMutex

	stopChan   chan struct{}
	statusLock sync.RWMutex
	status     int32
}

// NewCron new
func NewCron() *Cron {
	c := &Cron{
		jobRecord: rbtree.NewRbTree(
			func(a interface{}, b interface{}) int {
				valueA := a.(time.Time).UnixNano()
				valueB := b.(time.Time).UnixNano()
				if valueA == valueB {
					return 0
				} else if valueA > valueB {
					return 1
				} else {
					return -1
				}
			}, true),
		jobList:  list.New(),
		stopChan: make(chan struct{}),
		status:   statusInitial,
	}
	return c
}

// JobAdd add a task and check whether the task is legal
func (c *Cron) JobAdd(expr string, cmd CmdCron) error {
	c.statusLock.RLock()
	defer c.statusLock.RUnlock()

	if c.status == statusDone {
		return errors.New("service has been terminated")
	}

	cmdName, err := obtainFuncName(cmd)
	if err != nil {
		return errors.New("invalid cmd: " + err.Error())
	}

	j, err := obtainJob(expr, cmdName, cmd)
	if err != nil {
		return err
	}

	switch c.status {
	case statusInitial:
		c.jobListLock.Lock()
		c.jobList.PushBack(j)
		c.jobListLock.Unlock()
		return nil
	case statusRunning:
		c.jobRecordLock.Lock()
		j.updateNextTime()
		if err = c.jobAdd(j); err != nil {
			return err
		}
		c.jobRecordLock.Unlock()
		return nil
	default:
		return errors.New("invalid status")
	}
}

// thread unsafe
func (c *Cron) jobAdd(j *job) error {
	var jobList *list.List
	jobNodeList := c.jobRecord.Find(j.nextTime)
	// unique
	switch len(jobNodeList) {
	case 1:
		jobList = jobNodeList[0].V.(*list.List)
	case 0:
		jobList = list.New()
		c.jobRecord.Insert(j.nextTime, jobList)
	default:
		return errors.New("jobRecord error")
	}
	jobList.PushBack(j)
	return nil
}

// JobRemove delete tasks through expr and cmd
func (c *Cron) JobRemove(expr string, cmd func()) error {
	c.statusLock.RLock()
	defer c.statusLock.RUnlock()

	cmdName, err := obtainFuncName(cmd)
	if err != nil {
		return errors.New("invalid cmd: " + err.Error())
	}

	switch c.status {
	case statusInitial:
		c.jobListLock.Lock()
		for item := c.jobList.Front(); item != c.jobList.Back(); item = item.Next() {
			j := item.Value.(*job)
			if j.expr == expr && j.cmdName == cmdName {
				_ = c.jobList.Remove(item)
				break
			}
		}
		c.jobListLock.Unlock()
		return nil
	case statusRunning:
		j, err := obtainJob(expr, cmdName, cmd)
		if err != nil {
			return err
		}

		c.jobRecordLock.Lock()
		defer c.jobRecordLock.Unlock()
		j.updateNextTime()
		var jobList *list.List
		jobNodeList := c.jobRecord.Find(j.nextTime)
		switch len(jobNodeList) {
		case 0:
			return nil
		case 1:
			jobList = jobNodeList[0].V.(*list.List)
			var deleteJob *list.Element
			for item := jobList.Front(); item != nil; item = item.Next() {
				if item.Value.(*job).cmdName == cmdName {
					deleteJob = item
					break
				}
			}
			if deleteJob != nil {
				jobList.Remove(deleteJob)
				if jobList.Len() <= 0 {
					c.jobRecord.DeleteByKey(j.nextTime)
				}
			}
		default:
			return errors.New("jobRecord error")
		}
		return nil
	default:
		return errors.New("invalid status")
	}
}

// Start start
func (c *Cron) Start() {
	c.statusLock.Lock()
	defer c.statusLock.Unlock()

	// if cron has already started, do nothing
	if c.status != statusInitial {
		return
	}
	c.status = statusRunning

	c.jobListLock.Lock()
	for c.jobList.Len() != 0 {
		c.jobRecordLock.Lock()
		j := c.jobList.Remove(c.jobList.Front()).(*job) // can't be nil
		j.updateNextTime()
		c.jobAdd(j)
		c.jobRecordLock.Unlock()
	}
	c.jobList = nil // clean job list
	c.jobListLock.Unlock()

	go c.schedule()

}

func (c *Cron) schedule() {
	for {
		c.jobRecordLock.RLock()
		nextNode := c.jobRecord.First()
		oldKey := nextNode.K
		c.jobRecordLock.RUnlock()
		select {
		case <-time.After(nextNode.K.(time.Time).Sub(time.Now())):
			c.jobRecordLock.Lock()
			for item := nextNode.V.(*list.List).Front(); item != nil; item = item.Next() {
				if item.Value == nil {
					continue
				}
				j := item.Value.(*job)
				if j.cmd == nil {
					continue
				}
				go item.Value.(*job).cmd()
				j.updateNextTime()
				_ = c.jobAdd(j)
			}
			c.jobRecord.DeleteByKey(oldKey)
			c.jobRecordLock.Unlock()
		case <-c.stopChan:
			c.jobRecordLock.Lock()
			c.jobRecord = nil
			c.jobRecordLock.Unlock()
			return
		}
	}
}

// Stop stop
func (c *Cron) Stop() {
	c.statusLock.Lock()
	defer c.statusLock.Unlock()

	switch c.status {
	case statusDone:
		return
	case statusRunning:
		c.stopChan <- struct{}{}
		c.status = statusDone
		return
	case statusInitial:
		c.jobListLock.Lock()
		c.jobList = nil
		c.jobListLock.Unlock()
		c.status = statusDone
		return
	default:
	}
}
