// Copyright 2014 The Serviced Authors.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package isvcs

import (
	"github.com/Sirupsen/logrus"
	"github.com/control-center/serviced/dfs/docker"
	"github.com/control-center/serviced/domain"
	"github.com/control-center/serviced/logging"
	"github.com/control-center/serviced/utils"

	"fmt"
	"github.com/control-center/serviced/config"
	"os"
	"strings"
	"time"
)

var (
	Mgr *Manager
	log = logging.PackageLogger()
)

const (
	IMAGE_REPO    = "zenoss/serviced-isvcs"
	IMAGE_TAG     = "v60"
	ZK_IMAGE_REPO = "zenoss/isvcs-zookeeper"
	ZK_IMAGE_TAG  = "v10"
)

type IServiceHealthResult struct {
	ServiceName    string
	ContainerName  string
	ContainerID    string
	HealthStatuses []domain.HealthCheckStatus
}

//
func PreInit() error {
	// Setup the initial Isvcs
	InitAllIsvcs()

	// Set the environment map.
	if err := setIsvcsEnv(); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to set isvcs options: %s\n", err)
		return err
	}

	return nil
}

func Init(esStartupTimeoutInSeconds int, dockerLogDriver string, dockerLogConfig map[string]string, dockerAPI docker.Docker) {
	if err := PreInit(); err != nil {
		log.WithFields(logrus.Fields{
			"isvc": "PreInit",
		}).WithError(err).Fatal("Unable to initialize ISVCS")
	}

	elasticsearch_serviced.StartupTimeout = time.Duration(esStartupTimeoutInSeconds) * time.Second
	elasticsearch_logstash.StartupTimeout = time.Duration(esStartupTimeoutInSeconds) * time.Second

	Mgr = NewManager(utils.LocalDir("images"), utils.TempDir("var/isvcs"), dockerLogDriver, dockerLogConfig)

	elasticsearch_serviced.docker = dockerAPI
	if err := Mgr.Register(elasticsearch_serviced); err != nil {
		log.WithFields(logrus.Fields{
			"isvc": "elasticsearch-serviced",
		}).WithError(err).Fatal("Unable to register internal service")
	}
	elasticsearch_logstash.docker = dockerAPI
	if err := Mgr.Register(elasticsearch_logstash); err != nil {
		log.WithFields(logrus.Fields{
			"isvc": "elasticsearch-logstash",
		}).WithError(err).Fatal("Unable to register internal service")
	}
	zookeeper.docker = dockerAPI
	if err := Mgr.Register(zookeeper); err != nil {
		log.WithFields(logrus.Fields{
			"isvc": "zookeeper",
		}).WithError(err).Fatal("Unable to register internal service")
	}
	logstash.docker = dockerAPI
	if err := Mgr.Register(logstash); err != nil {
		log.WithFields(logrus.Fields{
			"isvc": "logstash",
		}).WithError(err).Fatal("Unable to register internal service")
	}
	opentsdb.docker = dockerAPI
	if err := Mgr.Register(opentsdb); err != nil {
		log.WithFields(logrus.Fields{
			"isvc": "opentsdb",
		}).WithError(err).Fatal("Unable to register internal service")
	}
	dockerRegistry.docker = dockerAPI
	if err := Mgr.Register(dockerRegistry); err != nil {
		log.WithFields(logrus.Fields{
			"isvc": "docker-registry",
		}).WithError(err).Fatal("Unable to register internal service")
	}
	kibana.docker = dockerAPI
	if err := Mgr.Register(kibana); err != nil {
		log.WithFields(logrus.Fields{
			"isvc": "kibana",
		}).WithError(err).Fatal("Unable to register internal service")
	}
}

func InitServices(isvcNames []string, dockerLogDriver string, dockerLogConfig map[string]string, dockerAPI docker.Docker) {
	if err := PreInit(); err != nil {
		log.WithFields(logrus.Fields{
			"isvc": "PreInit",
		}).WithError(err).Fatal("Unable to initialize ISVCS")
	}

	Mgr = NewManager(utils.LocalDir("images"), utils.TempDir("var/isvcs"), dockerLogDriver, dockerLogConfig)
	for _, isvcName := range isvcNames {
		switch isvcName {
		case "zookeeper":
			zookeeper.docker = dockerAPI
			if err := Mgr.Register(zookeeper); err != nil {
				log.WithFields(logrus.Fields{
					"isvc": "zookeeper",
				}).WithError(err).Fatal("Unable to register internal service")
			}
		}
	}
}

// This function sets up key pieces of information for ISVCS in the environment map.
// (Adapted from the cmd.go cli function of the same name)
func setIsvcsEnv() error {
	options := config.GetOptions()

	if zkid := options.IsvcsZKID; zkid > 0 {
		if err := AddEnv(fmt.Sprintf("zookeeper:ZKID=%d", zkid)); err != nil {
			return err
		}
	}
	if zkquorum := strings.Join(options.IsvcsZKQuorum, ","); zkquorum != "" {
		if err := AddEnv("zookeeper:ZK_QUORUM=" + zkquorum); err != nil {
			return err
		}
	}
	for _, val := range options.IsvcsENV {
		if err := AddEnv(val); err != nil {
			return err
		}
	}
	return nil
}
