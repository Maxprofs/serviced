// Copyright 2014, The Serviced Authors. All rights reserved.
// Use of this source code is governed by a
// license that can be found in the LICENSE file.

package web

import (
	"github.com/zenoss/glog"
	"github.com/zenoss/go-json-rest"
	"github.com/zenoss/serviced/domain/host"
	"github.com/zenoss/serviced/rpc/agent"

	"net/url"
	"strings"
)

//restGetHosts gets all hosts. Response is map[host-id]host.Host
func restGetHosts(w *rest.ResponseWriter, r *rest.Request, ctx *requestContext) {
	response := make(map[string]*host.Host)
	client, err := ctx.getMasterClient()
	if err != nil {
		restServerError(w)
		return
	}

	hosts, err := client.GetHosts()
	if err != nil {
		glog.Errorf("Could not get hosts: %v", err)
		restServerError(w)
		return
	}
	glog.V(2).Infof("Returning %d hosts", len(hosts))
	for _, host := range hosts {
		response[host.ID] = host
	}

	w.WriteJson(&response)
}

//restGetHost retrieves a host. Response is Host
func restGetHost(w *rest.ResponseWriter, r *rest.Request, ctx *requestContext) {
	hostID, err := url.QueryUnescape(r.PathParam("hostId"))
	if err != nil {
		restBadRequest(w)
		return
	}

	client, err := ctx.getMasterClient()
	if err != nil {
		restServerError(w)
		return
	}

	host, err := client.GetHost(hostID)
	if err != nil {
		glog.Error("Could not get host: ", err)
		restServerError(w)
		return
	}

	glog.V(4).Infof("restGetHost: id %s, host %#v", hostID, host)
	w.WriteJson(&host)
}

//restAddHost adds a Host. Request input is host.Host
func restAddHost(w *rest.ResponseWriter, r *rest.Request, ctx *requestContext) {
	var payload host.Host
	err := r.DecodeJsonPayload(&payload)
	if err != nil {
		glog.V(1).Infof("Could not decode host payload: %v", err)
		restBadRequest(w)
		return
	}
	// Save the pool ID and IP address for later. GetInfo wipes these
	ipAddr := payload.IPAddr
	parts := strings.Split(ipAddr, ":")
	hostIP := parts[0]

	agentClient, err := agent.NewClient(payload.IPAddr)
	//	remoteClient, err := serviced.NewAgentClient(payload.IPAddr)
	if err != nil {
		glog.Errorf("Could not create connection to host %s: %v", payload.IPAddr, err)
		restServerError(w)
		return
	}

	IPs := []string{}
	for _, ip := range payload.IPs {
		IPs = append(IPs, ip.IPAddress)
	}
	buildRequest := agent.BuildHostRequest{
		IP:     hostIP,
		PoolID: payload.PoolID,
	}
	host, err := agentClient.BuildHost(buildRequest)
	if err != nil {
		glog.Errorf("Unable to get remote host info: %v", err)
		restBadRequest(w)
		return
	}
	masterClient, err := ctx.getMasterClient()
	if err != nil {
		glog.Errorf("Unable to add host: %v", err)
		restServerError(w)
		return
	}
	err = masterClient.AddHost(*host)
	if err != nil {
		glog.Errorf("Unable to add host: %v", err)
		restServerError(w)
		return
	}
	glog.V(0).Info("Added host ", host.ID)
	w.WriteJson(&simpleResponse{"Added host", hostLinks(host.ID)})
}

//restUpdateHost updates a host. Request input is host.Host
func restUpdateHost(w *rest.ResponseWriter, r *rest.Request, ctx *requestContext) {
	hostID, err := url.QueryUnescape(r.PathParam("hostId"))
	if err != nil {
		restBadRequest(w)
		return
	}
	glog.V(3).Infof("Received update request for %s", hostID)
	var payload host.Host
	err = r.DecodeJsonPayload(&payload)
	if err != nil {
		glog.V(1).Infof("Could not decode host payload: %v", err)
		restBadRequest(w)
		return
	}

	masterClient, err := ctx.getMasterClient()
	if err != nil {
		glog.Errorf("Unable to add host: %v", err)
		restServerError(w)
		return
	}
	err = masterClient.UpdateHost(payload)
	if err != nil {
		glog.Errorf("Unable to update host: %v", err)
		restServerError(w)
		return
	}
	glog.V(1).Info("Updated host ", hostID)
	w.WriteJson(&simpleResponse{"Updated host", hostLinks(hostID)})
}

//restRemoveHost removes a host using host-id
func restRemoveHost(w *rest.ResponseWriter, r *rest.Request, ctx *requestContext) {
	hostID, err := url.QueryUnescape(r.PathParam("hostId"))
	if err != nil {
		restBadRequest(w)
		return
	}

	masterClient, err := ctx.getMasterClient()
	if err != nil {
		glog.Errorf("Unable to add host: %v", err)
		restServerError(w)
		return
	}
	err = masterClient.RemoveHost(hostID)
	if err != nil {
		glog.Errorf("Could not remove host: %v", err)
		restServerError(w)
		return
	}
	glog.V(0).Info("Removed host ", hostID)
	w.WriteJson(&simpleResponse{"Removed host", hostsLinks()})
}
