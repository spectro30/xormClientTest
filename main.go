package main

import (
	"context"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"log"
	"path/filepath"
	"flag"

	_ "github.com/go-sql-driver/mysql"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"kmodules.xyz/client-go/tools/portforward"
	"xorm.io/xorm"
)
var (
	namespace *string
	podName *string
)

func init(){
	namespace = flag.String("ns", "demo", "Namespace Name")
	podName = flag.String("pod", "demo-quickstart-0", "Pod Name")
	flag.Parse()

}

func main() {
	masterURL := ""
	kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
	if err != nil {
		log.Fatalf("Could not get Kubernetes config: %s", err)
	}

	kc := kubernetes.NewForConfigOrDie(config)

	pod, err := kc.CoreV1().Pods("demo").Get(context.TODO(), "demo-quickstart-0", metav1.GetOptions{})
	if err != nil {
		fmt.Println("Error in Get Pods")
		panic(err)
	}


	fmt.Println(".......................podName", pod.Name)
	tnl := portforward.NewTunnel(
		kc.CoreV1().RESTClient(),
		config,
		*namespace,
		*podName,
		3306,
	)
	err = tnl.ForwardPort()
	if err != nil {
		fmt.Println("Error in port forward failed")
		panic(err)
	}

	defer tnl.Close()
	secret, err := kc.CoreV1().Secrets("demo").Get(context.TODO(), "demo-quickstart-auth", metav1.GetOptions{})
	if err != nil {
		fmt.Println("Error in Get Secrets")
		panic(err)
	}
	userName := string(secret.Data["username"])
	pass := string(secret.Data["password"])

	fmt.Println("user, pass=", userName, pass)

	spew.Dump(tnl.Local)

	cnnstr := fmt.Sprintf("%v:%v@tcp(127.0.0.1:%v)/%s", userName, pass, tnl.Local, "mysql")
	en, err := xorm.NewEngine("mysql", cnnstr)
	if err != nil {
		fmt.Println("Error in Start Engine")
		panic(err)
	}
	en.ShowSQL(true)
	fmt.Println("......................................1")
	err = en.Ping()
	if err != nil {
		fmt.Println("Error in ping")
		panic(err)
	}
	fmt.Println("......................................2")
	result, err := en.QueryString("SHOW STATUS LIKE 'wsrep_cluster_size';")
	if err != nil {
		fmt.Println("Error in Query")
		panic(err)
	}
	res := result[0]
	x, ok := res["Value"]
	fmt.Println(x)
	if !ok{

	}
	//spew.Dump(result)
}
