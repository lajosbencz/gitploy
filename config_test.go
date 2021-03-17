package main

import "testing"

func TestConfig(t *testing.T) {
	cfg, err := ConfigFromYamlFile("fixture/config.yaml")
	if err != nil {
		t.Fatal(err.Error())
	}
	var p *ConfigProject
	if p, err = cfg.GetProject("dummy"); err != nil {
		t.Fatal(err.Error())
	}
	if p.Local != "/var/www/gitploy-dummy" {
		t.Fatal("local path mismatch")
	}
}
