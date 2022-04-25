package controller

import (
	"fmt"
	"time"

	v1 "github.com/StatCan/kubeflow-controller/pkg/apis/kubeflowcontroller/v1"
	kubeflow "github.com/StatCan/kubeflow-controller/pkg/generated/clientset/versioned"
	informers "github.com/StatCan/kubeflow-controller/pkg/generated/informers/externalversions/kubeflowcontroller/v1"

	//v1 "github.com/StatCan/kubeflow-apis/apis/kubeflow/v1"
	//kubeflow "github.com/StatCan/kubeflow-apis/clientset/versioned"
	//informers "github.com/StatCan/kubeflow-apis/informers/externalversions/kubeflow/v1"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	k8sinformers "k8s.io/client-go/informers/core/v1"
	k8slisters "k8s.io/client-go/listers/core/v1"

	rbacv1informers "k8s.io/client-go/informers/rbac/v1"
	rbacv1listers "k8s.io/client-go/listers/rbac/v1"

	rbacv1 "k8s.io/api/rbac/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

const controllerAgentName = "internal-user-controller"

// Controller responds to new resources and applies the necessary configuration
type Controller struct {
	kubeflowClientset kubeflow.Interface

	podInformer k8sinformers.PodInformer
	podLister   k8slisters.PodLister
	podSynched  cache.InformerSynced

	profileInformerLister informers.ProfileInformer
	profileSynched        cache.InformerSynced

	roleBindingInformer rbacv1informers.RoleBindingInformer
	roleBindingLister   rbacv1listers.RoleBindingLister
	roleBindingSynced   cache.InformerSynced

	workqueue workqueue.RateLimitingInterface
	recorder  record.EventRecorder
}

// NewController creates a new Controller object.
func NewController(
	kubeclientset kubernetes.Interface,
	kubeflowclientset kubeflow.Interface,
	profileInformer informers.ProfileInformer,
	podInformer k8sinformers.PodInformer,
	roleBindingInformer rbacv1informers.RoleBindingInformer) *Controller {

	// Create event broadcaster
	log.Info("creating event broadcaster")

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeflowClientset:     kubeflowclientset,
		podInformer:           podInformer,
		podLister:             podInformer.Lister(),
		podSynched:            podInformer.Informer().HasSynced,
		profileInformerLister: profileInformer,
		profileSynched:        profileInformer.Informer().HasSynced,
		roleBindingInformer:   roleBindingInformer,
		roleBindingLister:     roleBindingInformer.Lister(),
		roleBindingSynced:     roleBindingInformer.Informer().HasSynced,
		workqueue:             workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "PodPolicy"),
		recorder:              recorder,
	}

	// Set up an event handler for when Profile resources change
	profileInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueProfile,
		UpdateFunc: func(old, new interface{}) {
			np := new.(*v1.Profile)
			op := old.(*v1.Profile)
			if np.ResourceVersion == op.ResourceVersion {
				return
			}
			controller.enqueueProfile(new)
		},
		DeleteFunc: controller.enqueueProfile,
	})

	// Set up an event handler for when Pod resources change
	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handlePodObject,
		UpdateFunc: func(old, new interface{}) {
			npod := new.(*corev1.Pod)
			opod := old.(*corev1.Pod)
			if npod.ResourceVersion == opod.ResourceVersion {
				return
			}
			controller.handlePodObject(npod)
		},
		DeleteFunc: controller.handlePodObject,
	})

	// Set up an event handler for when RoleBinding resources change
	roleBindingInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleRoleBindingObject,
		UpdateFunc: func(old, new interface{}) {
			newrb := new.(*rbacv1.RoleBinding)
			oldrb := old.(*rbacv1.RoleBinding)
			if newrb.ResourceVersion == oldrb.ResourceVersion {
				return
			}
			controller.handleRoleBindingObject(newrb)
		},
		DeleteFunc: controller.handleRoleBindingObject,
	})
	return controller
}

func (c *Controller) handlePodObject(npod interface{}) {
	pod := npod.(*corev1.Pod)
	namespace := pod.GetNamespace()
	existingProfile, err := c.profileInformerLister.Lister().Get(namespace)
	if err != nil {
		log.Errorf("failed to get profile: %v", err)
		return
	}
	c.enqueueProfile(existingProfile)
}

func (c *Controller) handleRoleBindingObject(newrb interface{}) {
	roleBinding := newrb.(*rbacv1.RoleBinding)
	namespace := roleBinding.GetNamespace()
	existingProfile, err := c.profileInformerLister.Lister().Get(namespace)
	if err != nil {
		log.Errorf("failed to get profile - rb: %v", err)
		return
	}
	c.enqueueProfile(existingProfile)
}

//Run runs the controller
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	if ok := cache.WaitForCacheSync(stopCh, c.podSynched, c.profileSynched); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	log.Info("starting workers")
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	log.Info("started workers")
	<-stopCh
	log.Info("shutting down workers")

	return nil
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var key string
		var ok bool

		if key, ok = obj.(string); !ok {
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}

		if err := c.syncHandler(key); err != nil {
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error synching %q: %v, requeing", key, err)
		}

		c.workqueue.Forget(obj)
		log.Infof("successfully synched %q", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) syncHandler(key string) error {

	profile, err := c.profileInformerLister.Lister().Get(key)
	if err != nil {
		log.Errorf("failed to get profile: %v", err)
		return err
	}

	// Handle the profile
	err = c.handleProfile(profile)
	if err != nil {
		log.Errorf("failed to handle profile: %v", err)
		return err
	}

	return nil
}

func (c *Controller) enqueueProfile(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}
