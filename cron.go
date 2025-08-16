package gqfsl

import (
	"log"
	"strconv"
	"sync"
	"time"
)

type cron struct {
	config      Config
	emailSender emailSender
	store       messageStore
	stopChan    chan struct{}
	wg          sync.WaitGroup
}

func startCron(config Config, emailSender emailSender, store messageStore) (*cron, error) {
	c := &cron{
		config:      config,
		emailSender: emailSender,
		store:       store,
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

	messages, err := c.store.List()
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

// cron.go

func (c *cron) processMessage(msg *Message) {
	if msg.Status == nil {
		msg.Status = make(map[string]any)
	}

	// Updating the attempt counter
	if status, ok := msg.Status["status"]; ok && status == "sent" {
		c.checkForDeletion(msg)
		return
	}

	// Обработка неотправленных сообщений
	count, _ := msg.Status["count_try_send"].(float64)
	msg.Status["count_try_send"] = count + 1
	// Trying to send a message
	err := c.emailSender.Send(*msg)
	if err != nil {
		msg.Status["error"] = err.Error()
		msg.Status["last_try"] = time.Now().Unix()
	} else {
		msg.Status["time_send"] = time.Now().Unix()
		msg.Status["status"] = "sent"
		delete(msg.Status, "error")
	}

	if err := c.store.Update(*msg); err != nil {
		log.Printf("cron: failed to update message %s: %v\n", msg.ID, err)
	}

	// Clearing successfully sent messages
	if msg.Status["status"] == "sent" {
		c.checkForDeletion(msg)
	}
}

func (c *cron) checkForDeletion(msg *Message) {
	if c.config.Cron.DurationSaveSuccess > 0 {
		if timeSend, ok := msg.Status["time_send"].(int64); ok {
			if time.Since(time.Unix(timeSend, 0)) > c.config.Cron.DurationSaveSuccess {
				id, err := strconv.ParseInt(msg.ID, 10, 64)
				if err != nil {
					log.Printf("id message failed convert, %s", msg.ID)
					return
				}
				if err := c.store.Delete(id); err != nil {
					log.Printf("cron: failed to delete message %s: %v\n", msg.ID, err)
				}
			}
		}
	}
}
