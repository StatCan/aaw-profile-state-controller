package controller

import (
	"fmt"
	"time"

	v1 "github.com/StatCan/kubeflow-controller/pkg/apis/kubeflowcontroller/v1"
	clientset "github.com/StatCan/kubeflow-controller/pkg/generated/clientset/versioned"
	kubeflow "github.com/StatCan/kubeflow-controller/pkg/generated/clientset/versioned"
	informers "github.com/StatCan/kubeflow-controller/pkg/generated/informers/externalversions/kubeflowcontroller/v1"
	k8sinformers "k8s.io/client-go/informers/core/v1"
	k8slisters "k8s.io/client-go/listers/core/v1"
	"github.com/prometheus/common/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
type Controller struct{
	kubeflowClientset				kubeflow.Interface

	podInformer						k8sinformers.PodInformer
	podLister						k8slisters.PodLister
	podSynched						cache.InformerSynced

	profileInformerLister			informers.ProfileInformer
	profileSynched					cache.InformerSynced

	workqueue						workqueue.RateLimitingInterface
	recorder						record.EventRecorder
}

// NewController creates a new Controller object.
func NewController(
	kubeclientset kubernetes.Interface,
	kubeflowclientset clientset.Interface,
	profileInformer informers.ProfileInformer,
	podInformer k8sinformers.PodInformer) *Controller {

	// Create event broadcaster
	log.Info("creating event broadcaster")

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeflowClientset:				kubeflowclientset,
		podInformer:					podInformer,
		podLister:						podInformer.Lister(),
		podSynched:						podInformer.Informer().HasSynced,
		profileInformerLister:			profileInformer,
		profileSynched: 				profileInformer.Informer().HasSynced,
		workqueue:						workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "PodPolicy"),
		recorder:						recorder,
	}

	// Set up an event handler for when Profile resources change
	profileInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueProfile,
		UpdateFunc: func(old, new interface{}){
			np := new.(*v1.Profile)
			op := old.(*v1.Profile)
			if np.ResourceVersion == op.ResourceVersion{
				return
			}
			controller.enqueueProfile(new)
		},
		DeleteFunc: controller.enqueueProfile,
	})

	// Set up an event handler for when Pod resources change
	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueuePod,
		UpdateFunc: func(old, new interface{}) {
			npod := new.(*corev1.Pod)
			opod := old.(*corev1.Pod)
			if npod.ResourceVersion == opod.ResourceVersion {
				return
			}
			controller.enqueuePod(new)
		},
		DeleteFunc: controller.enqueuePod,
	})

	return controller
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
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the pods object
	pods, err := c.podLister.Pods(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("pod %q in work queue no longer exists", key))
			return nil
		}
		return err
	}

	// Handle the profile
	err = c.handleProfile(pods)
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

func (c *Controller) enqueuePod(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

/*func (c *Controller) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		log.Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}
	log.Infof("Processing object: %s", object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		// If this object is not owned by a Profile, we should not do anything more
		// with it.
		if ownerRef.Kind != "Profile" {
			return
		}
		log.Infof("OWNER REF %v %v", ownerRef.Name, ownerRef.Kind)
		profile, err := c.profileInformerLister.Lister().Get(ownerRef.Name)
		if err != nil {
			log.Infof("ignoring orphaned object '%s' of profile '%s'", object.GetSelfLink(), ownerRef.Name)
			return
		}

		c.enqueueProfile(profile)
		return
	}
}*/
