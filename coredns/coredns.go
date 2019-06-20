package coredns

import (
	"flag"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/rancher/rdns-server/coredns/plugin"
	"github.com/rancher/rdns-server/coredns/plugin/rdns"

	"github.com/coredns/coredns/core/dnsserver"
	"github.com/mholt/caddy"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	// Plug in CoreDNS
	_ "github.com/coredns/coredns/core/plugin"
)

const (
	CoreVersion = "1.5.0"
	CoreType    = "dns"
	CoreName    = "CoreDNS"
	CoreFile    = "Corefile"
)

var (
	conf string
	cpu  string
	port string
)

func init() {
	caddy.AppName = CoreName
	caddy.AppVersion = CoreVersion
	caddy.DefaultConfigFile = CoreFile
	caddy.Quiet = true

	dnsserver.Directives = plugin.Directives
}

func StartCoreDNSDaemon() {
	prepareFlags()

	caddy.RegisterCaddyfileLoader("flag", caddy.LoaderFunc(confLoader))
	caddy.SetDefaultCaddyfileLoader("default", caddy.LoaderFunc(defaultLoader))

	caddy.RegisterPlugin("rdns", caddy.Plugin{
		ServerType: "dns",
		Action:     rdns.Setup,
	})

	caddy.TrapSignals()

	if err := setCPU(cpu); err != nil {
		logrus.Fatal(err)
	}

	f, err := caddy.LoadCaddyfile(CoreType)
	if err != nil {
		logrus.Fatal(err)
	}

	instance, err := caddy.Start(f)
	if err != nil {
		logrus.Fatal(err)
	}

	instance.Wait()
}

func confLoader(serverType string) (caddy.Input, error) {
	if conf == "" {
		return nil, nil
	}

	if conf == "stdin" {
		return caddy.CaddyfileFromPipe(os.Stdin, serverType)
	}

	contents, err := ioutil.ReadFile(conf)
	if err != nil {
		return nil, err
	}
	return caddy.CaddyfileInput{
		Contents:       contents,
		Filepath:       conf,
		ServerTypeName: serverType,
	}, nil
}

func defaultLoader(serverType string) (caddy.Input, error) {
	contents, err := ioutil.ReadFile(caddy.DefaultConfigFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return caddy.CaddyfileInput{
		Contents:       contents,
		Filepath:       caddy.DefaultConfigFile,
		ServerTypeName: serverType,
	}, nil
}

func setCPU(cpu string) error {
	var numCPU int

	availCPU := runtime.NumCPU()

	if strings.HasSuffix(cpu, "%") {
		var percent float32
		pctStr := cpu[:len(cpu)-1]
		pctInt, err := strconv.Atoi(pctStr)
		if err != nil || pctInt < 1 || pctInt > 100 {
			return errors.New("invalid CPU value: percentage must be between 1-100")
		}
		percent = float32(pctInt) / 100
		numCPU = int(float32(availCPU) * percent)
	} else {
		num, err := strconv.Atoi(cpu)
		if err != nil || num < 1 {
			return errors.New("invalid CPU value: provide a number or percent greater than 0")
		}
		numCPU = num
	}

	if numCPU > availCPU {
		numCPU = availCPU
	}

	runtime.GOMAXPROCS(numCPU)
	return nil
}

func prepareFlags() {
	conf = os.Getenv("CORE_DNS_FILE")
	cpu = os.Getenv("CORE_DNS_CPU")
	port = os.Getenv("CORE_DNS_PORT")

	if err := flag.Set("dns.port", port); err != nil {
		logrus.Error(err)
	}
}
