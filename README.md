# Kubernetes-YARN

A version of [Kubernetes](https://github.com/GoogleCloudPlatform/kubernetes) using [Apache Hadoop YARN](http://hadoop.apache.org/docs/current/hadoop-yarn/hadoop-yarn-site/YARN.html) as the scheduler. Integrating Kubernetes with YARN lets users run [Docker](https://www.docker.com/whatisdocker/) containers packaged as pods (using Kubernetes) and YARN applications (using YARN), while ensuring common resource management across these (PaaS and data) workloads. 

## Kubernetes-YARN is currently in the protoype/alpha phase
This integration is under development. Please expect bugs and significant changes as we work towards making things more stable and adding additional features.


## Getting started
### Dev Environment
Kubernetes and Kubernetes-YARN are written in [Go](http://golang.org). Currently, only a [vagrant](http://www.vagrantup.com/)-based setup is supported. That said, bringing up a vagrant-based Kubernetes-YARN environment is fairly easy. 

Please ensure you have [boot2docker](http://boot2docker.io/), Go (at least 1.3), Vagrant (at least 1.6), VirtualBox (at least 4.3.x) and git installed. 

```
$ boot2docker init && boot2docker start
$ echo $GOPATH
/home/user/goproj
$ mkdir -p $GOPATH/src/github.com/hortonworks/
$ cd $GOPATH/src/github.com/hortonworks/
$ git clone git@github.com:hortonworks/kubernetes-yarn.git
$ cd kubernetes-yarn
$ export DOCKER_HOST=tcp://$(/usr/local/bin/boot2docker ip 2>/dev/null):2375 #provides docker daemon location for release builds
$ build/release.sh #builds kubernetes release binaries 
$ hack/build-go.sh #builds kubernetes client binaries
$ cluster/kube-up.sh #brings up kubernetes cluster
```
Following these steps will bring up a multi-VM cluster (1 master and 3 minions, by default) running Kubernetes and YARN. Please note that, depending on your local hardware and available bandwidth, bringing the cluster up could take a while to complete.
### YARN Resource Manager UI
By default, the kubernetes master is assigned the IP 10.245.1.2. The YARN resource manager runs on the name host. Once the vagrant cluster is running, the YARN UI is accessible at http://10.245.1.2:8088/cluster/apps

### Interacting with the Kubernetes cluster
For instructions on creating pods, running containers and other interactions with the Kubernetes cluster, please see Kubernetes' vagrant instructions [here](https://github.com/GoogleCloudPlatform/kubernetes/blob/master/docs/getting-started-guides/vagrant.md#running-containers)
