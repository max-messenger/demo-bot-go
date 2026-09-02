package rediscli

type KeyFormatter interface {
	FormatKey(key string) string
}
