package getcached

// Picker picks a proxy from a list according
// to the current requested origin.
type Picker interface {
	Pick(origin string) string
	Set(proxies ...string)
}
