package pipeline

import (
	"github.com/dmwm/CMSMonitoring/src/go/intelligence/models"
	"github.com/dmwm/CMSMonitoring/src/go/intelligence/utils"
	"log"
)

// Module     : intelligence
// Author     : Rahul Indra <indrarahul2013 AT gmail dot com>
// Created    : Wed, 1 July 2020 11:04:01 GMT
// Description: CMS MONIT infrastructure Intelligence Module

// Preprocess - function make required changes to alerts and filter only SSB and GGUS alerts
func Preprocess(data <-chan models.AmJSON) <-chan models.AmJSON {
	if utils.ConfigJSON.Server.Verbose > 0 {
		log.Println("Preprocess step")
	}
	utils.IfSilencedMap = make(map[string]utils.SilenceMapVals)

	err := updateSilencedMap()
	if err != nil {
		log.Printf("Unable to update the IfSilenced Map, error: %v\n", err)
	}

	if utils.ConfigJSON.Server.Verbose > 1 {
		log.Printf("Current IfSilenced Map has %d entries", len(utils.IfSilencedMap))
		if utils.ConfigJSON.Server.Verbose > 2 {
			log.Printf("Current IfSilenced Map: %v", utils.IfSilencedMap)
		}
	}

	preprocessedData := make(chan models.AmJSON)
	go func() {
		defer close(preprocessedData)
		for each := range data {
			if utils.ConfigJSON.Server.Verbose > 1 {
				log.Println(each.String())
			}
			for _, service := range utils.ConfigJSON.Services {

				srvLabel, ok := utils.Get(each.Labels, utils.ConfigJSON.Alerts.ServiceLabel)
				if ok && srvLabel == service.Name {
					if val, ok := utils.Get(each.Labels, utils.ConfigJSON.Alerts.UniqueLabel); ok {
						utils.DataReadWriteLock.RLock()
						_, alertFoundInSilencedMap := utils.IfSilencedMap[val]
						utils.DataReadWriteLock.RUnlock()
						if !alertFoundInSilencedMap {
							preprocessedData <- each
						}

					}
				}
			}
		}
	}()

	return preprocessedData
}

// updateSilencedMap -function for updating the ifSilenced Map to help us not to push redundant silences
func updateSilencedMap() error {

	data, err := utils.GetSilences()
	if err != nil {
		log.Printf("Unable to Update Silence Map, error: %v", err)
	}

	for _, each := range data.Data {
		if each.Status.State == utils.ConfigJSON.Silence.SilenceStatus[0] {
			utils.ChangeCounters.NoOfActiveSilences++
		}
		if each.Status.State == utils.ConfigJSON.Silence.SilenceStatus[1] {
			utils.ChangeCounters.NoOfExpiredSilences++
		}
		if each.Status.State == utils.ConfigJSON.Silence.SilenceStatus[2] {
			utils.ChangeCounters.NoOfPendingSilences++
		}
		for _, matcher := range each.Matchers {
			if matcher.Name == utils.ConfigJSON.Alerts.UniqueLabel {
				if each.Status.State == utils.ConfigJSON.Silence.SilenceStatus[0] {
					utils.DataReadWriteLock.Lock()
					utils.IfSilencedMap[matcher.Value] = utils.SilenceMapVals{IfAvail: 1, SilenceID: each.ID}
					utils.DataReadWriteLock.Unlock()
				}
			}
		}
	}

	return nil
}
