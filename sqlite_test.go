package gqfsl

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteCRUD(t *testing.T) {
	// Setup in-memory database
	cfg := Config{
		Sql: SqlConf{
			BaseName: "queue",
			Timeout:  1 * time.Second,
			Path:     "./",
		},
	}
	db, err := startSQL(cfg)
	require.NoError(t, err)
	require.NotNil(t, db)

	// Test message with minimal required fields
	msg := Message{
		From:     "from@example.com",
		To:       "to@example.com",
		Subject:  "Test Subject",
		BodyType: "text/plain",
		Body:     "Test body",
	}

	t.Run("Add message", func(t *testing.T) {
		err := db.Add(msg)
		assert.NoError(t, err)
	})

	t.Run("Get message", func(t *testing.T) {
		// First add the message to ensure it exists
		err := db.Add(msg)
		assert.NoError(t, err)

		// Now try to get it
		result, err := db.Get(1) // ID should be 1 for first message
		assert.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, msg.From, result.From)
		assert.Equal(t, msg.To, result.To)
		assert.Equal(t, msg.Subject, result.Subject)
		assert.Equal(t, msg.BodyType, result.BodyType)
		assert.Equal(t, msg.Body, result.Body)
	})

	t.Run("Update message", func(t *testing.T) {
		// First add the message
		err := db.Add(msg)
		assert.NoError(t, err)

		// Prepare update
		updateMsg := Message{
			ID:       "1",
			From:     "from@example.com",
			To:       "to@example.com",
			Subject:  "Test Subject",
			BodyType: "text/plain",
			Body:     "Test body",
			Status: map[string]any{
				"count_try_send": float64(1),
				"status":         "updated",
			},
		}

		err = db.Update(updateMsg)
		assert.NoError(t, err)

		// Verify update
		updated, err := db.Get(1)
		assert.NoError(t, err)
		require.NotNil(t, updated)

		assert.Equal(t, float64(1), updated.Status["count_try_send"])
		assert.Equal(t, "updated", updated.Status["status"])
	})

	t.Run("Delete message", func(t *testing.T) {
		// First add the message
		err := db.Add(msg)
		assert.NoError(t, err)

		// Delete it
		err = db.Delete(1)
		assert.NoError(t, err)

		// Verify deletion
		result, err := db.Get(1)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	err = db.Close()
	if err != nil {
		log.Fatal(err)
	}

	err = os.Remove("queue.sqlite")
	if err != nil {
		log.Fatal(err)
	}
}
