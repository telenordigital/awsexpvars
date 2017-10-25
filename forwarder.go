package main

/*
**   Copyright 2017 Telenor Digital AS
**
**  Licensed under the Apache License, Version 2.0 (the "License");
**  you may not use this file except in compliance with the License.
**  You may obtain a copy of the License at
**
**      http://www.apache.org/licenses/LICENSE-2.0
**
**  Unless required by applicable law or agreed to in writing, software
**  distributed under the License is distributed on an "AS IS" BASIS,
**  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
**  See the License for the specific language governing permissions and
**  limitations under the License.
 */
import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

// Forwarder is documented
type Forwarder struct {
	sourceURI  string
	filter     *Filter
	cw         *cloudwatch.CloudWatch
	metricName string
	instanceID string
}

// NewForwarder is documented
func NewForwarder(uri string, filter *Filter, metricName string, instanceID string) (Forwarder, error) {
	awsSession, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return Forwarder{}, err
	}
	return Forwarder{uri, filter, cloudwatch.New(awsSession), metricName, instanceID}, nil
}

// ReadAndForward will forward metrics to CloudWatch. If there's an error
// it returns false.
func (f *Forwarder) ReadAndForward() error {
	// Read the endpoint, update AWS
	resp, err := http.Get(f.sourceURI)
	if err != nil {
		return fmt.Errorf("unable to read the expvar endpoint: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned %d", expvarURI, resp.StatusCode)
	}

	values := make(map[string]interface{})
	if err := json.NewDecoder(resp.Body).Decode(&values); err != nil {
		return err
	}
	for k, v := range values {
		if f.filter.Include(k) {
			v, ok := v.(float64)
			if !ok {
				return fmt.Errorf("value isn't what I expected: %s = %v (type is %T)", k, v, v)
			}
			_, err := f.cw.PutMetricData(&cloudwatch.PutMetricDataInput{
				Namespace: aws.String(f.metricName),
				MetricData: []*cloudwatch.MetricDatum{
					&cloudwatch.MetricDatum{
						MetricName: aws.String(k),
						Unit:       aws.String(cloudwatch.StandardUnitNone),
						Value:      aws.Float64(v),
						Dimensions: []*cloudwatch.Dimension{
							&cloudwatch.Dimension{
								Name:  aws.String("InstanceId"),
								Value: aws.String(f.instanceID),
							},
						},
					},
				}})
			if err != nil {
				return fmt.Errorf("unable to put metric data: %v", err)
			}
		}
	}
	return nil
}
