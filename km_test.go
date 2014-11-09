package main

import "testing"

func TestParseConfig(t *testing.T) {
	if _, err := parseConfig("/afilethatprobablynotexists"); err == nil {
		t.Error("expected error on invalid filename")
	}
	config, err := parseConfig("testdata/config.yml")
	if err != nil {
		t.Errorf("error parsing test data config: %s", err)
	}
	if config.Env != "testing" {
		t.Errorf("envirnment expected to be 'testing', got:%s", config.Env)
	}

	config, err = parseConfig("testdata/invalid_yml.yml")
	if config.Env != "" || config.Log != "" {
		t.Error("invalid yaml should result in empty config struct")
	}
}

func TestMain(t *testing.T) {
	main()
}
