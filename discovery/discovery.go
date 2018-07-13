// 
// Description
//
package main

import (
    "fmt"
    "time"
    "flag"

    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    kblabels "k8s.io/apimachinery/pkg/labels"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"

    "net/http"
    "io/ioutil"
    "os"
    "bytes"
)

func main() {
    var reloadUrl string
    var sourceFilename string
    var targetFilename string
    var namespace string
    var labelName string
    var labelValue string
    flag.StringVar(&reloadUrl, "reloadurl", "http://localhost:9090/-/reload", "URL to use to reload configuration")
    flag.StringVar(&sourceFilename, "source", "/etc/prometheus/prometheus.yaml.base", "File to use as a source for the prometheus.yaml file")
    flag.StringVar(&targetFilename, "target", "/etc/prometheus/prometheus.yaml", "File to use as a target for the prometheus.yaml file")
    flag.StringVar(&namespace, "namespace", "", "K8S namespace to search in for the endpoints")
    flag.StringVar(&labelName, "labelname", "", "K8S label name to search for")
    flag.StringVar(&labelValue, "labelvalue", "", "K8S label value to search for")
    flag.Parse()

    promconfig, err := ioutil.ReadFile(sourceFilename)
    if err != nil {
        panic(err.Error())      
    }

    if labelName == "" || labelValue == "" {
        panic("labelname and labelvalue must be specified ")
    }
 

    // creates the in-cluster config
    config, err := rest.InClusterConfig()
    if err != nil {
        panic(err.Error())
    }
    // creates the clientset
    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        panic(err.Error())
    }

    lsel := kblabels.Set{labelName: labelValue}.AsSelector()
    listOptions := metav1.ListOptions{LabelSelector : lsel.String()}

    for {
            proxies, err := clientset.CoreV1().Endpoints(namespace).List(listOptions)
        if err != nil {
            panic(err.Error())
        }

        f, err := os.Create(targetFilename)
        if err != nil {
            fmt.Print("Error creating new config: %s\n",err)
        }
        _, err = f.Write(promconfig)
        if err != nil {
            fmt.Print("Error writing to new config: %s\n", err)
        }
        // loop over all proxy items
        for _,v := range proxies.Items {
            for _,s := range v.Subsets {
                for _,a := range s.Addresses {
                    clientUrl := fmt.Sprintf("http://%s:%d/clients", a.IP, s.Ports[0].Port)
                    resp, err := http.Get(clientUrl)
                    if err != nil {
                        fmt.Printf("Error getting clients from %s : %s\n", clientUrl, err)
                    } else {
                        defer resp.Body.Close()
                        body, err := ioutil.ReadAll(resp.Body)
                        if err != nil {
                            fmt.Printf("Error reading body response from %s : %s\n", clientUrl, err)
                        } else {
                            ioutil.WriteFile(fmt.Sprintf("/prometheus/%s.json",a.IP), body, 0644);
                            f.WriteString(fmt.Sprintf("  - job_name: '%s'\n", a.IP))
                            f.WriteString(fmt.Sprintf("    file_sd_configs:\n"));
                            f.WriteString(fmt.Sprintf("      - files: ['/prometheus/%s.json']\n", a.IP));
                            f.WriteString(fmt.Sprintf("      proxy_url: http://%s:%d/\n\n",a.IP, s.Ports[0].Port))
                        }
                    }
                }
            }
        }
        f.Sync()
        f.Close()
        http.Post(reloadUrl, "text/plain",  bytes.NewBuffer([]byte{}))
        time.Sleep(60 * time.Second)
    }
}
