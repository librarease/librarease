package config

// Header constants.
const (
	HEADER_KEY_X_USER_ID = "X-User-Id"
)

const (
	ENV_KEY_APP_ENV = "APP_ENV"
)

type ContextKey uint

const (
	_ ContextKey = iota
	CTX_KEY_USER_ID
	CTX_KEY_USER_ROLE
)
