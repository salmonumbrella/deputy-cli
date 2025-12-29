package secrets

import (
	"encoding/json"
	"time"
)

type Credentials struct {
	Token     string    `json:"token"`
	Install   string    `json:"install"`
	Geo       string    `json:"geo"`
	CreatedAt time.Time `json:"created_at"`
}

func (c *Credentials) BaseURL() string {
	return "https://" + c.Install + "." + c.Geo + ".deputy.com/api/v1"
}

func (c *Credentials) BaseURLV2() string {
	return "https://" + c.Install + "." + c.Geo + ".deputy.com/api/v2"
}

func (c *Credentials) Marshal() ([]byte, error) {
	return json.Marshal(c)
}

func UnmarshalCredentials(data []byte) (*Credentials, error) {
	var c Credentials
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
