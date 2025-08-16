package gqfsl

import (
	"log"
	"strconv"
	"sync"
	"time"
)

type cron struct {
	config      Config
	emailServer *emailServer
	sql         *sqLite
	stopChan    chan struct{}
	wg          sync.WaitGroup
}

func startCron(config Config, emailServer *emailServer, sql *sqLite) (*cron, error) {
	c := &cron{
		config:      config,
		emailServer: emailServer,
		sql:         sql,
		stopChan:    make(chan struct{}),
	}

	c.wg.Add(1)
	go c.run()

	return c, nil
}

func (c *cron) run() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.config.Cron.DurationRetry)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.processMessages()
		case <-c.stopChan:
			return
		}
	}
}

func (c *cron) Stop() {
	close(c.stopChan)
	c.wg.Wait()
}

func (c *cron) processMessages() {

	messages, err := c.sql.List()
	if err != nil {
		log.Printf("cron: failed to get messages: %v\n", err)
		return
	}

	if c.config.Cron.KeepingTheOrder {
		// Sequential processing
		for _, msg := range messages {
			c.processMessage(msg)
		}
	} else {
		// Parallel processing
		var wg sync.WaitGroup
		for _, msg := range messages {
			wg.Add(1)
			go func(m *Message) {
				defer wg.Done()
				c.processMessage(m)
			}(msg)
		}
		wg.Wait()
	}
}

func (c *cron) processMessage(msg *Message) {

	if msg.Status == nil {
		msg.Status = make(map[string]any)
	}

	// Updating the attempt counter
	count, _ := msg.Status["count_try_send"].(float64)
	msg.Status["count_try_send"] = count + 1

	// Trying to send a message
	err := c.emailServer.Send(*msg)
	if err != nil {
		// Remember the mistake
		msg.Status["error"] = err.Error()
		msg.Status["last_try"] = time.Now().Unix()
		// log.Printf("cron: failed to send message %s: %v\n", msg.ID, err)
	} else {
		// Successful Sending Mark
		msg.Status["time_send"] = time.Now().Unix()
		msg.Status["status"] = "sent"
		delete(msg.Status, "error")
		// log.Printf("cron: successfully sent message %s\n", msg.ID)
	}

	// Updating the message in the database
	if err := c.sql.Update(*msg); err != nil {
		log.Printf("cron: failed to update message %s: %v\n", msg.ID, err)
	}

	// Clearing successfully sent messages
	if c.config.Cron.DurationSaveSuccess > 0 && msg.Status["status"] == "sent" {
		if time.Since(time.Unix(msg.Status["time_send"].(int64), 0)) > c.config.Cron.DurationSaveSuccess {
			id, err := strconv.ParseInt(msg.ID, 10, 64)
			if err != nil {
				log.Printf("id message failed convert, %s", msg.ID)
			}
			if err := c.sql.Delete(id); err != nil {
				log.Printf("cron: failed to delete message %s: %v\n", msg.ID, err)
			}
		}
	}
}
