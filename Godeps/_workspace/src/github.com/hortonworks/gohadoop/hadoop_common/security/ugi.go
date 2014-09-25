package security

import (
	"github.com/hortonworks/gohadoop/hadoop_common"
	"log"
	"os/user"
	"sync"
)

/** a (very) basic UserGroupInformation implementation for storing user data/tokens,
  This implementation is currently *not* thread-safe
*/

type UserGroupInformation struct {
	rwMutex    sync.RWMutex
	userInfo   *hadoop_common.UserInformationProto
	userTokens []*hadoop_common.TokenProto
}

var once sync.Once
var currentUserGroupInformation *UserGroupInformation
var maxTokens = 16

func CreateCurrentUserInfoProto() (*hadoop_common.UserInformationProto, error) {
	// Figure the current user-name
	var username string
	if currentUser, err := user.Current(); err != nil {
		log.Fatal("user.Current", err)
		return nil, err
	} else {
		username = currentUser.Username
	}

	return &hadoop_common.UserInformationProto{EffectiveUser: nil, RealUser: &username}, nil
}

func Allocate(userInfo *hadoop_common.UserInformationProto, userTokens []*hadoop_common.TokenProto) *UserGroupInformation {
	ugi := new(UserGroupInformation)

	if userInfo != nil {
		ugi.userInfo = userInfo
	} else {
		currentUserInfo, _ := CreateCurrentUserInfoProto()
		ugi.userInfo = currentUserInfo
	}

	if userTokens != nil {
		ugi.userTokens = userTokens
	} else {
		ugi.userTokens = make([]*hadoop_common.TokenProto, 0, maxTokens) //empty, with room for maxTokens tokens.
	}

	return ugi
}

func initializeCurrentUser() {
	once.Do(func() {
		currentUserGroupInformation = Allocate(nil, nil)
	})
}

func (ugi *UserGroupInformation) GetUserInformation() *hadoop_common.UserInformationProto {
	return ugi.userInfo
}

func (ugi *UserGroupInformation) GetUserTokens() []*hadoop_common.TokenProto {
	return ugi.userTokens
}

func (ugi *UserGroupInformation) AddUserToken(token *hadoop_common.TokenProto) {
	if length := len(ugi.userTokens); length < maxTokens {
		//ref: http://blog.golang.org/slices
		ugi.userTokens = append(ugi.userTokens, token)
	} else {
		log.Fatal("user already has maxTokens:", maxTokens)
	}
}

func GetCurrentUser() *UserGroupInformation {
	initializeCurrentUser()

	return currentUserGroupInformation
}
