package config

// Header constants.
const (
	HEADER_KEY_X_UID       = "X-Uid"
	HEADER_KEY_X_CLIENT_ID = "X-Client-Id"
)

const (
	ENV_KEY_APP_ENV        = "APP_ENV"
	ENV_KEY_CLIENT_ID      = "CLIENT_ID"
	ENV_KEY_SESSION_COOKIE = "SESSION_COOKIE_NAME"
	ENV_KEY_S3_BUCKET      = "AWS_S3_BUCKET_NAME"
	ENV_KEY_S3_TEMP_PATH   = "AWS_S3_BUCKET_TEMP_PATH"
)

type ContextKey uint

const (
	_ ContextKey = iota
	CTX_KEY_USER_ID
	CTX_KEY_USER_ROLE
)

const (
	PRESIGN_URL_EXPIRE_MINUTES = 15
)
