package api_common

import (
	"net/http"
	"pandora-pay/config/config_reward"
	"pandora-pay/config/config_stake"
)

type APIStakingInfoRequest struct {
	Height uint64 `json:"height,omitempty" msgpack:"height,omitempty"`
}

type APIStakingInfoReply struct {
	BlockReward        uint64 `json:"blockReward" msgpack:"blockReward"`
	RequiredStake      uint64 `json:"requiredStake" msgpack:"requiredStake"`
	PendingStakeWindow uint64 `json:"pendingStakeWindow" msgpack:"pendingStakeWindow"`
}

func (api *APICommon) GetStakingInfo(r *http.Request, args *APIStakingInfoRequest, reply *APIStakingInfoReply) error {

	reply.BlockReward = config_reward.GetRewardAt(args.Height)
	reply.RequiredStake = config_stake.GetRequiredStake(args.Height)
	reply.PendingStakeWindow = config_stake.GetPendingStakeWindow(args.Height)

	return nil
}
