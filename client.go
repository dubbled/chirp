package chirp

// Write writes p to a Client's writer
func (c *Client) Write(p []byte) error {
	if c.Writer == nil {
		return ErrNoWriter
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	_, err := c.Writer.Write(p)
	if err != nil {
		c.errors = append(c.errors, err)
	}
	return err
}

func (c *Client) SetID(id string) *Client {
	c.id = id
	return c
}
