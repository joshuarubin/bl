package client

// Email response from bitly api
type Email struct {
	IsPrimary  bool   `json:"is_primary"`
	IsVerified bool   `json:"is_verified"`
	Email      string `json:"email"`
}

// User response from bitly api
type User struct {
	DefaultGroupGUID string  `json:"default_group_guid"`
	Name             string  `json:"name"`
	Created          string  `json:"created"`
	IsActive         bool    `json:"is_active"`
	Modified         string  `json:"modified"`
	IsSSOUser        bool    `json:"is_sso_user"`
	Is2faEnabled     bool    `json:"is_2fa_enabled"`
	Login            string  `json:"login"`
	Emails           []Email `json:"emails"`
}

// Links response from bitly api
type Links struct {
	Pagination Pagination `json:"pagination"`
	Links      []Link     `json:"links"`
}

// Pagination response from bitly api
type Pagination struct {
	Total int    `json:"total"`
	Size  int    `json:"size"`
	Prev  string `json:"prev"`
	Page  int    `json:"page"`
	Next  string `json:"next"`
}

// Deeplink response from bitly api
type Deeplink struct {
	Bitlink     string `json:"bitlink"`
	InstallURL  string `json:"install_url"`
	Created     string `json:"created"`
	AppURIPath  string `json:"app_uri_path"`
	Modified    string `json:"modified"`
	InstallType string `json:"install_type"`
	AppGUID     string `json:"app_guid"`
	GUID        string `json:"guid"`
	OS          string `json:"os"`
}

// Link response from bitly api
type Link struct {
	References     map[string]string `json:"references"`
	Archived       bool              `json:"archived"`
	Tags           []string          `json:"tags"`
	CreatedAt      string            `json:"created_at"`
	Title          string            `json:"title"`
	Deeplinks      []Deeplink        `json:"deeplinks"`
	CreatedBy      string            `json:"created_by"`
	LongURL        string            `json:"long_url"`
	ClientID       string            `json:"client_id"`
	CustomBitlinks []string          `json:"custom_bitlinks"`
	Link           string            `json:"link"`
	ID             string            `json:"id"`
}

// Metric response from bitly api
type Metric struct {
	Clicks int    `json:"clicks"`
	Value  string `json:"value"`
}

// Metrics response from bitly api
type Metrics struct {
	Units         int      `json:"units"`
	Facet         string   `json:"facet"`
	UnitReference string   `json:"unit_reference"`
	Unit          string   `json:"unit"`
	Metrics       []Metric `json:"metrics"`
}
