package full

import (
	"open_im_sdk/internal/friend"
	"open_im_sdk/internal/group"
	"open_im_sdk/internal/user"
	"open_im_sdk/open_im_sdk_callback"
	"open_im_sdk/pkg/common"
	sdk "open_im_sdk/pkg/sdk_params_callback"
	api "open_im_sdk/pkg/server_api_params"
)

type Full struct {
	user   *user.User
	friend *friend.Friend
	group  *group.Group
}

func NewFull(user *user.User, friend *friend.Friend, group *group.Group) *Full {
	return &Full{user: user, friend: friend, group: group}
}

func (u *Full) getUsersInfo(callback open_im_sdk_callback.Base, userIDList sdk.GetUsersInfoParam, operationID string) sdk.GetUsersInfoCallback {
	friendList := u.friend.GetDesignatedFriendListInfo(callback, []string(userIDList), operationID)
	blackList := u.friend.GetDesignatedBlackListInfo(callback, []string(userIDList), operationID)
	notIn := make([]string, 0)
	for _, v := range userIDList {
		inFriendList := 0
		for _, friend := range friendList {
			if v == friend.FriendUserID {
				inFriendList = 1
				break
			}
		}
		inBlackList := 0
		for _, black := range blackList {
			if v == black.BlockUserID {
				inBlackList = 1
				break
			}
		}
		if inFriendList == 0 && inBlackList == 0 {
			notIn = append(notIn, v)
		}
	}
	//from svr
	publicList := make([]*api.PublicUserInfo, 0)
	if len(notIn) > 0 {
		publicList = u.user.GetUsersInfoFromSvr(callback, notIn, operationID)
	}

	return common.MergeUserResult(publicList, friendList, blackList)
}
