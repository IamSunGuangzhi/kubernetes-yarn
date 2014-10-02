package scheduler

import (
	"errors"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/hortonworks/gohadoop/hadoop_common/security"
	"github.com/hortonworks/gohadoop/hadoop_yarn"
	"github.com/hortonworks/gohadoop/hadoop_yarn/conf"
	"github.com/hortonworks/gohadoop/hadoop_yarn/yarn_client"

	"log"
	"net"
	"os"
	"time"
)

func YARNInit() (*yarn_client.YarnClient, *yarn_client.AMRMClient) {
	var err error

	//Hack! This should be external, but doing this here for demo purposes
	hadoopHome := "/home/vagrant/hadoop/install/hadoop-2.6.0-SNAPSHOT"
	os.Setenv("HADOOP_HOME", hadoopHome)
	os.Setenv("HADOOP_COMMON_HOME", hadoopHome)
	os.Setenv("HADOOP_CONF_DIR", hadoopHome+"/etc/hadoop")
	os.Setenv("HADOOP_HDFS_HOME", hadoopHome)
	os.Setenv("HADOOP_MAPRED_HOME", hadoopHome)

	// Create YarnConfiguration
	conf, _ := conf.NewYarnConfiguration()

	// Create YarnClient
	yarnClient, _ := yarn_client.CreateYarnClient(conf)

	// Create new application to get ApplicationSubmissionContext
	_, asc, _ := yarnClient.CreateNewApplication()

	// Some useful information
	queue := "default"
	appName := "kubernetes"
	appType := "PAAS"
	unmanaged := true
	clc := hadoop_yarn.ContainerLaunchContextProto{}

	// Setup ApplicationSubmissionContext for the application
	asc.AmContainerSpec = &clc
	asc.ApplicationName = &appName
	asc.Queue = &queue
	asc.ApplicationType = &appType
	asc.UnmanagedAm = &unmanaged

	// Submit!
	err = yarnClient.SubmitApplication(asc)
	if err != nil {
		log.Fatal("yarnClient.SubmitApplication ", err)
	}
	log.Println("Successfully submitted unmanaged application: ", asc.ApplicationId)
	time.Sleep(1 * time.Second)

	appReport, err := yarnClient.GetApplicationReport(asc.ApplicationId)
	if err != nil {
		log.Fatal("yarnClient.GetApplicationReport ", err)
	}
	appState := appReport.GetYarnApplicationState()
	for appState != hadoop_yarn.YarnApplicationStateProto_ACCEPTED {
		log.Println("Application in state ", appState)
		time.Sleep(1 * time.Second)
		appReport, err = yarnClient.GetApplicationReport(asc.ApplicationId)
		appState = appReport.GetYarnApplicationState()
		if appState == hadoop_yarn.YarnApplicationStateProto_FAILED || appState == hadoop_yarn.YarnApplicationStateProto_KILLED {
			log.Fatal("Application in state ", appState)
		}
	}

	amRmToken := appReport.GetAmRmToken()

	if amRmToken != nil {
		savedAmRmToken := *amRmToken
		service, _ := conf.GetRMSchedulerAddress()
		savedAmRmToken.Service = &service
		security.GetCurrentUser().AddUserToken(&savedAmRmToken)
	}

	log.Println("Application in state ", appState)

	// Create AMRMClient
	var attemptId int32
	attemptId = 1
	applicationAttemptId := hadoop_yarn.ApplicationAttemptIdProto{ApplicationId: asc.ApplicationId, AttemptId: &attemptId}

	rmClient, _ := yarn_client.CreateAMRMClient(conf, &applicationAttemptId)
	log.Println("Created RM client: ", rmClient)

	// Wait for ApplicationAttempt to be in Launched state
	appAttemptReport, err := yarnClient.GetApplicationAttemptReport(&applicationAttemptId)
	appAttemptState := appAttemptReport.GetYarnApplicationAttemptState()
	for appAttemptState != hadoop_yarn.YarnApplicationAttemptStateProto_APP_ATTEMPT_LAUNCHED {
		log.Println("ApplicationAttempt in state ", appAttemptState)
		time.Sleep(1 * time.Second)
		appAttemptReport, err = yarnClient.GetApplicationAttemptReport(&applicationAttemptId)
		appAttemptState = appAttemptReport.GetYarnApplicationAttemptState()
	}
	log.Println("ApplicationAttempt in state ", appAttemptState)

	// Register with ResourceManager
	log.Println("About to register application master.")
	err = rmClient.RegisterApplicationMaster("", -1, "")
	if err != nil {
		log.Fatal("rmClient.RegisterApplicationMaster ", err)
	}
	log.Println("Successfully registered application master.")

	return yarnClient, rmClient
}

