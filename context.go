package dockergen

import (
	"bufio"
	"os"
	"regexp"
	"sync"

	"github.com/fsouza/go-dockerclient"
)

var (
	mu         sync.RWMutex
	dockerInfo Docker
	dockerEnv  *docker.Env
)

type Context []*RuntimeContainer

func (c *Context) Env() map[string]string {
	return splitKeyValueSlice(os.Environ())
}

func (c *Context) Docker() Docker {
	mu.RLock()
	defer mu.RUnlock()
	return dockerInfo
}

func SetServerInfo(d *docker.DockerInfo) {
	mu.Lock()
	defer mu.Unlock()
	dockerInfo = Docker{
		Name:               d.Name,
		NumContainers:      d.Containers,
		NumImages:          d.Images,
		Version:            dockerEnv.Get("Version"),
		ApiVersion:         dockerEnv.Get("ApiVersion"),
		GoVersion:          dockerEnv.Get("GoVersion"),
		OperatingSystem:    dockerEnv.Get("Os"),
		Architecture:       dockerEnv.Get("Arch"),
		CurrentContainerID: GetCurrentContainerID(),
	}
}

func SetDockerEnv(d *docker.Env) {
	mu.Lock()
	defer mu.Unlock()
	dockerEnv = d
}

type Address struct {
	IP           string
	IP6LinkLocal string
	IP6Global    string
	Port         string
	HostPort     string
	Proto        string
	HostIP       string
}

type Network struct {
	IP                  string
	Name                string
	Gateway             string
	EndpointID          string
	IPv6Gateway         string
	GlobalIPv6Address   string
	MacAddress          string
	GlobalIPv6PrefixLen int
	IPPrefixLen         int
}

type Volume struct {
	Path      string
	HostPath  string
	ReadWrite bool
}

type State struct {
	Running bool
}

type RuntimeContainer struct {
	ID           string
	Addresses    []Address
	Networks     []Network
	Gateway      string
	Name         string
	Hostname     string
	Image        DockerImage
	Env          map[string]string
	Volumes      map[string]Volume
	Node         SwarmNode
	Labels       map[string]string
	IP           string
	IP6LinkLocal string
	IP6Global    string
	Mounts       []Mount
	State        State
}

func (r *RuntimeContainer) Equals(o RuntimeContainer) bool {
	return r.ID == o.ID && r.Image == o.Image
}

func (r *RuntimeContainer) PublishedAddresses() []Address {
	mapped := []Address{}
	for _, address := range r.Addresses {
		if address.HostPort != "" {
			mapped = append(mapped, address)
		}
	}
	return mapped
}

type DockerImage struct {
	Registry   string
	Repository string
	Tag        string
}

func (i *DockerImage) String() string {
	ret := i.Repository
	if i.Registry != "" {
		ret = i.Registry + "/" + i.Repository
	}
	if i.Tag != "" {
		ret = ret + ":" + i.Tag
	}
	return ret
}

type SwarmNode struct {
	ID      string
	Name    string
	Address Address
}

type Mount struct {
	Name        string
	Source      string
	Destination string
	Driver      string
	Mode        string
	RW          bool
}

type Docker struct {
	Name               string
	NumContainers      int
	NumImages          int
	Version            string
	ApiVersion         string
	GoVersion          string
	OperatingSystem    string
	Architecture       string
	CurrentContainerID string
}

func GetCurrentContainerID() string {
	file, err := os.Open("/proc/self/cgroup")

	if err != nil {
		return ""
	}

	reader := bufio.NewReader(file)
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		_, lines, err := bufio.ScanLines([]byte(scanner.Text()), true)
		if err == nil {
			strLines := string(lines)
			if id := matchDockerCurrentContainerID(strLines); id != "" {
				return id
			} else if id := matchECSCurrentContainerID(strLines); id != "" {
				return id
			}else if id := matchDockerLimitCurrentContainerID(strLines); id != "" {
				return id
			}
		}
	}

	return ""
}

func matchDockerCurrentContainerID(lines string) string {
	regex := "/docker[/-]([[:alnum:]]{64})(\\.scope)?$"
	re := regexp.MustCompilePOSIX(regex)

	if re.MatchString(lines) {
		submatches := re.FindStringSubmatch(string(lines))
		containerID := submatches[1]

		return containerID
	}
	return ""
}

func matchECSCurrentContainerID(lines string) string {
	regex := "/ecs\\/[^\\/]+\\/(.+)$"
	re := regexp.MustCompilePOSIX(regex)

	if re.MatchString(string(lines)) {
		submatches := re.FindStringSubmatch(string(lines))
		containerID := submatches[1]

		return containerID
	}

	return ""
}

func matchDockerLimitCurrentContainerID(lines string) string {
	regex := "/docker[_-]limit.slice[/-]([[:alnum:]]{64})(\\.scope)?$"
	re := regexp.MustCompilePOSIX(regex)

	if re.MatchString(string(lines)) {
		submatches := re.FindStringSubmatch(string(lines))
		containerID := submatches[1]

		return containerID
	}

	return ""
}