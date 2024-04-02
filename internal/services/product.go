package services

import (
	"errors"
	"gopay/internal/exts/db"
	"gopay/internal/models"

	"github.com/google/uuid"
)

type ProductService struct {
}

func GetProductsByCustomer(pagination *Pagination) error {
	query := db.DB.Where("status = 1")
	query = query.Order((&models.Product{}).DefaultOrder())

	err := Paginate[models.Product](pagination, query)
	if err != nil {
		return err
	}
	return nil
}
func GetProductByIDByCustomer(productID uuid.UUID) (models.Product, error) {
	var product models.Product
	if err := db.DB.Where("id=? and status = 1", productID).First(&product).Error; err != nil {
		return product, err
	}
	return product, nil
}

func CreateProduct(product *models.Product) error {
	result := db.DB.Create(&product)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func UpdateProduct(productID uuid.UUID, updateMap map[string]interface{}) error {
	result := db.DB.Model(&models.Product{}).Where("id=?", productID).Updates(updateMap)
	if result.RowsAffected == 0 {
		return errors.New("not_found")
	}
	return nil
}

//
//func DeleteProducts(productIDs []uuid.UUID) error {
//	if err := db.DB.Where("id in ?", productIDs).Delete(&models.Product{}).Error; err != nil {
//		return err
//	}
//
//	//tx := db.DB.Begin()
//	//defer tx.Rollback()
//	//
//	//// Delete related ProductItems first
//	//if err := tx.Where("product_id in ?", productIDs).Delete(&models.ProductItem{}).Error; err != nil {
//	//	return err
//	//}
//	//
//	////if err := tx.Model(models.Order{}).Where("product_id in ?", productIDs).Updates(map[string]interface{}{
//	////	"product_id": gorm.Expr("NULL"),
//	////}).Error; err != nil {
//	////	return err
//	////}
//	//
//	//// Then delete the Product
//	//if err := tx.Where("id in ?", productIDs).Delete(&models.Product{}).Error; err != nil {
//	//	return err
//	//}
//	//
//	//// Commit the transaction
//	//err := tx.Commit().Error
//	//if err != nil {
//	//	return err
//	//}
//
//	return nil
//}

func UpdateProductInStockCount(productIDs []uuid.UUID) error {
	//// 查询数量
	//var inStockCount int64
	//if err := db.DB.Model(&models.ProductItem{}).Where("product_id=? and status=0", productID).Count(&inStockCount).Error; err != nil {
	//	return err
	//}
	//
	////更新数量
	//if err := db.DB.Model(&models.Product{}).Where("id=?", productID).Update("in_stock_count", inStockCount).Error; err != nil {
	//	return err
	//}

	result := db.DB.Exec(`
			UPDATE product 
			SET in_stock_count = (
				SELECT COUNT(*)
				FROM product_item
				WHERE product_item.product_id = product.id AND product_item.status = 1
			)
			WHERE id IN ?;
		`, productIDs)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
