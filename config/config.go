package config

import (
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/ngaut/log"
	"io/ioutil"
	"os"
	"strconv"
)

type ConfStruct struct {
	Servers  []string
	Bind     string
	Port     int
	HostName string
}

var Config ConfStruct

func (c *ConfStruct) GetBindAddr() string {
	return JoinAddrPort(c.Bind, c.Port)
}

func (c *ConfStruct) GetHostAddr() string {
	return JoinAddrPort(c.HostName, c.Port)
}

func JoinAddrPort(addr string, port int) string {
	return addr + ":" + strconv.Itoa(port)
}

var path = flag.String("config", "config.toml",
	"")

func init() {
	flag.Parse()
	log.Debug("flags is:", *path)

	fd, err := os.Open(*path)
	defer func() {
		err := fd.Close()
		if err != nil {
			log.Fatal("failed to close file", err)
		}
	}()
	if err != nil {
		log.Fatal("can't read config file:", err)
	}
	contents, err := ioutil.ReadAll(fd)
	if err != nil {
		log.Fatal("can't read config file:", err)
	}
	_, err = toml.Decode(string(contents), &Config)
	if err != nil {
		log.Fatal("toml decode error:", err)
		return
	}
	log.Infof("config: %+v", Config)
}
