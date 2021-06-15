package dockergen

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"sync"

	docker "github.com/fsouza/go-dockerclient"
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
	Killing bool
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

// GetCurrentContainerID attempts to extract the current container ID from the provided file paths.
// If no files paths are provided, it will default to /proc/1/cpuset, /proc/self/cgroup and /proc/self/mountinfo.
// It attempts to match the HOSTNAME first then use the fallback method, and returns with the first valid match.
func GetCurrentContainerID(filepaths ...string) (id string) {
	if len(filepaths) == 0 {
		filepaths = []string{"/proc/1/cpuset", "/proc/self/cgroup", "/proc/self/mountinfo"}
	}

	var files []io.Reader

	for _, filepath := range filepaths {
		file, err := os.Open(filepath)
		if err != nil {
			continue
		}
		defer file.Close()
		files = append(files, file)
	}

	reader := io.MultiReader(files...)
	var buffer bytes.Buffer
	tee := io.TeeReader(reader, &buffer)

	// We try to match a 64 character hex string starting with the hostname first
	scanner := bufio.NewScanner(tee)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		_, lines, err := bufio.ScanLines([]byte(scanner.Text()), true)
		if err == nil {
			strLines := string(lines)
			if id = matchContainerIDWithHostname(strLines); len(id) == 64 {
				return
			}
		}
	}

	// If we didn't get any ID that matches the hostname, fall back to matching the first 64 character hex string
	scanner = bufio.NewScanner(&buffer)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		_, lines, err := bufio.ScanLines([]byte(scanner.Text()), true)
		if err == nil {
			strLines := string(lines)
			if id = matchContainerID(strLines); len(id) == 64 {
				return
			}
		}
	}

	return
}

func matchContainerIDWithHostname(lines string) string {
	hostname := os.Getenv("HOSTNAME")
	re := regexp.MustCompilePOSIX("^[[:alnum:]]{12}$")

	if re.MatchString(hostname) {
		regex := fmt.Sprintf("(%s[[:alnum:]]{52})", hostname)
		re := regexp.MustCompilePOSIX(regex)

		if re.MatchString(lines) {
			submatches := re.FindStringSubmatch(string(lines))
			containerID := submatches[1]

			return containerID
		}
	}
	return ""
}

func matchContainerID(lines string) string {
	regex := "([[:alnum:]]{64})"
	re := regexp.MustCompilePOSIX(regex)

	if re.MatchString(lines) {
		submatches := re.FindStringSubmatch(string(lines))
		containerID := submatches[1]

		return containerID
	}
	return ""
}
