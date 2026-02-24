package node

import (
	"encoding/json"

	"github.com/evilCYH/NodeHub/internal/utils/generic"
	"github.com/cespare/xxhash/v2"
)

const (
	Alive     uint64 = 1 << 0
	Country   uint64 = 1 << 1
	TikTok    uint64 = 1 << 2
	TikTokIDC uint64 = 1 << 3
)

type Data struct {
	Base
	Info *Info
}

type Base struct {
	Raw       []byte `json:"-"`
	SubId     uint16 `json:"sub_id"`
	UniqueKey uint64 `json:"unique_key"`
}

type InitStatus string

const (
	InitUnknown InitStatus = "unknown"
	InitPassed  InitStatus = "passed"
	InitFailed  InitStatus = "failed"
)

type UniqueKey struct {
	Server     string `yaml:"server"`
	Servername string `yaml:"servername"`
	Port       string `yaml:"port"`
	Type       string `yaml:"type"`
	Uuid       string `yaml:"uuid"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
}

type Info struct {
	SpeedUp     generic.Queue[uint32] `json:"-"`
	SpeedDown   generic.Queue[uint32] `json:"-"`
	Delay       generic.Queue[uint16] `json:"-"`
	Risk        uint8                 `json:"risk"`
	AliveStatus uint64                `json:"alive_status"`
	IP          uint32                `json:"ip"`
	Country     string                `json:"country"`
}

type SimpleInfo struct {
	SpeedUp   uint32 `json:"speed_up"`
	SpeedDown uint32 `json:"speed_down"`
	Delay     uint16 `json:"delay"`
	Risk      uint8  `json:"risk"`
	Count     uint32 `json:"count"`
}

type Filter struct {
	SubId         []uint16 `json:"sub_id"`
	SubIdExclude  bool     `json:"sub_id_exclude"`
	SpeedUpMore   uint32   `json:"speed_up_more"`
	SpeedDownMore uint32   `json:"speed_down_more"`
	Country       []string `json:"country"`
	CountryExclude bool    `json:"country_exclude"`
	DelayLessThan uint16   `json:"delay_less_than"`
	AliveStatus   uint64   `json:"alive_status"`
	RiskLessThan  uint8    `json:"risk_less_than"`
}

func (i *Info) SetAliveStatus(AliveStatus uint64, status bool) {
	if status {
		i.AliveStatus |= AliveStatus
	} else {
		i.AliveStatus &= ^AliveStatus
	}
}

func (u *UniqueKey) Gen() uint64 {
	bytes, _ := json.Marshal(u)
	return xxhash.Sum64(bytes)
}
