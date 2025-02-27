# ols-load-generator
Load generator tool for openshift lightspeed service (OLS).

## **Prerequisites**
### **Running on openshift cluster**
OLS deployed on openshift cluster. Please refer to the instructions [here](https://github.com/openshift/lightspeed-operator?tab=readme-ov-file#running-on-the-cluster).
### **Running on local machine (Optional)**
A running instance of OLS (to test). Please refer to the instructions [here](https://github.com/openshift/lightspeed-service?tab=readme-ov-file#installation).

## **Installation**
```
make build; make install
```
> **NOTE**: You might want to add **sudo** to the install command as it involves creating `ols-load-generator` binary in your $PATH.   

If running on openshift simply run `make` which will build and push the image to specified image registry.

## **Usage**
### **Usage on openshift platform**
Once we have OLS deployed and running on openshift cluster, in order to trigger the load test follow the below steps.


#### **Create a secret with your cluster kubeconfig**
```
oc create secret generic kubeconfig-secret --from-file=kubeconfig=<YOUR-KUBECONFIG-PATH> -n ols-load-generator
```

#### **Deploy [config/ols-load-generator.yaml](https://github.com/openshift/ols-load-generator/blob/main/config/ols-load-generator.yaml) replaced with your corresponding envs values onto the cluster**

#### **Envs**
* `OLS_TEST_HOST` - String indicating OLS endpoint to perform load testing.
* `OLS_TEST_UUID`(Optional) - String specifying an unique ID. Will be helpful while comparing two runs or while looking at a specific run results.
* `OLS_TEST_AUTH_TOKEN` - OLS auth token string.
* `OLS_TEST_DURATION` - Load testing duration on each API endpoint.
* `OLS_TEST_WORKERS` - Amount of parallel workers to trigger load on OLS. This will basically help us send requests in parallel.
* `OLS_TEST_METRIC_STEP` - Step size for the cluster prometheus metrics to be captured.
* `OLS_TEST_PROFILES` - List of metric profiles that contain queries to be executed on prometheus.
* `OLS_TEST_ES_HOST`(Optional) - Elastic search host url. If not specified metrics will be indexed locally.
* `OLS_TEST_ES_INDEX`(Optional) - Elastic search index name to store the data. If not specified metrics will be indexed locally.
* `OLS_TEST_QUERY_ONLY`(Optional) - Flag to enable load tests only on `/v1/query` and `/v1/streaming_query` endpoints.

#### **Example Usage**
```
oc apply -f ~/config/ols-load-generator.yaml
```
Once applied it will create a job in the specified namespace and will start running the tests with above mentioned values. We can tail the logs in order to look at the benchmark results.

### **Usage on Local Machine**
```
NAME:
   ols-load-generator - A command-line tool to load test openshift lightspeed service (OLS).

USAGE:
   ols-load-generator [global options] command [command options] [arguments...]

DESCRIPTION:
   A command-line tool to load test openshift lightspeed service (OLS).

COMMANDS:
   attack   ols-load-generator attack
   index    ols-load-generator index
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   -D          print debugging logs (default: false)
   -W          quieter log output (default: false)
   --help, -h  show help
```
### attack sub-command
Subcommand to trigger the load test on openshift lightspeed service.
```
NAME:
   ols-load-generator attack - ols-load-generator attack

USAGE:
   ols-load-generator attack [command options] [arguments...]

DESCRIPTION:
   perform attack on ols endpoints

OPTIONS:
   --host value        --host localhost:6060 (default: "http://localhost:6060") [$OLS_TEST_HOST]
   --authtoken value   --authtoken authtoken [$OLS_TEST_AUTH_TOKEN]
   --uuid value        --uuid f519d9b2-aa62-44ab-9ce8-4156b712f6d2 (default: "76d8f64d-2bf3-49ac-82c3-22011ddc2284") [$OLS_TEST_UUID]
   --duration value    --duration 1m (default: 1m0s) [$OLS_TEST_DURATION]
   --workers value     --workers 10 (default: 10) [$OLS_TEST_WORKERS]
   --eshost value      --eshost eshosturl [$OLS_TEST_ES_HOST]
   --esindex value     --esindex esindex [$OLS_TEST_ES_INDEX]
   --metricstep value  --metricstep 30 (default: 30) [$OLS_TEST_METRIC_STEP]
   --profiles value    --profiles metrics.yaml,metrics-report.yaml (default: "attacker/assets/profiles/metrics-report.yaml,attacker/assets/profiles/metrics-timeseries.yaml") [$OLS_TEST_PROFILES]
   --queryonly         --query (default: false) [$OLS_TEST_QUERY_ONLY]
   --help, -h          show help
```
#### Example Usage
```
export KUBECONFIG=<your-kubeconfig-path>
ols-load-generator attack --host https://127.0.0.1:9001 --uuid random-uuid --authtoken 'auth-token' --duration 1m --workers 10
```
### index sub-command
Subcommand to scrape and index cluster prometheus metrics within a given time range. 
```
NAME:
   ols-load-generator index - ols-load-generator index

USAGE:
   ols-load-generator index [command options] [arguments...]

DESCRIPTION:
   Indexes metrics within given timerange

OPTIONS:
   --uuid value        --uuid f519d9b2-aa62-44ab-9ce8-4156b712f6d2 (default: "3c2e0d71-07f6-4bfa-b604-61a3243bbec7") [$OLS_TEST_UUID]
   --eshost value      --eshost eshosturl [$OLS_TEST_ES_HOST]
   --esindex value     --esindex esindex [$OLS_TEST_ES_INDEX]
   --metricstep value  --metricstep 30 (default: 30) [$OLS_TEST_METRIC_STEP]
   --start value       --start 1720410990 (default: 1720482385) [$OLS_TEST_START]
   --end value         --end 1720470990 (default: 1720485985) [$OLS_TEST_END]
   --profiles value    --profiles metrics.yaml,metrics-report.yaml (default: "attacker/assets/profiles/metrics-report.yaml,attacker/assets/profiles/metrics-timeseries.yaml") [$OLS_TEST_PROFILES]
   --queryonly         --query (default: false) [$OLS_TEST_QUERY_ONLY]
   --help, -h          show help
```
#### Example Usage
```
export KUBECONFIG=<your-kubeconfig-path>
ols-load-generator index --profiles "metrics.yaml" --start 1720410990 --end 1720410990
```