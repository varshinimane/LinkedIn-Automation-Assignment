package config

// Config holds all tunable settings for the LinkedIn bot.
type Config struct {
	BaseURL  string `yaml:"base_url"`
	Headless bool   `yaml:"headless"`
	Debug    bool   `yaml:"debug"`

	Credentials struct {
		Email    string `yaml:"email"`
		Password string `yaml:"password"`
	} `yaml:"credentials"`

	Browser struct {
		UserAgents []string `yaml:"user_agents"`
		Viewports  []struct {
			Width  int     `yaml:"width"`
			Height int     `yaml:"height"`
			Scale  float64 `yaml:"scale"`
		} `yaml:"viewports"`
		Timezone string `yaml:"timezone"`
		Proxy    string `yaml:"proxy"`
	} `yaml:"browser"`

	Search struct {
		JobTitle        string `yaml:"job_title"`
		Location        string `yaml:"location"`
		Company         string `yaml:"company"`
		PaginationPages int    `yaml:"pagination_pages"`
	} `yaml:"search"`

	Limits struct {
		DailyConnections int `yaml:"daily_connections"`
		DailyMessages    int `yaml:"daily_messages"`
	} `yaml:"limits"`

	Messaging struct {
		Templates         []string `yaml:"templates"`
		FollowupDelayHour int      `yaml:"followup_delay_hour"`
	} `yaml:"messaging"`

	Humanization struct {
		MinDelayMs    int `yaml:"min_delay_ms"`
		MaxDelayMs    int `yaml:"max_delay_ms"`
		JitterMs      int `yaml:"jitter_ms"`
		ThinkTimeMs   int `yaml:"think_time_ms"`
		ScrollPauseMs int `yaml:"scroll_pause_ms"`
		BusinessHours struct {
			Start int `yaml:"start"`
			End   int `yaml:"end"`
		} `yaml:"business_hours"`
	} `yaml:"humanization"`

	Paths struct {
		CookieFile string `yaml:"cookie_file"`
		DBPath     string `yaml:"db_path"`
	} `yaml:"paths"`
}



