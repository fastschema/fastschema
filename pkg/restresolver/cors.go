package restresolver

func MiddlewareCors(c *Context) error {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	if c.Method() == "OPTIONS" {
		c.Status(200)
		return nil
	}

	return c.Next()
}
