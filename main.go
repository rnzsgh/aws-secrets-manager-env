package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"

	log "github.com/golang/glog"
)

var secretNames arrayFlags

func init() {
	flag.Var(&secretNames, "secret", "Some description for this param.")
	flag.Parse()
	flag.Lookup("logtostderr").Value.Set("true")
}

func main() {

	fmt.Println(os.Args)

	if len(os.Args) < 2 {
		log.Fatalf("You must pass the name of the secrets to load - e.g., aws-secrets-manager-env prod/appA prod/common")
	}

	/*

		region := region()
		if len(region) == 0 {
			log.Infof("Outside of AWS - not looking up or setting environment variables from Secrets Manager")
			return
		}

	*/

	region := "us-east-1"

	log.Infof("region %s", region)

	svc := secretsmanager.New(
		session.New(),
		aws.NewConfig().WithRegion(region).WithMaxRetries(3),
	)

	for _, s := range secretNames {
		log.Infof("secret: %s", s)
		//log.Infof("Loading - secret: %s", os.Args[i])
		if err := secret(svc, s); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

	}

	/*
		for i := 0; i < len(os.Args); i++ {
			log.Infof("Loading - secret: %s", os.Args[i])
				if err := secret(svc, os.Args[i]); err != nil {
					log.Fatal(err)
					os.Exit(1)
				}
		}
	*/

	log.Flush()
}

func secret(svc *secretsmanager.SecretsManager, name string) error {

	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)

	result, err := svc.GetSecretValueWithContext(ctx, &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(name),
		VersionStage: aws.String("AWSCURRENT"),
	})

	if err != nil {
		return fmt.Errorf("Unable to load - secret: %s - reason: %v", name, err)
	}

	m := make(map[string]interface{})

	if err = json.Unmarshal([]byte(*result.SecretString), &m); err != nil {
		return fmt.Errorf("Unable to parse json - secret: %s - reason: %v:", name, err)
	}

	for k, v := range m {
		if err = os.Setenv(k, v.(string)); err != nil {
			return fmt.Errorf("Unable to set environment variable - secret: %s - reason: %v:", name, err)
		} else {
			log.Infof("Setting environment variable - secret: %s - key: %s", name, k)
		}
	}

	return nil
}

// region returns AWS region or an empty string if it is called outside of aws.
func region() string {

	client := http.Client{Timeout: time.Duration(40 * time.Millisecond)}
	r, err := client.Get("http://169.254.169.254/latest/dynamic/instance-identity/document")
	if err != nil {
		log.Error(err)
		return ""
	}
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		panic(fmt.Errorf("Unable to read region lookup body - reason: %v", err))
	}

	m := make(map[string]interface{})

	if err = json.Unmarshal(body, &m); err != nil {
		panic(fmt.Errorf("Unable to parse region json - reason: %v:", err))
	}

	return m["region"].(string)
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return ""
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}
