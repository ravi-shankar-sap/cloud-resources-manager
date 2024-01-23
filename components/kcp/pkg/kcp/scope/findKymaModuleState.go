package scope

import (
	"context"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/util"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
)

func findKymaModuleState(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	state.moduleState = util.GetKymaModuleState(state.kyma, "cloud-manager")

	if state.moduleState != util.KymaModuleStateReady {
		return composed.StopAndForget, nil
	}

	return nil, nil
}