type YARNScheduler struct {
	yarnClient          *yarn_client.YarnClient
	rmClient            *yarn_client.AMRMClient
	podsToContainersMap map[string]*hadoop_yarn.ContainerIdProto
}

func NewYARNScheduler() Scheduler {
	yarnC, rmC := YARNInit()
	podsToContainers := make(map[string]*hadoop_yarn.ContainerIdProto)

	return &YARNScheduler{
		yarnClient:          yarnC,
		rmClient:            rmC,
		podsToContainersMap: podsToContainers}
}

func (yarnScheduler *YARNScheduler) Delete(id string) error {
	log.Printf("attempting to delete pods with id: %s", id)
	rmClient := yarnScheduler.rmClient
	containerId, found := yarnScheduler.podsToContainersMap[id]
	if !found {
		return errors.New("attempting to delete a pod that doesn't have an associated container")
	}

	rmClient.ReleaseAssignedContainer(containerId)
	allocateResponse, err := rmClient.Allocate()
	if err != nil {
		return errors.New("YARN returned an allocation error when releasing container")
	}

	log.Println("allocateResponse(release): ", allocateResponse)
	return nil
}

func (yarnScheduler *YARNScheduler) Schedule(pod api.Pod, minionLister MinionLister) (string, error) {
	rmClient := yarnScheduler.rmClient

	// Add resource requests
	const numContainers = int32(1)
	memory := int32(128)
	resource := hadoop_yarn.ResourceProto{Memory: &memory}
	numAllocatedContainers := int32(0)
	const maxAllocationAttempts = int(5)
	allocationAttempts := 0
	allocatedContainers := make([]*hadoop_yarn.ContainerProto, numContainers, numContainers)

	rmClient.AddRequest(1, "*", &resource, numContainers)

	for numAllocatedContainers < numContainers && allocationAttempts < maxAllocationAttempts {
		// Try to get containers now...
		allocateResponse, err := rmClient.Allocate()
		if err != nil {
			return "<invalid_host>", errors.New("YARN returned an allocation error")
		}

		log.Println("allocateResponse(resource): ", *allocateResponse)

		for _, container := range allocateResponse.AllocatedContainers {
			allocatedContainers[numAllocatedContainers] = container
			numAllocatedContainers++
			log.Println("#containers allocated so far: ", numAllocatedContainers)

			yarnScheduler.podsToContainersMap[pod.ID] = container.GetId()
			host := *container.NodeId.Host
			log.Println("allocated container on: ", host)

			//We have the hostname available. return from here.
			return findMinionForHost(host, minionLister)
		}

		allocationAttempts++
		log.Println("#containers allocated: ", len(allocateResponse.AllocatedContainers))
		log.Println("Total #containers allocated so far: ", numAllocatedContainers)

		// Sleep for a while before trying again
		log.Println("Sleeping...")
		time.Sleep(3 * time.Second)
		log.Println("Sleeping... done!")
	}

	log.Println("Final #containers allocated: ", numAllocatedContainers)

	return "<invalid_host>", errors.New("unable to schedule pod! YARN didn't allocate a container")
}

/* YARN returns hostnames, but minions maybe using IPs.
TODO: This is an expensive mechanism to find the right minion corresponding to the YARN node.
      Find a better mechanism if possible
*/
func findMinionForHost(host string, minionLister MinionLister) (string, error) {
	hostIPs, err := net.LookupIP(host)

	if err != nil {
		return "<invalid_host>", errors.New("unable to lookup IPs for YARN host: " + host)
	}

	for _, hostIP := range hostIPs {
		minions, err := minionLister.List()
		if err != nil {
			return "<invalid_host>", errors.New("update to list minions")
		}

		for _, minion := range minions {
			minionIPs, err := net.LookupIP(minion)

			if err != nil {
				return "<invalid_host>", errors.New("unable to lookup IPs for minion: " + minion)
			}

			for _, minionIP := range minionIPs {
				if hostIP.Equal(minionIP) {
					log.Printf("YARN node %s maps to minion: %s", host, minion)
					return minion, nil
				}
			}
		}
	}

	return "<invalid_host>", errors.New("unable to find minion for YARN host: " + host)
}
