package main

import (
	"errors"

	"github.com/lucasmachadolopes/cadencePoc/workflows"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/worker"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/transport/tchannel"
	"go.uber.org/zap"
)

const (
	serviceNameCadenceClient   = "cadence-client"
	serviceNameCadenceFrontend = "cadence-frontend"
)

func NewWorkflowClient() (workflowserviceclient.Interface, error) {
	ch, err := tchannel.NewChannelTransport(tchannel.ServiceName(serviceNameCadenceClient))
	if err != nil {
		return nil, err
	}
	dispatcher := yarpc.NewDispatcher(yarpc.Config{
		Name: serviceNameCadenceClient,
		Outbounds: yarpc.Outbounds{
			serviceNameCadenceFrontend: {Unary: ch.NewSingleOutbound("127.0.0.1:7933")},
		},
	})

	if dispatcher == nil {
		return nil, errors.New("failed to create dispatcher")
	}

	if err := dispatcher.Start(); err != nil {
		panic(err)
	}

	return workflowserviceclient.New(dispatcher.ClientConfig(serviceNameCadenceFrontend)), nil
}

func main() {

	wfClient, err := NewWorkflowClient()
	if err != nil {
		panic(err)
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	w := worker.New(wfClient, "cadence-poc", "pocTasklist", worker.Options{
		Logger: logger,
	})

	w.RegisterWorkflow(workflows.HelloWorldWorkflow)

	err = w.Run()

	if err != nil {
		panic(err)
	}
}
