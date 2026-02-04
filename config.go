package main

import (
    "io/ioutil"
    "os"

    "gopkg.in/yaml.v3"
)

type Config struct {
    TargetAppIn  string   `yaml:"target_app_in"`
    TargetAppOut string   `yaml:"target_app_out"`
    Devices      []string `yaml:"devices"`
}

func LoadConfig(path string) (Config, error) {
    var cfg Config
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return cfg, err
    }
    b, err := ioutil.ReadFile(path)
    if err != nil {
        return cfg, err
    }
    if err := yaml.Unmarshal(b, &cfg); err != nil {
        return cfg, err
    }
    return cfg, nil
}
