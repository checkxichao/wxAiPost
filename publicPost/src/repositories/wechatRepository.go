package repositories

import (
	"errors"
	"gorm.io/gorm"
	"math/rand"
	"publicPost/src/models"
	"strings"
	"time"
)

type wechatRepository struct {
	db *gorm.DB
}
type WechatRepository interface {
	GetWechatInfo() ([]*models.WechatAccount, error)
	GetAllInfoByWxid(wxid string) ([]*models.WechatMediaModel, error)
	GetMediaListByWxid(wxid string, query string) ([]*models.WechatMediaModel, int, error)
	GetWxNowState(wxid string) (int, error)
	GetWxInfoByWxid(wxid string) (*models.WechatAccount, error)
	GetThumbMediaIdList(wxid string) ([]*models.ThumbMediaModel, error)
	GetTitleList(page int, pageSize int) ([]*models.TitleListModel, int64, error)
	GetTitleListSearch(page int, pageSize int, search string) ([]*models.TitleListModel, int64, error)
	GetRandomTitles(count int) ([]string, error)
	GetSequentialTitles(count int) ([]string, error)
	GetTitlesByWxid(wxid string) ([]*models.WechatMediaModel, error)
	GetWechatMediaByBelongWxid(mediaId string, startDate, endDate time.Time) []*models.WechatMediaModel

	GetTitleState(title string) *models.TitleListModel
	CreateWechat(wechat *models.WechatAccount) error

	UpdateWechat(id int, name string, wxid string, isuse bool, binder string, secret string) error
	SetStateByMediaId(mediaId string) error
	SetMediaId(wxid string, mediaId string, selfTitle string) error
	SetWxNowState(wxid string, state int) error
	SetMediaState(mediaId string, state int) error
	SetAccessToken(wxid string, accessToken string) error
	SetThumbMediaId(wxid string, mediaId string, note string, imgUrl string) error
	SetTitle(title string) error
	SetTitleState(title string, state bool) error
	EditTitle(id int, title string) error
	SetTitlesBatch(titles []string) error
	DeleteTitle(id int, title string) error
	DeleteTitleByTitle(title string) error
	DeleteTitleBatch(id []int) error
	DeleteMediaId(mediaId string) error
	DeleteThumbId(thumbId string) error
	DeleteWechat(wxid string) error
}

