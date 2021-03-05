// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

// This package provides a implementation of a ProtocolDriver interface.
//
package driver

import (
	"fmt"
	"sync"
	"time"
	"net/http"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"io"
	"strconv"

	dsModels "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var once sync.Once
var driver *VirtualDriver
var mqttClient mqtt.Client

type VirtualDriver struct {
	lc            logger.LoggingClient
	asyncCh       chan<- *dsModels.AsyncValues
}

func NewProtocolDriver() dsModels.ProtocolDriver {
	once.Do(func() {
		driver = new(VirtualDriver)
	})
	return driver
}

func (d *VirtualDriver) DisconnectDevice(deviceName string, protocols map[string]models.ProtocolProperties) error {
	d.lc.Info(fmt.Sprintf("VirtualDriver.DisconnectDevice: virtual-device driver is disconnecting to %s", deviceName))
	return nil
}

func (d *VirtualDriver) Initialize(lc logger.LoggingClient, asyncCh chan<- *dsModels.AsyncValues, deviceCh chan<- []dsModels.DiscoveredDevice) error {
	d.lc = lc
	d.asyncCh = asyncCh

	opts := mqtt.NewClientOptions().AddBroker("localhost:1883")
	var MsgHandler mqtt.MessageHandler = func(client mqtt.Client, mqtt_msg mqtt.Message) {
		v := string(mqtt_msg.Payload())
		fmt.Println("TOPIC: "+mqtt_msg.Topic())
		fmt.Printf("MSG: %s\n",v)

		resp, err := http.Get("http://localhost:48081/api/v1/device/name/virtual-device-random")
		defer resp.Body.Close()

		dev_info_json, _ := ioutil.ReadAll(resp.Body)
		dev_info := make(map[string]string)
		json.Unmarshal(dev_info_json, &dev_info)
		edgex_id := dev_info["id"]

		now := strconv.FormatInt(time.Now().UnixNano() / int64(time.Millisecond),10)
		body := bytes.NewBufferString("{\"device\":\""+edgex_id+"\",\"created\":"+now+",\"origin\":"+now+",\"modified\":0,\"readings\":[{\"name\":\"Device_Data\", \"value\":\""+v+"\",\"created\":"+now+",\"origin\":"+now+",\"modified\":0}]}")

		resp, err = http.Post("http://localhost:48080/api/v1/event", "text/plain", body)
		if err != nil {
			panic(err)
		}
		if resp != nil {
			defer resp.Body.Close()
		}

		io.Copy(ioutil.Discard, resp.Body)
	}
	opts.SetDefaultPublishHandler(MsgHandler)
	mqttClient = mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	if token := mqttClient.Subscribe("device",0,nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		return nil
	}

	return nil
}

func (d *VirtualDriver) HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []dsModels.CommandRequest) (res []*dsModels.CommandValue, err error) {

	return nil, nil
}

func (d *VirtualDriver) HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []dsModels.CommandRequest,
	params []*dsModels.CommandValue) error {

	for _, param := range params {
		switch param.DeviceResourceName {
		case "Device_Command":
			v, err := param.StringValue()
			if err != nil {
				return fmt.Errorf("VirtualDriver.HandleWriteCommands: %v", err)
			}

			mqttClient.Publish("edgex", 0, false, "ID:"+deviceName+",cmd:"+v)
		default:
			return fmt.Errorf("VirtualDriver.HandleWriteCommands: there is no matched device resource for %s", param.String())
		}
	}

	return nil
}

func (d *VirtualDriver) Stop(force bool) error {
	d.lc.Info("VirtualDriver.Stop: device-random driver is stopping...")
	return nil
}

func (d *VirtualDriver) AddDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	d.lc.Debug(fmt.Sprintf("a new Device is added: %s", deviceName))
	return nil
}

func (d *VirtualDriver) UpdateDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	d.lc.Debug(fmt.Sprintf("Device %s is updated", deviceName))
	return nil
}

func (d *VirtualDriver) RemoveDevice(deviceName string, protocols map[string]models.ProtocolProperties) error {
	d.lc.Debug(fmt.Sprintf("Device %s is removed", deviceName))
	return nil
}
