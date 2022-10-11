package resources

import (
	"context"
	"errors"
	"github.com/guoyk93/gg"
	"k8s.io/client-go/kubernetes"
	"sync"
)

type Factory interface {
	Pull(ctx context.Context, client *kubernetes.Clientset, opts PullOptions) (data []byte, err error)
	Push(ctx context.Context, client *kubernetes.Clientset, opts PushOptions) (err error)
	List(ctx context.Context, client *kubernetes.Clientset, opts ListOptions) (names []string, err error)
}

var (
	factories   = map[string]Factory{}
	factoryLock = &sync.Mutex{}
)

func Register(kind string, fac Factory) {
	factoryLock.Lock()
	defer factoryLock.Unlock()
	factories[kind] = fac
}

func GetFactory(kind string) (fac Factory, err error) {
	factoryLock.Lock()
	defer factoryLock.Unlock()
	var ok bool
	fac, ok = factories[kind]
	if !ok {
		err = errors.New("missing factory for kind '" + kind + "'")
	}
	return
}

type PullOptions struct {
	Kind      string
	Namespace string
	Name      string
}

func Pull(ctx context.Context, client *kubernetes.Clientset, opts PullOptions) (data []byte, err error) {
	defer gg.Guard(&err)
	return gg.Must(GetFactory(opts.Kind)).Pull(ctx, client, opts)
}

type PushOptions struct {
	Kind      string
	Namespace string
	Name      string
	Data      []byte
}

func Push(ctx context.Context, client *kubernetes.Clientset, opts PushOptions) (err error) {
	defer gg.Guard(&err)
	return gg.Must(GetFactory(opts.Kind)).Push(ctx, client, opts)
}

type ListOptions struct {
	Kind      string
	Namespace string
}

func List(ctx context.Context, client *kubernetes.Clientset, opts ListOptions) (names []string, err error) {
	defer gg.Guard(&err)
	return gg.Must(GetFactory(opts.Kind)).List(ctx, client, opts)
}
