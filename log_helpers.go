package main

// tokenPrefix returns a short prefix to help trace tokens in logs without exposing full values.
func tokenPrefix(token string) string {
	const n = 8
	if len(token) <= n {
		return token
	}
	return token[:n]
}
