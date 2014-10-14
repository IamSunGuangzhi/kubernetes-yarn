# Kubernetes-YARN

A fork of [Kubernetes](https://github.com/GoogleCloudPlatform/kubernetes) using [Apache Hadoop YARN](http://hadoop.apache.org/docs/current/hadoop-yarn/hadoop-yarn-site/YARN.html) as the default scheduler.

YARN is multi-workload resource manager for batch, interactive and real-time workloads on Hadoop. Kubernetes is an open source implementation of container cluster management. Integrating kubernetes with YARN lets us use common resource management across data and PaaS workloads.  

## Kubernetes-YARN is currently in the protoype/alpha phase
This integration is under development. Please expect bugs and significant changes as we work towards making things more stable and adding additional features.


## Getting started
### Dev Environment
Kubernetes and Kubernetes-YARN are written in [Go](http://golang.org). Currently, only a [vagrant](http://www.vagrantup.com/)-based setup. That said, bringing up a vagrant-based Kubernetes-YARN environment is fairly easy. 

Please ensure you have Go (at least 1.3), Vagrant (at least 1.6), VirtualBox (at least 4.3.x) and git installed. 


```
$ echo $GOPATH
/home/user/goproj
$ mkdir -p $GOPATH/src/github.com/hortonworks/
$ cd $GOPATH/src/github.com/hortonworks/
$ git clone git@github.com:hortonworks/kubernetes-yarn.git
$ cd kubernetes-yarn
$ hack/build-go.sh
$ vagrant up
```

Following these steps will bring up a multi-VM cluster (1 master and 3 minions, by default) running kubernetes and YARN. Please note that, depending on your local hardware and available bandwidth, `vagrant up` could take a while to complete.

### Interacting with the kubernetes cluster
For instructions on creating pods, running containers and other interactions with the kubernetes cluster, please see Kubernetes' vagrant instructions [here](https://github.com/GoogleCloudPlatform/kubernetes/blob/master/docs/getting-started-guides/vagrant.md#running-containers)
