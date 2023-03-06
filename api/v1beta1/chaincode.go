package v1beta1

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
