package maas

import (
	"launchpad.net/juju-core/environs"
	"launchpad.net/juju-core/environs/config"
	"launchpad.net/juju-core/log"
	"launchpad.net/juju-core/state"
	"launchpad.net/juju-core/state/api"
)

type maasEnviron struct {
	name string
}

var _ environs.Environ = (*maasEnviron)(nil)

func (env *maasEnviron) Name() string {
	return env.name
}

func (env *maasEnviron) Bootstrap(uploadTools bool, stateServerCert, stateServerKey []byte) error {
	log.Printf("environs/maas: bootstrapping environment %q.", env.Name())
	panic("Not implemented.")
}

func (*maasEnviron) StateInfo() (*state.Info, *api.Info, error) {
	panic("Not implemented.")
}

func (*maasEnviron) Config() *config.Config {
	panic("Not implemented.")
}

func (env *maasEnviron) SetConfig(cfg *config.Config) error {
	env.name = cfg.Name()
	panic("Not implemented.")
}

func (*maasEnviron) StartInstance(machineId string, info *state.Info, apiInfo *api.Info, tools *state.Tools) (environs.Instance, error) {
	panic("Not implemented.")
}

func (*maasEnviron) StopInstances([]environs.Instance) error {
	panic("Not implemented.")
}

func (*maasEnviron) Instances([]state.InstanceId) ([]environs.Instance, error) {
	panic("Not implemented.")
}

func (*maasEnviron) AllInstances() ([]environs.Instance, error) {
	panic("Not implemented.")
}

func (*maasEnviron) Storage() environs.Storage {
	panic("Not implemented.")
}

func (*maasEnviron) PublicStorage() environs.StorageReader {
	panic("Not implemented.")
}

func (env *maasEnviron) Destroy([]environs.Instance) error {
	log.Printf("environs/maas: destroying environment %q", env.name)
	panic("Not implemented.")
}

func (*maasEnviron) AssignmentPolicy() state.AssignmentPolicy {
	panic("Not implemented.")
}

func (*maasEnviron) OpenPorts([]state.Port) error {
	panic("Not implemented.")
}

func (*maasEnviron) ClosePorts([]state.Port) error {
	panic("Not implemented.")
}

func (*maasEnviron) Ports() ([]state.Port, error) {
	panic("Not implemented.")
}

func (*maasEnviron) Provider() environs.EnvironProvider {
	panic("Not implemented.")
}
