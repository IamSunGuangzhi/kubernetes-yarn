#!/usr/bin/bash

HADOOP_INSTALL_DIR=/home/vagrant/hadoop/install
HADOOP_HOME=${HADOOP_INSTALL_DIR}/hadoop-2.6.0-SNAPSHOT

source ${HADOOP_HOME}/env.sh

${HADOOP_HOME}/sbin/yarn-daemon.sh stop resourcemanager
${HADOOP_HOME}/sbin/yarn-daemon.sh start resourcemanager

DELAY=30

echo -n "Giving YARN Resource Manager $DELAY seconds to initialize "

for ((I=0; $I<$DELAY; I++))
do
        echo -n "."
        sleep 1
done

echo " done."

