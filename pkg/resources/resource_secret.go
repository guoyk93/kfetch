package resources

import (
	"context"
	"encoding/json"
	"github.com/guoyk93/kfetch"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
	"strconv"
	"strings"
)

func init() {
	knownResources = append(knownResources, &Resource{
		Kind: "secret",
		List: func(ctx context.Context, client *kubernetes.Clientset, namespace string) (names []string, err error) {
			var items *corev1.SecretList
			if items, err = client.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{}); err != nil {
				return
			}
			for _, item := range items.Items {
				// ignore service account token
				if item.Type == corev1.SecretTypeServiceAccountToken {
					continue
				}
				if strings.Contains(string(item.Type), "helm.sh") {
					continue
				}
				if strings.HasPrefix(item.Name, "ezopsdb") {
					continue
				}
				if strings.HasPrefix(item.Name, "gopsdb") {
					continue
				}
				// ignore replicated secret
				if replicated, _ := strconv.ParseBool(item.Annotations["autoops.auto-replicate-secret/replicated"]); replicated {
					continue
				}
				names = append(names, item.Name)
			}
			return
		},
		GetJSON: func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (data []byte, err error) {
			var obj *corev1.Secret
			if obj, err = client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{}); err != nil {
				return
			}
			data, err = json.Marshal(obj)
			return
		},
		SetJSON: func(ctx context.Context, client *kubernetes.Clientset, namespace, name string, data []byte) (err error) {
			var obj corev1.Secret
			if err = json.Unmarshal(data, &obj); err != nil {
				return
			}
			obj.Namespace = namespace
			obj.Name = name

			var current *corev1.Secret
			if current, err = client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{}); err != nil {
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

			if _, err = client.CoreV1().Secrets(namespace).Update(ctx, &obj, metav1.UpdateOptions{}); err != nil {
				if errors.IsNotFound(err) {
					obj.ResourceVersion = ""
					if _, err = client.CoreV1().Secrets(namespace).Create(ctx, &obj, metav1.CreateOptions{}); err != nil {
						return
					}
				}
				return
			}
			return
		},
	})
	knownResourceNames = append(knownResourceNames, "secret")
}
