package resources

import (
	"context"
	"encoding/json"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
	"strings"
)

var (
	IgnoredConfigMapPrefixes = []string{
		"kube-root-ca",
		"istio-ca-root",
	}
)

func init() {
	Register("configmap", &FactoryWithAPI[corev1.ConfigMap, corev1.ConfigMapList]{
		FuncCreateAPI: func(client *kubernetes.Clientset, namespace string) API[corev1.ConfigMap, corev1.ConfigMapList] {
			return client.CoreV1().ConfigMaps(namespace)
		},
		FuncAfterList: func(list *corev1.ConfigMapList) (names []string) {
			for _, item := range list.Items {
				if HasPrefix(item.Name, IgnoredConfigMapPrefixes) {
					continue
				}
				names = append(names, item.Name)
			}
			return
		},
	})

	knownResources = append(knownResources, &Resource{
		Kind: "configmap",
		List: func(ctx context.Context, client *kubernetes.Clientset, namespace string) (names []string, err error) {
			var items *corev1.ConfigMapList
			if items, err = client.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{}); err != nil {
				return
			}
		outerLoop:
			for _, item := range items.Items {
				for _, ignored := range IgnoredConfigMapPrefixes {
					if strings.HasPrefix(item.Name, ignored) {
						continue outerLoop
					}
				}
				names = append(names, item.Name)
			}
			return
		},
		GetJSON: func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (data []byte, err error) {
			var obj *corev1.ConfigMap
			if obj, err = client.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{}); err != nil {
				return
			}
			data, err = json.Marshal(obj)
			return
		},
		SetJSON: func(ctx context.Context, client *kubernetes.Clientset, namespace, name string, data []byte) (err error) {
			var obj corev1.ConfigMap
			if err = json.Unmarshal(data, &obj); err != nil {
				return
			}
			obj.Namespace = namespace
			obj.Name = name

			var current *corev1.ConfigMap
			if current, err = client.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{}); err != nil {
				if errors.IsNotFound(err) {
					err = nil
				} else {
					return
				}
			} else {
				if main.GateNoUpdate.IsOn() {
					log.Println("SKIP")
					return
				}
				obj.ResourceVersion = current.ResourceVersion
			}

			if _, err = client.CoreV1().ConfigMaps(namespace).Update(ctx, &obj, metav1.UpdateOptions{}); err != nil {
				if errors.IsNotFound(err) {
					obj.ResourceVersion = ""
					if _, err = client.CoreV1().ConfigMaps(namespace).Create(ctx, &obj, metav1.CreateOptions{}); err != nil {
						return
					}
				}
				return
			}
			return
		},
	})
	knownResourceNames = append(knownResourceNames, "configmap")
}
