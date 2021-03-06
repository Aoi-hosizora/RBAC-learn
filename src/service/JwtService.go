package service

import (
	"github.com/Aoi-hosizora/RBAC-learn/src/common/exception"
	"github.com/Aoi-hosizora/RBAC-learn/src/common/result"
	"github.com/Aoi-hosizora/RBAC-learn/src/config"
	"github.com/Aoi-hosizora/RBAC-learn/src/model/po"
	"github.com/Aoi-hosizora/RBAC-learn/src/util"
	"github.com/Aoi-hosizora/ahlib/xdi"
	"github.com/gin-gonic/gin"
	"log"
)

type JwtService struct {
	Config    *config.ServerConfig `di:"~"`
	UserRepo  *UserService         `di:"~"`
	TokenRepo *TokenService        `di:"~"`

	UserKey string `di:"-"`
}

func NewJwtService(dic *xdi.DiContainer) *JwtService {
	srv := &JwtService{UserKey: "user"}
	dic.MustInject(srv)
	return srv
}

func (j *JwtService) GetToken(c *gin.Context) string {
	if token := c.Request.Header.Get("Authorization"); token != "" {
		return token
	} else {
		return c.DefaultQuery("Authorization", "")
	}
}

func (j *JwtService) JwtCheck(token string) (*po.User, *exception.Error) {
	if token == "" {
		return nil, exception.UnAuthorizedError
	}

	// fake
	if j.Config.MetaConfig.RunMode == "debug" {
		if uid, ok := j.Config.JwtConfig.FakeToken[token]; ok {
			log.Println(uid)
			user, ok := j.UserRepo.QueryById(uid)
			if !ok {
				return nil, exception.UnAuthorizedError
			}
			return user, nil
		}
	}

	// parse
	claims, err := util.AuthUtil.ParseToken(token, j.Config.JwtConfig)
	if err != nil {
		if util.AuthUtil.IsTokenExpired(err) {
			return nil, exception.TokenExpiredError
		} else {
			return nil, exception.UnAuthorizedError
		}
	}

	// redis
	ok := j.TokenRepo.Query(token)
	if !ok {
		return nil, exception.UnAuthorizedError
	}

	// mysql
	user, ok := j.UserRepo.QueryById(claims.UserId)
	if !ok {
		return nil, exception.UnAuthorizedError
	}

	return user, nil
}

func (j *JwtService) GetContextUser(c *gin.Context) *po.User {
	_user, exist := c.Get(j.UserKey)
	if exist { // has jwtMw
		user, ok := _user.(*po.User)
		if !ok {
			result.Error(exception.UnAuthorizedError).JSON(c)
			c.Abort() // abort
			return nil
		}
		return user
	} else { // no jwtMw
		token := j.GetToken(c)
		user, err := j.JwtCheck(token)
		if err != nil {
			return nil // auth failed
		} else {
			c.Set(j.UserKey, user)
			return user
		}
	}
}
