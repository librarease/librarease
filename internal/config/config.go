package config

// Header constants.
const (
	HEADER_KEY_X_USER_ID = "X-User-Id"
)

const (
	ENV_KEY_APP_ENV = "APP_ENV"
)

type ContextKey string

const (
	CTX_KEY_USER_ID ContextKey = "user_id"
	CTX_KEY_FB_UID  ContextKey = "fb_uid"
)
