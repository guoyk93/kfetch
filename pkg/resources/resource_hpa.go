package resources

import (
	"context"
	"encoding/json"
	"github.com/guoyk93/kfetch"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
)

func init() {
	knownResources = append(knownResources, &Resource{
		Kind: "hpa",
		List: func(ctx context.Context, client *kubernetes.Clientset, namespace string) (names []string, err error) {
			var items *autoscalingv2beta2.HorizontalPodAutoscalerList
			if items, err = client.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace).List(ctx, metav1.ListOptions{}); err != nil {
				return
			}
			for _, item := range items.Items {
				names = append(names, item.Name)
			}
			return
		},
		GetJSON: func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (data []byte, err error) {
			var obj *autoscalingv2beta2.HorizontalPodAutoscaler
			if obj, err = client.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace).Get(ctx, name, metav1.GetOptions{}); err != nil {
				return
			}
			data, err = json.Marshal(obj)
			return
		},
		SetJSON: func(ctx context.Context, client *kubernetes.Clientset, namespace, name string, data []byte) (err error) {
			var obj autoscalingv2beta2.HorizontalPodAutoscaler
			if err = json.Unmarshal(data, &obj); err != nil {
				return
			}
			obj.Namespace = namespace
			obj.Name = name

			var current *autoscalingv2beta2.HorizontalPodAutoscaler
			if current, err = client.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace).Get(ctx, name, metav1.GetOptions{}); err != nil {
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

			if _, err = client.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace).Update(ctx, &obj, metav1.UpdateOptions{}); err != nil {
				if errors.IsNotFound(err) {
					obj.ResourceVersion = ""
					if _, err = client.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace).Create(ctx, &obj, metav1.CreateOptions{}); err != nil {
						return
					}
				}
				return
			}
			return
		},
	})
	knownResourceNames = append(knownResourceNames, "hpa")
}
