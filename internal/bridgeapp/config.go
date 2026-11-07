package bridgeapp

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"time"
)

type Config struct {
	OverwatchURL, OverwatchToken, StatePath, ListenAddress string
	OverwatchConnectionID, HerdrConnectionID, AgentName    string
	HerdrBinary, HerdrMachine, WorkingDirectory            string
	AgentCommand                                           []string
	PollInterval, HeartbeatInterval                        time.Duration
	LeaseSeconds, MaxConcurrent                            int
}

func LoadConfig() (Config, error) {
	c := Config{OverwatchURL: os.Getenv("FORMA_BRIDGE_OVERWATCH_URL"), OverwatchToken: os.Getenv("FORMA_BRIDGE_OVERWATCH_TOKEN"), StatePath: value("FORMA_BRIDGE_STATE_PATH", "./forma-bridge-state.json"), ListenAddress: value("FORMA_BRIDGE_LISTEN", "127.0.0.1:8091"), OverwatchConnectionID: value("FORMA_BRIDGE_OVERWATCH_CONNECTION_ID", "overwatch"), HerdrConnectionID: value("FORMA_BRIDGE_HERDR_CONNECTION_ID", "herdr"), AgentName: value("FORMA_BRIDGE_AGENT_NAME", "forma-bridge"), HerdrBinary: value("FORMA_BRIDGE_HERDR_BINARY", "hc"), HerdrMachine: os.Getenv("FORMA_BRIDGE_HERDR_MACHINE"), WorkingDirectory: os.Getenv("FORMA_BRIDGE_WORKING_DIRECTORY"), PollInterval: 15 * time.Second, HeartbeatInterval: 5 * time.Minute, LeaseSeconds: 900, MaxConcurrent: 1}
	if raw := os.Getenv("FORMA_BRIDGE_AGENT_COMMAND_JSON"); json.Unmarshal([]byte(raw), &c.AgentCommand) != nil || len(c.AgentCommand) == 0 {
		return Config{}, errors.New("FORMA_BRIDGE_AGENT_COMMAND_JSON must be a non-empty JSON string array")
	}
	var err error
	if c.PollInterval, err = duration("FORMA_BRIDGE_POLL_INTERVAL", c.PollInterval); err != nil {
		return Config{}, err
	}
	if c.HeartbeatInterval, err = duration("FORMA_BRIDGE_HEARTBEAT_INTERVAL", c.HeartbeatInterval); err != nil {
		return Config{}, err
	}
	if c.LeaseSeconds, err = integer("FORMA_BRIDGE_LEASE_SECONDS", c.LeaseSeconds); err != nil {
		return Config{}, err
	}
	if c.MaxConcurrent, err = integer("FORMA_BRIDGE_MAX_CONCURRENT", c.MaxConcurrent); err != nil {
		return Config{}, err
	}
	if c.OverwatchURL == "" || c.OverwatchToken == "" || c.MaxConcurrent < 1 {
		return Config{}, errors.New("Overwatch URL/token and positive concurrency are required")
	}
	if c.HeartbeatInterval >= time.Duration(c.LeaseSeconds)*time.Second {
		return Config{}, errors.New("heartbeat interval must be shorter than the lease")
	}
	return c, nil
}

func value(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
func duration(key string, fallback time.Duration) (time.Duration, error) {
	if v := os.Getenv(key); v != "" {
		d, e := time.ParseDuration(v)
		if e != nil || d <= 0 {
			return 0, errors.New(key + " must be a positive duration")
		}
		return d, nil
	}
	return fallback, nil
}
func integer(key string, fallback int) (int, error) {
	if v := os.Getenv(key); v != "" {
		n, e := strconv.Atoi(v)
		if e != nil {
			return 0, errors.New(key + " must be an integer")
		}
		return n, nil
	}
	return fallback, nil
}