func NewWechatRepository(db *gorm.DB) WechatRepository {
	return &wechatRepository{db: db}
}
func (r *wechatRepository) GetAllInfoByWxid(wxid string) ([]*models.WechatMediaModel, error) {
	result := make([]*models.WechatMediaModel, 0)
	err := r.db.Table("wechat_media").Where("belong_wxid = ?", wxid).Where("state = ?", 0).Find(&result).Error
	return result, err
}
func (r *wechatRepository) GetMediaListByWxid(wxid string, query string) ([]*models.WechatMediaModel, int, error) {
	var mediaList []*models.WechatMediaModel
	var total int64

	db := r.db.Table("wechat_media").Where("belong_wxid = ? AND state = ?", wxid, 0)

	if query != "" {
		likeQuery := "%" + query + "%"
		db = db.Where("media_id LIKE ? OR title LIKE ?", likeQuery, likeQuery)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("created_at desc").Find(&mediaList).Error; err != nil {
		return nil, 0, err
	}

	return mediaList, int(total), nil
}
func (r *wechatRepository) SetStateByMediaId(mediaId string) error {
	return r.db.Table("wechat_media").Where("media_id = ?", mediaId).Update("state", 1).Error
}
func (r *wechatRepository) SetMediaId(wxid string, mediaId string, selfTitle string) error {
	medias := models.WechatMediaModel{
		Id:         0,
		MediaId:    mediaId,
		BelongWxid: wxid,
		State:      false,
		CreatedAt:  time.Time{},
		UpdatedAt:  time.Time{},
		Title:      selfTitle,
	}
	return r.db.Table("wechat_media").Create(&medias).Error
}
func (r *wechatRepository) GetWechatInfo() ([]*models.WechatAccount, error) {
	var info []*models.WechatAccount

	if err := r.db.Table("wechat_account").Select("*").Where("`show` = 1").Find(&info).Error; err != nil {
		return nil, err
	}

	return info, nil
}
func (r *wechatRepository) GetWxNowState(wxid string) (int, error) {
	wx := models.WechatAccount{}

	err := r.db.Table("wechat_account").Where("wxid", wxid).Find(&wx).Error
	if err != nil {
		return -1, err
	}
	return wx.NowState, err
}
func (r *wechatRepository) SetWxNowState(wxid string, state int) error {

	return r.db.Table("wechat_account").Where("wxid", wxid).Update("now_state", state).Error
}
func (r *wechatRepository) SetMediaState(mediaId string, state int) error {
	return r.db.Table("wechat_media").
		Where("media_id", mediaId).
		Updates(map[string]interface{}{
			"state":      state,
			"updated_at": time.Now(),
		}).Error
}

func (r *wechatRepository) GetTitlesByWxid(wxid string) ([]*models.WechatMediaModel, error) {
	result := make([]*models.WechatMediaModel, 0)
	err := r.db.Table("wechat_media").Where("belong_wxid", wxid).Find(&result).Error
	return result, err
}
func (r *wechatRepository) UpdateWechat(id int, name string, wxid string, isuse bool, binder string, secret string) error {

	updates := map[string]interface{}{
		"name":        name,
		"wxid":        wxid,
		"is_use":      isuse,
		"bind_wechat": binder,
		"secret":      secret,
	}

	return r.db.Table("wechat_account").Where("id", id).Updates(updates).Error
}
func (r *wechatRepository) CreateWechat(wechat *models.WechatAccount) error {
	wechat.GetTime = time.Now()
	wechat.EndTime = time.Now()

	return r.db.Table("wechat_account").Omit("end_time", "get_time").Create(wechat).Error
}

func (r *wechatRepository) DeleteWechat(wxid string) error {
	return r.db.Table("wechat_account").
		Where("wxid = ?", wxid).
		Delete(nil).
		Error
}

func (r *wechatRepository) SetAccessToken(wxid string, accessToken string) error {

	currentTime := time.Now()
	expirationTime := currentTime.Add(2 * time.Hour)

	return r.db.Table("wechat_account").Where("wxid = ?", wxid).Updates(map[string]interface{}{
		"access_token": accessToken,
		"get_time":     currentTime,
		"end_time":     expirationTime,
	}).Error
}

func (r *wechatRepository) GetWxInfoByWxid(wxid string) (*models.WechatAccount, error) {
	result := &models.WechatAccount{}
	err := r.db.Table("wechat_account").Where("wxid = ?", wxid).Find(&result).Error
	return result, err
}
func (r *wechatRepository) SetThumbMediaId(wxid string, mediaId string, note string, imgUrl string) error {
	insert := models.ThumbMediaModel{
		Id:           0,
		ThumbMediaId: mediaId,
		Wxid:         wxid,
		Note:         note,
		ImgUrl:       imgUrl,
		CreatedAt:    time.Time{},
	}
	return r.db.Table("thumb_media").Create(&insert).Error
}
func (r *wechatRepository) GetThumbMediaIdList(wxid string) ([]*models.ThumbMediaModel, error) {
	result := make([]*models.ThumbMediaModel, 0)
	err := r.db.Table("thumb_media").Where("wxid = ?", wxid).Find(&result).Error
	return result, err
}
func (r *wechatRepository) SetTitle(title string) error {
	model := models.TitleListModel{
		Id:        0,
		Title:     title,
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}
	return r.db.Table("title_list").Create(&model).Error
}
func (r *wechatRepository) EditTitle(id int, title string) error {

	return r.db.Table("title_list").Where("id", id).Update("title", title).Error
}
func (r *wechatRepository) DeleteTitle(id int, title string) error {
	return r.db.Table("title_list").Where("id", id).Where("title", title).Delete(nil).Error
}
func (r *wechatRepository) DeleteTitleByTitle(title string) error {
	return r.db.Table("title_list").Where("title", title).Delete(nil).Error
}
func (r *wechatRepository) DeleteTitleBatch(ids []int) error {
	if len(ids) == 0 {
		return nil // 没有 ID 需要删除，直接返回
	}

	result := r.db.Table("title_list").Where("id IN ?", ids).Delete(nil)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
func (r *wechatRepository) GetRandomTitles(count int) ([]string, error) {
	var titles []string

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	var total int64
	if err := r.db.Table("title_list").Count(&total).Error; err != nil {

		return nil, err
	}

	if total == 0 {
		return titles, nil
	}

	offsets := make(map[int]struct{})
	for len(offsets) < count && len(offsets) < int(total) {
		offset := rng.Intn(int(total))
		offsets[offset] = struct{}{}
	}

	for offset := range offsets {
		var title string
		if err := r.db.Table("title_list").
			Select("title").
			Order("id ASC").
			Where("state = ?", 0).
			Offset(offset).
			Limit(1).
			Pluck("title", &title).Error; err != nil {

			continue
		}
		titles = append(titles, title)
	}

	return titles, nil
}
func (r *wechatRepository) GetSequentialTitles(count int) ([]string, error) {
	var titles []string
	err := r.db.Table("title_list").
		Select("title").
		Order("id ASC").
		Where("state = ?", 0).
		Limit(count).
		Pluck("title", &titles).Error
	if err != nil {

		return nil, err
	}
	return titles, nil
}
func (r *wechatRepository) GetTitleList(page int, pageSize int) ([]*models.TitleListModel, int64, error) {
	var result []*models.TitleListModel
	var total int64
	if page == 0 && pageSize == 0 {
		errs := r.db.Table("title_list").Find(&result).Error
		if errs != nil {
			return nil, 0, errs
		}
		return result, 0, nil
	}

	offset := (page - 1) * pageSize

	if err := r.db.Table("title_list").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.Table("title_list").Limit(pageSize).Offset(offset).Find(&result).Where("state = 0").Error; err != nil {
		return nil, 0, err
	}

	return result, total, nil
}
func (r *wechatRepository) GetTitleListSearch(page int, pageSize int, search string) ([]*models.TitleListModel, int64, error) {
	var result []*models.TitleListModel
	var total int64

	query := r.db.Table("title_list")

	if search != "" {
		likeQuery := "%" + search + "%"
		query = query.Where("title LIKE ?", likeQuery)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page == 0 && pageSize == 0 {
		if err := query.Find(&result).Error; err != nil {
			return nil, 0, err
		}
		return result, total, nil
	}

	offset := (page - 1) * pageSize

	if err := query.Limit(pageSize).Offset(offset).Find(&result).Where("state = 0").Error; err != nil {
		return nil, 0, err
	}

	return result, total, nil
}

func (r *wechatRepository) SetTitlesBatch(titles []string) error {
	if len(titles) == 0 {
		return errors.New("no titles ")
	}

	var titleModels []models.TitleListModel
	for _, title := range titles {
		trimmedTitle := strings.TrimSpace(title)
		if trimmedTitle == "" {
			return errors.New("titles1")
		}
		titleModels = append(titleModels, models.TitleListModel{
			Title: trimmedTitle,
		})
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("title_list").Create(&titleModels).Error; err != nil {
			return err
		}
		return nil
	})
}
func (r *wechatRepository) DeleteMediaId(mediaId string) error {
	return r.db.Table("wechat_media").Where("media_id", mediaId).Delete(nil).Error
}
func (r *wechatRepository) DeleteThumbId(thumbId string) error {
	return r.db.Table("thumb_media").Where("thumb_media_id", thumbId).Delete(nil).Error
}
func (r *wechatRepository) GetWechatMediaByBelongWxid(belongWxid string, startDate, endDate time.Time) []*models.WechatMediaModel {
	var result []*models.WechatMediaModel
	query := r.db.Table("wechat_media").Where("state = ?", true)

	if belongWxid != "" {
		query = query.Where("belong_wxid = ?", belongWxid)
	}
	if !startDate.IsZero() && !endDate.IsZero() {
		query = query.Where("updated_at BETWEEN ? AND ?", startDate, endDate)
	}

	query.Find(&result)
	return result
}

func (r *wechatRepository) SetTitleState(title string, state bool) error {
	return r.db.Table("title_list").
		Where("title = ?", title).
		Updates(map[string]interface{}{
			"state":      state,
			"updated_at": time.Now(),
		}).Error
}
func (r *wechatRepository) GetTitleState(title string) *models.TitleListModel {
	var result models.TitleListModel

	err := r.db.Table("title_list").Where("title = ?", title).Take(&result).Error
	if err != nil {
		return nil
	}
	return &result
}
