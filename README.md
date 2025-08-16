# GQFSL

guaranteed queue for sending letters

# Introduction

GQFSL - implements a sending system that guarantees that if a letter cannot be sent, it will use a storage base and, with the help of an event scheduler, will be guaranteed to be sent or stored in a temporary base according to a limit for further processing.

# Dependencies

- email send - [gomail](https://github.com/go-gomail/gomail)
- sqlite - ["modernc.org/sqlite"](https://pkg.go.dev/modernc.org/sqlite)

# License

- MIT

# Example

```

gqfsl.New(config gqfsl.Config) (*gqfsl.Queue, error)


type Gqfsl interface {
    // Add delayed dispatch
    Add(message Message) error
    // Send immediate dispatch
    Send(message Message) error
    // Get  get a specific message
    Get(ID int64) (Message, error)
    // List get list of messages
    List() ([]Message, error)
    // Delete delete a specific message
    Delete(ID int64) error
}

```
