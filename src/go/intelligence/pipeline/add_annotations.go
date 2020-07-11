package pipeline

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/intelligence/models"
	"go/intelligence/utils"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

// Module     : intelligence
// Author     : Rahul Indra <indrarahul2013 AT gmail dot com>
// Created    : Wed, 1 July 2020 11:04:01 GMT
// Description: CMS MONIT infrastructure Intelligence Module

//AddAnnotation function for adding annotations to dashboards
func AddAnnotation(data <-chan models.AmJSON) <-chan models.AmJSON {

	dataAfterAnnotation := make(chan models.AmJSON)

	go func() {
		dashboards := utils.FindDashboard()
		for each := range data {
			for _, service := range utils.ConfigJSON.Services {
				if each.Labels[utils.ConfigJSON.Alerts.ServiceLabel] == service.Name {
					for _, action := range service.AnnotationMap.Actions {
						if val, ok := each.Annotations[service.AnnotationMap.Label].(string); ok {
							if strings.Contains(strings.ToLower(val), strings.ToLower(action)) {
								for _, system := range service.AnnotationMap.Systems {
									if val, ok := each.Annotations[service.AnnotationMap.Label].(string); ok {
										if strings.Contains(strings.ToLower(val), strings.ToLower(system)) {
											for _, dashboard := range dashboards {
												var dashboardData models.GrafanaDashboard
												if val, ok := dashboard["id"].(int); ok {
													dashboardData.DashboardID = val
												}
												dashboardData.Time = each.StartsAt.Unix() * 1000
												dashboardData.TimeEnd = each.EndsAt.Unix() * 1000
												dashboardData.Tags = utils.ParseTags()
												if val, ok := each.Annotations["shortDescription"].(string); ok {
													dashboardData.Text = val
												}
												dData, err := json.Marshal(dashboardData)
												if err != nil {
													log.Printf("Unable to convert the data into JSON %v, error: %v\n", dashboardData, err)
												}
												if utils.ConfigJSON.Server.Verbose > 0 {
													log.Printf("Annotation: %v", dashboardData)
												}
												addAnnotationHelper(dData)
											}
										}
									}
								}
							}
						}
					}
				}
			}
			dataAfterAnnotation <- each
		}

		close(dataAfterAnnotation)
	}()
	return dataAfterAnnotation
}

//addAnnotationHelper helper function
// The following block of code was taken from
// https://github.com/dmwm/CMSMonitoring/blob/master/src/go/MONIT/monit.go#L639
func addAnnotationHelper(data []byte) {
	var headers [][]string
	bearer := fmt.Sprintf("Bearer %s", utils.ConfigJSON.Annotation.Token)
	h := []string{"Authorization", bearer}
	headers = append(headers, h)
	h = []string{"Content-Type", "application/json"}
	headers = append(headers, h)

	apiURL := utils.ValidateURL(utils.ConfigJSON.Annotation.URL, utils.ConfigJSON.Annotation.AnnotationAPI)

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(data))
	if err != nil {
		log.Printf("Unable to make request to %s, error: %s", apiURL, err)
	}
	for _, v := range headers {
		if len(v) == 2 {
			req.Header.Add(v[0], v[1])
		}
	}

	if utils.ConfigJSON.Server.Verbose > 1 {
		dump, err := httputil.DumpRequestOut(req, true)
		if err == nil {
			log.Println("request: ", string(dump))
		}
	}

	timeout := time.Duration(utils.ConfigJSON.Server.HTTPTimeout) * time.Second
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)

	if err != nil {
		log.Printf("Unable to get response from %s, error: %s", apiURL, err)
	}
	if utils.ConfigJSON.Server.Verbose > 1 {
		dump, err := httputil.DumpResponse(resp, true)
		if err == nil {
			log.Println("response:", string(dump))
		}
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if utils.ConfigJSON.Server.Verbose > 1 {
		log.Println("response Status:", resp.Status)
		log.Println("response Headers:", resp.Header)
		log.Println("response Body:", string(body))
	}
}
