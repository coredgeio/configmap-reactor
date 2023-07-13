package reactor

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/coredgeio/configmap-reactor/pkg/k8s"
)

type ConfigMapReactor struct {
	namespace     string
	deviceCmLabel string
	clientSet     *kubernetes.Clientset
}

// getConfigMapsWatch - get the watch interface for config maps
func (r *ConfigMapReactor) getConfigMapsWatch() (watch.Interface, error) {
	watchInterface, err := r.clientSet.CoreV1().ConfigMaps(r.namespace).Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return watchInterface, nil
}

// getConfigMaps - get all config maps running
func (r *ConfigMapReactor) getConfigMaps() ([]corev1.ConfigMap, error) {
	configMapsClient := r.clientSet.CoreV1().ConfigMaps(r.namespace)

	configMaps, err := configMapsClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return configMaps.Items, err
	} else if len(configMaps.Items) == 0 {
		return configMaps.Items, errors.New("no config maps available")
	}

	return configMaps.Items, nil
}

// getDeployments - get all deployments running
func (r *ConfigMapReactor) getDeployments() ([]appsv1.Deployment, error) {
	deploymentsClient := r.clientSet.AppsV1().Deployments(r.namespace)

	deployments, err := deploymentsClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return deployments.Items, err
	} else if len(deployments.Items) == 0 {
		return deployments.Items, errors.New("no deployments available")
	}

	return deployments.Items, nil
}

// createDeviceConfigMap - creates a device config map
func (r *ConfigMapReactor) createDeviceConfigMap() error {
	_, err := r.clientSet.CoreV1().ConfigMaps(r.namespace).Create(context.Background(), &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DeviceConfigMap,
			Namespace: r.namespace,
		},
		Data: make(map[string]string),
	}, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

// handleConfigMapEvent - logic that reacts to config map update operation
func (r *ConfigMapReactor) handleConfigMapEvent(cmEvent watch.Event) {
	if cmEvent.Type == watch.Modified {
		configMap, ok := cmEvent.Object.(*corev1.ConfigMap)
		if !ok {
			log.Fatalln("failed to convert watch event object to config map")
		}
		if configMap.ObjectMeta.Name == DeviceConfigMap {
			deploymentList, err := r.getDeployments()
			if err != nil {
				log.Fatalln("unable to list deployments:", err)
			}
			for _, deployment := range deploymentList {
				for k := range deployment.Spec.Template.ObjectMeta.Labels {
					if k == r.deviceCmLabel {
						if deployment.Spec.Template.ObjectMeta.Annotations == nil {
							deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
						}
						deployment.Spec.Template.ObjectMeta.Annotations[RestartAnnotation] = time.Now().Format("20060102150405")
						jsonBytes, err := json.Marshal(deployment)
						if err != nil {
							log.Fatalln("unable to marshal deployment into JSON bytes", err)
						}
						_, err = r.clientSet.AppsV1().Deployments(r.namespace).Patch(context.Background(), deployment.ObjectMeta.Name, types.StrategicMergePatchType, jsonBytes, metav1.PatchOptions{})
						if err != nil {
							log.Fatalln("unable to patch deployment for rollout restart", err)
						}
						break
					}
				}
			}
		}
	}
}

func (r *ConfigMapReactor) init() {
	configMaps, err := r.getConfigMaps()
	if err != nil {
		log.Fatalln("failed to list config maps:", err)
	}
	cmExists := false
	for _, cm := range configMaps {
		if cm.ObjectMeta.Name == DeviceConfigMap {
			cmExists = true
			break
		}
	}
	if !cmExists {
		err := r.createDeviceConfigMap()
		if err != nil {
			log.Fatalln("failed to create device config map:", err)
		}
	}
	go func() {
		watcher, err := r.getConfigMapsWatch()
		if err != nil {
			log.Fatalf("failed to get config map watch interface err: %v\n", err)
			return
		}
		for {
			select {
			case e, ok := <-watcher.ResultChan():
				if !ok {
					watcher.Stop()
					watcher, err = r.getConfigMapsWatch()
					if err != nil {
						log.Fatalf("failed to get config map watch interface err: %v\n", err)
					}
					continue
				}
				r.handleConfigMapEvent(e)
			}
		}
	}()
}

func CreateConfigMapReactor(deviceCmLabel string) *ConfigMapReactor {
	ns, err := k8s.GetServiceAccountNamespace()
	if err != nil {
		log.Fatal("failed to find namespace from standard k8s path")
	}
	clientset, err := kubernetes.NewForConfig(ctrl.GetConfigOrDie())
	if err != nil {
		log.Fatalln("failed to load k8s config for controller ", err)
	}
	cmReactor := &ConfigMapReactor{
		namespace:     ns,
		deviceCmLabel: deviceCmLabel,
		clientSet:     clientset,
	}
	cmReactor.init()
	return cmReactor
}
