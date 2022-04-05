package controller

import (
	"fmt"
	"time"
	"strconv"
	"strings"

	v1 "github.com/StatCan/kubeflow-controller/pkg/apis/kubeflowcontroller/v1"
	clientset "github.com/StatCan/kubeflow-controller/pkg/generated/clientset/versioned"
	kubeflow "github.com/StatCan/kubeflow-controller/pkg/generated/clientset/versioned"
	informers "github.com/StatCan/kubeflow-controller/pkg/generated/informers/externalversions/kubeflowcontroller/v1"
	k8sinformers "k8s.io/client-go/informers/core/v1"
	k8slisters "k8s.io/client-go/listers/core/v1"
	"github.com/prometheus/common/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	profileInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleProfileObject,
		UpdateFunc: func(old, new interface{}){
			np := new.(*v1.Profile)
			op := old.(*v1.Profile)
			if np.ResourceVersion == op.ResourceVersion{
				return
			}
			controller.handleProfileObject(new)
		},
		DeleteFunc: controller.handleProfileObject,
	})

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handlePodObject,
		UpdateFunc: func(old, new interface{}){
			npod := new.(*corev1.Pod)
			opod := old.(*corev1.Pod)
			if npod.ResourceVersion == opod.ResourceVersion{
				return
			}
			controller.handlePodObject(new)
		},
		DeleteFunc: controller.handlePodObject,
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

	// Get the Pod object
	Pod, err := c.podLister.Pods(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("Pod %q in work queue no longer exists", key))
			return nil
		}

		return err
	}

	// Handle the Profile
	err = c.handleProfile(Pod)
	if err != nil {
		log.Errorf("failed to handle profile: %v", err)
		return err
	}

	return nil
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

func (c *Controller) handleProfileObject(obj interface{}) {
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

	namespace := object.ObjectMeta.Namespace
	hasEmpOnlyFeatures := false
	for _, pod := c.podLister.Pods(namespace) {
		if sasImage(pod) {
			hasEmpOnlyFeatures = true
			break
		}
	}
	object.ObjectMeta.Labels["state.aaw.statcan.gc.ca"] = strconv.FormatBool(hasEmpOnlyFeatures)
}

func sasImage(pod *corev1.Pod) bool {
	image := pod.Spec.Containers[0].Image
	sasImage := strings.HasPrefix(image, "k8scc01covidacr.azurecr.io/sas:")
	return sasImage
}

func (c *Controller) handlePodObject(obj interface{}) {
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

	ns := object.ObjectMeta.Namespace
	profile, err := c.profileInformerLister.Lister().Get(ns)
	c.handleProfileObject(profile)
	
	
}

