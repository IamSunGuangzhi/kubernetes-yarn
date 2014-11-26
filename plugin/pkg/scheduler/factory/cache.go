package factory

import (
	"github.com/GoogleCloudPlatform/kubernetes/pkg/client/cache"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/scheduler"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/util"

	"github.com/golang/glog"
)

//schedulerCache "wraps" a cache store and notifies the scheduler of
//certain caching events. For now, only the "delete" notification needs to be relayed
type schedulerCache struct {
	store     cache.Store
	scheduler scheduler.Scheduler
}

func NewSchedulerCache(store cache.Store, scheduler scheduler.Scheduler) cache.Store {
	return &schedulerCache{store: store, scheduler: scheduler}
}

func (c *schedulerCache) Add(id string, obj interface{}) {
	c.store.Add(id, obj)
}

func (c *schedulerCache) Update(id string, obj interface{}) {
	c.store.Update(id, obj)
}

func (c *schedulerCache) Delete(id string) {
	c.store.Delete(id)
	glog.V(0).Infof("the following pod has been deleted from cache: %s. notifying scheduler", id)

	if statefulScheduler, ok := c.scheduler.(scheduler.StatefulScheduler); ok {
		statefulScheduler.Delete(id)
	} else {
		glog.V(0).Infof("scheduler does not support deletes. this will likely lead to container leaks.")
	}
}

func (c *schedulerCache) List() []interface{} {
	return c.store.List()
}

func (c *schedulerCache) ContainedIDs() util.StringSet {
	return c.store.ContainedIDs()
}

func (c *schedulerCache) Get(id string) (item interface{}, exists bool) {
	return c.store.Get(id)
}

func (c *schedulerCache) Replace(idToObj map[string]interface{}) {
	c.store.Replace(idToObj)
}
