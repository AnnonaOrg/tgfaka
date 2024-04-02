package services

//func RegisterUser(userName string) *models.User {
//	var user *models.User
//	db.DB.Create(&user)
//	return user
//}

func LoginUserWithoutPassword(userName string) {

}

//func GetUserByUsername(user *models.User, username string) error {
//	result := db.DB.Where("username=?", username).Find(&user)
//	if result.Error != nil {
//		//return errors.New("db_err")
//	}
//	if result.RowsAffected == 0 {
//		return errors.New("not_found")
//	}
//	return nil
//}
//
//func CreateUser(user *models.User) error {
//	result := db.DB.Create(&user)
//	if result.Error != nil {
//		return errors.New("db_err")
//	}
//	return nil
//}

//func UpdateUserLogin(user *models.User, loginIP string) error {
//	result := db.DB.Model(&models.User{}).Where("id=?", user.ID).Updates(map[string]interface{}{
//		"login_time": time.Now().Unix(),
//		"login_ip":   loginIP,
//		//"meta.token": token,
//	})
//	if result.Error != nil {
//		return errors.New("db_err")
//	}
//	return nil
//}
