package config

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/liaodiansm/mtg/mtglib"
)

type Optional struct {
	Enabled TypeBool `json:"enabled"`
}

type ListConfig struct {
	Optional

	DownloadConcurrency TypeConcurrency `json:"downloadConcurrency"`
	UpdateEach          TypeDuration    `json:"updateEach"`
}

type Config struct {
	Debug                    TypeBool        `json:"debug"`
	AllowFallbackOnUnknownDC TypeBool        `json:"allowFallbackOnUnknownDc"`
	Secret                   mtglib.Secret   `json:"secret"`
	BindTo                   TypeHostPort    `json:"bindTo"`
	PreferIP                 TypePreferIP    `json:"preferIp"`
	DomainFrontingPort       TypePort        `json:"domainFrontingPort"`
	TolerateTimeSkewness     TypeDuration    `json:"tolerateTimeSkewness"`
	Concurrency              TypeConcurrency `json:"concurrency"`
	Defense                  struct {
		AntiReplay struct {
			Optional

			MaxSize   TypeBytes     `json:"maxSize"`
			ErrorRate TypeErrorRate `json:"errorRate"`
		} `json:"antiReplay"`
	} `json:"defense"`
	Network struct {
		Timeout struct {
			TCP  TypeDuration `json:"tcp"`
			HTTP TypeDuration `json:"http"`
			Idle TypeDuration `json:"idle"`
		} `json:"timeout"`
		DOHIP   TypeIP         `json:"dohIp"`
		Proxies []TypeProxyURL `json:"proxies"`
	} `json:"network"`
}

func (c *Config) Validate() error {
	if !c.Secret.Valid() {
		return fmt.Errorf("invalid secret %s", c.Secret.String())
	}

	if c.BindTo.Get("") == "" {
		return fmt.Errorf("incorrect bind-to parameter %s", c.BindTo.String())
	}

	return nil
}

func (c *Config) String() string {
	buf := &bytes.Buffer{}
	encoder := json.NewEncoder(buf)

	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(c); err != nil {
		panic(err)
	}

	return buf.String()
}
