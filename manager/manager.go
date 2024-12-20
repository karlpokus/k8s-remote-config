package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/karlpokus/k8s-remote-config/conf"
)

const NAMESPACE = "test"
const CONF_MAP_NAME = "server"
const CONF_MAP_FILE_NAME = "config"
const DEPLOYMENT_NAME = "server"
const DEPLOYMENT_KEY = "kubectl.kubernetes.io/rateUpdateAt"

var host = flag.String("h", "localhost", "HTTP host")
var port = flag.String("p", "7000", "HTTP port")

func main() {
	flag.Parse()
	err := startServer(net.JoinHostPort(*host, *port))
	if err != nil {
		log.Printf("server start err: %v", err)
	}
}

func startServer(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/update", updateHandler)
	mux.HandleFunc("/health", healthHandler)
	log.Printf("starting HTTP server on %s", addr)
	return http.ListenAndServe(addr, mux)
}

func parseQuery(r *http.Request) (string, string, error) {
	q := r.URL.Query()
	k := q.Get("k")
	v := q.Get("v")
	if k == "" || v == "" {
		return "", "", fmt.Errorf("queryparams k or v missing")
	}
	return k, v, nil
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	// log request
	log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
	k, v, err := parseQuery(r)
	if err != nil {
		// expose err during testing
		http.Error(w, err.Error(), 400)
		return
	}
	// TODO: cache client
	c, err := k8sClient()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	err = updateConfig(ctx, k, v, c)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	err = updateDeployment(ctx, c)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	fmt.Fprintf(w, "%s set to %v\n", k, v)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "ok")
}

func k8sClient() (*kubernetes.Clientset, error) {
	fpath := path.Join(os.Getenv("HOME"), ".kube", "config")
	if _, err := os.Stat(fpath); err != nil && os.IsNotExist(err) {
		fpath = ""
	}
	// fallbacks to inClusterConfig if both inputs are empty
	//
	// Note! This wrapper is chatty. Consider switching
	// to rest.InClusterConf instead.
	conf, err := clientcmd.BuildConfigFromFlags("", fpath)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(conf)
}

func updateConfig(ctx context.Context, k, v string, client *kubernetes.Clientset) error {
	cm, err := client.CoreV1().ConfigMaps(NAMESPACE).Get(ctx, CONF_MAP_NAME, metav1.GetOptions{})
	if err != nil {
		return err
	}
	s, ok := cm.Data[CONF_MAP_FILE_NAME]
	if !ok {
		return fmt.Errorf("no value for key: %s in configmap", CONF_MAP_FILE_NAME)
	}
	c := conf.Marshal([]byte(s))
	c[k] = v
	cm.Data[CONF_MAP_FILE_NAME] = conf.Unmarshal(c)
	_, err = client.CoreV1().ConfigMaps(NAMESPACE).Update(ctx, cm, metav1.UpdateOptions{})
	return err
}

func updateDeployment(ctx context.Context, client *kubernetes.Clientset) error {
	d, err := client.AppsV1().Deployments(NAMESPACE).Get(ctx, DEPLOYMENT_NAME, metav1.GetOptions{})
	if err != nil {
		return err
	}
	// Note!
	//
	// d.Spec.Template.Annotations is bogus
	// use d.Spec.Template.ObjectMeta.Annotations
	an := d.Spec.Template.ObjectMeta.Annotations
	if an == nil {
		an = make(map[string]string)
	}
	an[DEPLOYMENT_KEY] = metav1.Now().Format(time.RFC3339)
	d.Spec.Template.ObjectMeta.Annotations = an
	_, err = client.AppsV1().Deployments(NAMESPACE).Update(ctx, d, metav1.UpdateOptions{})
	return err
}
