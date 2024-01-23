package focal

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CommonObject interface {
	client.Object
	composed.ObjWithConditions

	ScopeRef() cloudresourcesv1beta1.ScopeRef
	SetScopeRef(scopeRef cloudresourcesv1beta1.ScopeRef)
}

type OperationType int

const (
	NONE OperationType = iota
	ADD
	MODIFY
	DELETE
)
