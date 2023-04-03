package v1beta1

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// ChaincodeProposalLabel the main purpose of this label is to make it easier to clean up the proposal when the chaincode is deleted.
	ChaincodeProposalLabel         = "bestchains.chaincode.delete.proposal"
	ChaincodeChannelLabel          = "bestchains.chaincode.channel"
	ChaincodeIDLabel               = "bestchains.chaincode.id"
	ChaincodeVersionLabel          = "bestchains.chaincode.version"
	ChaincodeUsedEndorsementPolicy = "bestchians.chaincode.endorsementpolicy"
)

var condOrder = []ChaincodeConditionType{ChaincodeCondDone, ChaincodeCondPackaged, ChaincodeCondInstalled,
	ChaincodeCondApproved, ChaincodeCondCommitted, ChaincodeCondRunning}

func NextCond(instance *Chaincode) ChaincodeConditionType {

	conditions := instance.Status.Conditions
	exp := 1
	if len(conditions) == 0 {
		return condOrder[exp]
	}

	lastCond := conditions[len(conditions)-1]
	nextAction := lastCond.Type
	if nextAction == ChaincodeCondError {
		return lastCond.NextStage
	}

	for ; exp < len(condOrder); exp++ {
		if condOrder[exp] == nextAction {
			exp = (exp + 1) % len(condOrder)
			break
		}
	}
	return condOrder[exp]
}

// GetChannel
func (i *Chaincode) GetChannel(r client.Reader) (*Channel, error) {
	ch := &Channel{}
	if err := r.Get(context.TODO(), types.NamespacedName{Name: i.Spec.Channel}, ch); err != nil {
		return nil, err
	}
	return ch, nil
}
