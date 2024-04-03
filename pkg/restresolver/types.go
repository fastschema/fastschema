package restresolver

import (
	"time"
)

type Handler func(c *Context) error

type Cookie struct {
	Name        string    `json:"name"`
	Value       string    `json:"value"`
	Path        string    `json:"path"`
	Domain      string    `json:"domain"`
	MaxAge      int       `json:"max_age"`
	Expires     time.Time `json:"expires"`
	Secure      bool      `json:"secure"`
	HTTPOnly    bool      `json:"http_only"`
	SameSite    string    `json:"same_site"`
	SessionOnly bool      `json:"session_only"`
}

type StaticConfig struct {
	Compress      bool          `json:"compress"`
	ByteRange     bool          `json:"byte_range"`
	Browse        bool          `json:"browse"`
	Download      bool          `json:"download"`
	Index         string        `json:"index"`
	CacheDuration time.Duration `json:"cache_duration"` // Default value 10 * time.Second.
	MaxAge        int           `json:"max_age"`        // Default value 0
}
