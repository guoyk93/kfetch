package resources

import (
	"context"
	"encoding/json"
	"github.com/guoyk93/gg"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type API[T any, TL any] interface {
	List(ctx context.Context, opts metav1.ListOptions) (*TL, error)
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*T, error)
	Update(ctx context.Context, obj *T, opts metav1.UpdateOptions) (*T, error)
	Create(ctx context.Context, obj *T, opts metav1.CreateOptions) (*T, error)
}

type FactoryWithAPI[T any, TL any] struct {
	FuncCreateAPI gg.F21[*kubernetes.Clientset, string, API[T, TL]]
	FuncAfterList gg.F21[*TL, ListOptions, []string]
	FuncAfterPull gg.F20[*T, PullOptions]
}

func (f *FactoryWithAPI[T, TL]) Pull(ctx context.Context, client *kubernetes.Clientset, opts PullOptions) (data []byte, err error) {
	defer gg.Guard(&err)
	api := f.FuncCreateAPI(client, opts.Namespace)
	obj := gg.Must(api.Get(ctx, opts.Name, metav1.GetOptions{}))
	f.FuncAfterPull(obj, opts)
	data = gg.Must(json.Marshal(obj))
	data = gg.Must(JSON2YAML(data))
	return
}

func (f *FactoryWithAPI[T, TL]) Push(ctx context.Context, client *kubernetes.Clientset, opts PushOptions) (err error) {
	defer gg.Guard(&err)

	opts.Data = gg.Must(YAML2JSON(opts.Data))

	var obj T
	gg.Must0(json.Unmarshal(opts.Data, &obj))

	api := f.FuncCreateAPI(client, opts.Namespace)

	var curr *T
	if curr, err = api.Get(ctx, opts.Name, metav1.GetOptions{}); err != nil {
		if errors.IsNotFound(err) {
			err = nil
		} else {
			return
		}
	} else {
	}

	return
}

func (f *FactoryWithAPI[T, TL]) List(ctx context.Context, client *kubernetes.Clientset, opts ListOptions) (names []string, err error) {
	defer gg.Guard(&err)
	api := f.FuncCreateAPI(client, opts.Namespace)
	items := gg.Must(api.List(ctx, metav1.ListOptions{}))
	names = f.FuncAfterList(items, opts)
	return
}
