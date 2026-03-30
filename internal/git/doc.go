// Package git implements read-only git subprocess calls with allowlisted arguments only.
//
// All commands use execx (no shell) and EnvGitReadOnly; user input is never forwarded as flags.
package git
