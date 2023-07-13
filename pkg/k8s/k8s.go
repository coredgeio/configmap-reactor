package k8s

import (
	"io/ioutil"
)

func GetServiceAccountNamespace() (string, error) {
	// get namespace of the pod in which it is running
	bytes, err := ioutil.ReadFile(K8sServiceAccountNamespacePath)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
