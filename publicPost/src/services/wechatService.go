package services

import (
	"publicPost/src/models"
	"publicPost/src/repositories"
	"time"
)

type WechatService struct {
	wechatRepo repositories.WechatRepository
}

func NewWechatService(wechatRepo repositories.WechatRepository) *WechatService {
	return &WechatService{
		wechatRepo: wechatRepo,
	}
}

func (ws *WechatService) AddAccount(name string, wxid string, secret string, bindWechat string) (*models.WechatAccount, error) {

	wechat := &models.WechatAccount{
		ID:         0,
		Name:       name,
		Wxid:       wxid,
		Secret:     secret,
		IsUse:      true,
		BindWechat: bindWechat,
		Show:       true,
		NowState:   0,
	}
	err := ws.wechatRepo.CreateWechat(wechat)
	if err != nil {
		return nil, err
	}
	return wechat, nil
}

func (ws *WechatService) DeleteAccount(wxid string) error {

	err := ws.wechatRepo.DeleteWechat(wxid)
	if err != nil {
		return err
	}
	return nil
}

func (ws *WechatService) UpdateAccount(account models.WechatAccount) error {

	err := ws.wechatRepo.UpdateWechat(int(account.ID), account.Name, account.Wxid, account.IsUse, account.BindWechat, account.Secret)
	if err != nil {
		return err
	}
	return nil
}

func (ws *WechatService) GetWechatInfo() ([]*models.WechatAccount, error) {

	var wechatAccounts []*models.WechatAccount
	wechatAccounts, err := ws.wechatRepo.GetWechatInfo()
	if err != nil {
		return make([]*models.WechatAccount, 0), err
	}

	return wechatAccounts, nil
}

func (ws *WechatService) GetWechatMediaByPost(belongWxid string, startDate, endDate time.Time) []*models.WechatMediaModel {
	return ws.wechatRepo.GetWechatMediaByBelongWxid(belongWxid, startDate, endDate)
}
